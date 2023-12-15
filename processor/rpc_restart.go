package processor

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
)

type RpcAble interface {
	Start([]string)
	Stop()
	HostCallSpecific(string, string, string, []byte) ([]byte, error)
	RegisterCommand(string, func(ctx context.Context, request, response interface{}) uint32)
	RpcRequest(string, proto.Message, proto.Message) uint32
	Source(context.Context, interface{}) string
	TraceId(context.Context, interface{}) string
	FiddleQueue(chan FiddleAB, chan FiddleAB)
	TcpFiddleQueue(chan TcpFiddleAB, chan TcpFiddleAB)
	UserPermissionJWT(string, string) (string, error)
}

func RegisterCommands(rpcHandler RpcAble) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			fmt.Fprintf(rw, "ParseForm() err: %v", err)
			return
		}
		commands := req.FormValue("commands")
		commandsList := strings.Split(commands, ",")
		rpcHandler.Stop()
		rpcHandler.Start(commandsList)
		registeredCommands = commandsList
	}
}
