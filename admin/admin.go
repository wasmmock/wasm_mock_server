package Admin

import (
	"net"
	"net/http"
	"strings"

	"github.com/wasmmock/wasm_mock_server/processor"
	Security "github.com/wasmmock/wasm_mock_server/security"
	"github.com/wasmmock/wasm_mock_server/tcpproxy"
	"github.com/wasmmock/wasm_mock_server/util"
)

type statusInfo struct {
	ConnectionType string `json:"connection_type"`
	Laddr          string `json:"laddr"`
	LaddrStatus    int    `json:"laddr_status"`
	LaddrRemote    string `json:"laddr_remote"`
	Raddr          string `json:"raddr"`
	RaddrStatus    int    `json:"raddr_status"`
	RaddrRemote    string `json:"raddr_remote"`
	Command        string `json:"command"`
	X_Api_Key      string `json:"x_api_key"`
}

func GetStatus() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var data = make([]statusInfo, 0)
		processor.MockCommandMockUidMap.Range(func(key string, value string) bool {
			d := statusInfo{
				Command: key,
			}
			tcpproxy.TcpLconnSafe.Range(func(laddr interface{}, lconn interface{}) bool {
				if la, ok := laddr.(string); ok {
					if strings.Contains(la, key) {
						d.Laddr = la

						d.ConnectionType = "tcp_mock"
						if lconn, ok := lconn.(net.Conn); ok {
							d.LaddrRemote += "|" + lconn.RemoteAddr().String()
							d.LaddrStatus = 1
						}
					}
				}
				return true
			})
			tcpproxy.TcpRconnSafe.Range(func(raddr interface{}, rconn interface{}) bool {
				if ra, ok := raddr.(string); ok {
					// if strings.Contains(ra, key) {
					// 	d.Raddr = ra
					if _, ok := rconn.(net.Conn); ok {
						d.RaddrRemote += "|" + ra
						d.RaddrStatus = 1
					}
					//}
				}
				return true
			})
			Security.X_Api_Key_Map.Range(func(key interface{}, inner interface{}) bool {
				if uid, ok := key.(string); ok {
					if uid == value {
						if x_api_key, ok := inner.(string); ok {
							d.X_Api_Key = x_api_key
						}
						return false
					}
				}
				return true
			})
			data = append(data, d)
			return true
		})
		processor.MockCommandUidMap.Range(func(key string, value string) bool {
			d := statusInfo{
				Command:        key,
				ConnectionType: "call",
			}
			data = append(data, d)
			return true
		})

		util.JsonResponse(rw, data)
	}

}
