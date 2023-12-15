package capabilities

import (
	"github.com/vmihailenco/msgpack"
)

type TcpItem struct {
	Payload string
	String  string
	Id      string
	Laddr   string
	Raddr   string
}
type TcpReq struct {
	Payload    string
	String     string
	Index      int
	Id         string
	Command    string
	ReportType string
	Timeout    bool
	Laddr      string
	Raddr      string
}
type TcpFiddlerAB struct {
	ResA []byte
	ResB []byte
}

func TcpNewReq(payload []byte, binding string, tcpNewWriteReqChan map[string]chan TcpReq, tcpNewWriteReqResponderChan map[string]chan TcpReq) (TcpReq, error) {
	if v, ok := tcpNewWriteReqChan[binding]; ok {
		//do something here
		var item TcpReq
		err := msgpack.Unmarshal(payload, &item)
		if err == nil {
			//sDec, _ := b64.StdEncoding.DecodeString(item.Payload)
			go func() {
				v <- item
			}()

			for {
				responder := <-tcpNewWriteReqResponderChan[item.Id]
				delete(tcpNewWriteReqResponderChan, item.Id)
				return responder, nil
			}
		} else {
			return TcpReq{}, err
		}
	}
	return TcpReq{}, nil
}
func TcpNewRes(payload []byte, binding string, tcpNewWriteResChan map[string]chan TcpReq) (string, string, error) {
	if v, ok := tcpNewWriteResChan[binding]; ok {
		//do something here
		var items []TcpItem
		err := msgpack.Unmarshal(payload, &items)
		for _, item := range items {
			if item.Payload != "/continue" {
				if err == nil {
					//sDec, _ := b64.StdEncoding.DecodeString(item.Payload)
					go func() {
						v <- TcpReq{Payload: item.Payload, Laddr: item.Laddr, Raddr: item.Raddr}
					}()

					return item.String, item.Id, nil
				} else {
					return "", "", err
				}
			}

		}

	}
	return "", "", nil
}
func TcpNewReqDial(payload []byte, binding string) {

}
