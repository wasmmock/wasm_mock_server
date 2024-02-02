package tcpproxy

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	b64 "encoding/base64"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/wapc/wapc-go"
	"github.com/wasmmock/wasm_mock_server/capabilities"
	"github.com/wasmmock/wasm_mock_server/util"
)

var TcpLconnSafe sync.Map
var TcpRconnSafe sync.Map

var TcpReqResponderChan map[string]chan capabilities.TcpReq = make(map[string]chan capabilities.TcpReq)
var removeQueue = make(chan capabilities.EntityWhole)
var idMap safeIdMapStruct
var replayIdMap safeIdMapStruct
var TcpListenerMap map[string]net.Listener = make(map[string]net.Listener)

func TcpRequest(requests []capabilities.TcpReq, portmap string) []capabilities.TcpReq {
	var responderArr = []capabilities.TcpReq{}
	for _, req := range requests {
		responder := capabilities.TcpReq{}
		if r, ok := TcpRconnSafe.Load(portmap + "_" + req.Laddr); ok {
			if rconn, ok := r.(net.Conn); ok {
				payload, _ := b64.StdEncoding.DecodeString(req.Payload)
				n, er := rconn.Write(payload)
				log.Println("TcpRconnSafe, rconn write", n, er, "len payload", len(payload))
				if req.Id != "" { //async
					c := make(chan capabilities.TcpReq)
					TcpReqResponderChan[req.Id] = c
					t := time.Now()
					replayIdMap.append(capabilities.EntityWhole{Id: req.Id, Time: t, Command: req.Command, ReportType: req.ReportType})
					//			idMap.append(capabilities.EntityWhole{Id: req.Id, Time: t, Command: req.Command})
					for {
						responder = <-TcpReqResponderChan[req.Id]
						delete(TcpReqResponderChan, req.Id)
					}
				}
			}

		}
		responderArr = append(responderArr, responder)
	}

	return responderArr
}
func TcpResponse(requests []capabilities.TcpReq, portmap string) {
	for _, req := range requests {
		payload, _ := b64.StdEncoding.DecodeString(req.Payload)
		TcpLconnSafe.Range(func(key interface{}, lconn interface{}) bool {
			if k, ok := key.(string); ok {
				if strings.Contains(k, portmap) {
					if lconn, ok := lconn.(net.Conn); ok {
						if req.Laddr == "" {
							lconn.Write(payload)
						} else if lconn.RemoteAddr().String() == req.Laddr {
							lconn.Write(payload)
							return false
						}
					}
				}
			}
			return true
		})
	}
}
func TcpMockTeardown(targetlist []string, MockCommandMockUidMap *util.SafeStringMap) {
	log.Println("TcpMockTeardown", targetlist)
	for _, portmap := range targetlist {
		TcpRconnSafe.Range(func(key interface{}, rconn interface{}) bool {
			if k, ok := key.(string); ok {
				if strings.Contains(k, portmap) {
					if rconn, ok := rconn.(net.Conn); ok {
						log.Println("close remote addr", rconn.RemoteAddr().String())
						rconn.Close()
					}
					TcpRconnSafe.Delete(key)
					log.Println("TcpRconnSafe , Delete", key)
				}
			}
			return true
		})
		TcpLconnSafe.Range(func(key interface{}, lconn interface{}) bool {
			if k, ok := key.(string); ok {
				if strings.Contains(k, portmap) {
					if lconn, ok := lconn.(net.Conn); ok {
						lconn.Close()
					}
					TcpLconnSafe.Delete(key)
				}
			}
			return true
		})

		if v, ok := TcpListenerMap[portmap]; ok {
			v.Close()
			delete(TcpListenerMap, portmap)
		}
		MockCommandMockUidMap.Delete(portmap)
	}
}

func PoolInit(uID string, portmap string, laddr, raddr *net.TCPAddr, Before_req, Before_res func([]byte, string, string, string, string) ([]byte, []capabilities.Entity), MockCommandMockUidMap *util.SafeStringMap, mockUidInstanceMap *util.SafeInstanceMap, wasmModule wapc.Module) {
	//listener, nerr := net.ListenTCP("tcp", laddr)
	// cert, err := tls.LoadX509KeyPair("cert_folder/proxy-ca.pem", "cert_folder/proxy-ca.key")
	// if err != nil {
	// 	log.Fatalf("server: loadkeys: %s", err)
	// }
	//config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	//processor.MockCommandUidMap.Store(portmap, uID)

	//instance, err := wasmModule.Instantiate("wasi_unstable")
	ctx := context.Background()
	instance, err := wasmModule.Instantiate(ctx)
	if err != nil {
		log.Println("wasmModule instantiate ", err.Error(), time.Now())
	}
	safe_instance := util.NewSafeInstance(instance)

	log.Println("after Instantiate")
	var is_tls = false
	is_tls_byte, ner := safe_instance.Invoke(ctx, "is_tls", []byte{})
	if ner == nil {
		is_tls_str := string(is_tls_byte)
		if is_tls_str == "true" {
			is_tls = true
		}
	}
	var listener net.Listener
	var nerr error
	if is_tls {
		if cert, err := tls.LoadX509KeyPair("cert_folder/proxy-ca.pem", "cert_folder/proxy-ca.key"); err == nil {
			config := tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true}
			config.Rand = rand.Reader
			if listener, nerr = net.Listen("tcp", laddr.String()); nerr == nil {
				listener = tls.NewListener(listener, &config)
			} else {
				listener, nerr = net.Listen("tcp", laddr.String())
			}
		} else {
			listener, nerr = net.Listen("tcp", laddr.String())
		}
	} else {
		listener, nerr = net.Listen("tcp", laddr.String())
	}
	if nerr == nil {
		TcpListenerMap[portmap] = listener
	} else {
		log.Println("before Instantiate listener err", laddr.String(), nerr.Error())
	}
	MockCommandMockUidMap.Store(portmap, uID)
	l_remote_add := laddr.String()
	safe_instance.Invoke(ctx, "save_ws_uid", []byte(uID))
	// _, err = safe_instance.Invoke(ctx, "add_functions", []byte(portmap))

	mockUidInstanceMap.Set(l_remote_add, &safe_instance)
	//go ConnectR(portmap, raddr, l_remote_add, actual_l_remote_add, Before_res, MockCommandMockUidMap)
	//newly added
	TcpLconnSafe.Store(portmap+"_"+l_remote_add, nil)
	if nerr == nil {
		go func() {

			for {
				lconn, err := listener.Accept()
				if err != nil {
					log.Println("listener.AcceptTCP err", err)
					return
				}
				_ = err
				//l_remote_add := lconn.RemoteAddr().String()
				log.Println("lconn RemoteAddr", lconn.RemoteAddr().String(), "listener..", listener.Addr())
				go ConnectR(portmap, raddr, l_remote_add, lconn.RemoteAddr().String(), Before_res, MockCommandMockUidMap)
				TcpLconnSafe.Store(portmap+"_"+l_remote_add+"_"+lconn.RemoteAddr().String(), lconn)

				go func() {
					time.Sleep(2 * time.Second)
					for {
						buff := make([]byte, 20000)
						n, err := lconn.Read(buff)
						if err != nil {
							// TcpLconnSafe.Delete(portmap)
							// delete(TcpLconns[portmap])
							log.Println("lconn.Read", err, lconn.RemoteAddr().String())
							// // if v, ok := mockUidInstanceMap.Get(l_remote_add); ok {
							// // 	v.Close()
							// // }
							// mockUidInstanceMap.Delete(l_remote_add)
							return
						}
						b := buff[:n]
						mb, es := Before_req(b, portmap, l_remote_add, lconn.RemoteAddr().String(), "")
						_ = es
						for _, e := range es {
							if e.Id != "" {
								t := time.Now()
								idMap.append(capabilities.EntityWhole{Id: e.Id, Time: t, Command: e.Command})
							}
						}

						if !bytes.Equal(mb, []byte("/continue")) {
							if rconn, ok := TcpRconnSafe.Load(portmap + "_" + lconn.RemoteAddr().String()); ok {
								if rconn, ok := rconn.(net.Conn); ok {
									b := bytes.NewBuffer(mb)
									n, er := b.WriteTo(rconn)
									if er != nil {
										log.Println("rconn er", er.Error())
									}
									log.Println("rconn write n ", n, "mb len ", len(mb), rconn.RemoteAddr())
								}
							} else {
								log.Println("cannot find rconn", portmap+"_"+lconn.RemoteAddr().String())
							}
						}

					}
				}()
			}
			log.Println("listener close at ", laddr.String())
			listener.Close()
		}()
	}

}

func ConnectR(portmap string, raddr *net.TCPAddr, l_remote_add string, actual_l_remote_add string, Before_res func([]byte, string, string, string, string) ([]byte, []capabilities.Entity), MockCommandMockUidMap *util.SafeStringMap) {
	var err error
	if _, ok := MockCommandMockUidMap.Get(portmap); !ok {
		return
	}
	log.Println("reconnect from l_remote_add", l_remote_add)
	rconn, err := net.DialTCP("tcp", nil, raddr)
	if err == nil {
		TcpRconnSafe.Store(portmap+"_"+actual_l_remote_add, rconn)
		//TcpRconnSafe.Store(portmap+"_"+actual_l_remote_add, rconn)
		for {
			buff := make([]byte, 20000)
			n, err := rconn.Read(buff)
			if err != nil {
				//p.err("Read failed '%s'\n", err)
				if err != io.EOF {
					//	p.Log.Warn(s, err)

				}
				log.Println("Connect  R ", err)
				// TcpRconnSafe.Delete(portmap)
				// rconn.Close()
				break
				//return
			}
			b := buff[:n]
			actual_r_local_add := rconn.LocalAddr().String()
			mb, es := Before_res(b, portmap, l_remote_add, actual_l_remote_add, actual_r_local_add)
			_ = es
			if !bytes.Equal(mb, []byte("/continue")) {
				TcpLconnSafe.Range(func(key interface{}, lconn interface{}) bool {
					if k, ok := key.(string); ok {
						if strings.Contains(k, portmap) {
							if lconn, ok := lconn.(net.Conn); ok {
								log.Println("lconn.RemoteAddr().String()", lconn.RemoteAddr().String(), "actual_l_remote_add", actual_l_remote_add)
								if actual_l_remote_add == "" {
									b := bytes.NewBuffer(mb)
									n, er := b.WriteTo(lconn)
									if er != nil {
										log.Println("lconn write ", er.Error(), "n :", n, actual_l_remote_add)
									} else {
										log.Println("oklconn write ", portmap, actual_l_remote_add)
									}
								} else if lconn.RemoteAddr().String() == actual_l_remote_add {

									b := bytes.NewBuffer(mb)
									n, er := b.WriteTo(lconn)
									if er != nil {
										log.Println("lconn write equal", er.Error(), "n :", n, actual_l_remote_add)
									} else {
										log.Println("oklconn write equal", portmap, actual_l_remote_add)
									}
									return false
								}

							}
						}

					}
					return true
				})
				t := time.Now()
				for _, e := range es {
					if e.Id != "" {
						e2 := capabilities.EntityWhole{Id: e.Id, Time: t, Command: e.Command, Payload: b}
						removeQueue <- e2
					}
				}

			} else {
				log.Println("there is continue")
			}

		}
		time.Sleep(4000 * time.Millisecond)
		ConnectR(portmap, raddr, l_remote_add, actual_l_remote_add, Before_res, MockCommandMockUidMap)
	} else {
		log.Println("err reconnect", err)
		time.Sleep(2000 * time.Millisecond)
		ConnectR(portmap, raddr, l_remote_add, actual_l_remote_add, Before_res, MockCommandMockUidMap)
	}

}
func SelfRemove() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			idMap.SelfRemove()
			replayIdMap.SelfRemove()
		}
	}()
}

func HandleRemove2() {
	go func() {
		log.Println("handleremove")
		for {
			v := <-removeQueue
			not_the_same := idMap.remove2(v)
			if not_the_same {
				replayIdMap.remove2(v)
			}
		}
	}()
}
