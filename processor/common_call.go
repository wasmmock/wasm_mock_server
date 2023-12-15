package processor

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/wasmmock/wasm_mock_server/model"
	"github.com/wasmmock/wasm_mock_server/util"
	"github.com/wasmmock/wasm_mock_server/util/postman"
	"github.com/google/uuid"
	wapc "github.com/wapc/wapc-go"
	"github.com/wapc/wapc-go/engines/wazero"
)

type genericWriter interface {
	loop() string
	rw() http.ResponseWriter
	response() model.CallApiResponse
	mock_targets() string
	wasmLang() string
	jsonResponser(uID string, er error, errcode uint32)
	wasmCode() ([]byte, error)
	requester(string, interface{}, interface{}) interface{}
	curlCommand() string
	postman() (postman.Item, error)
	saveHandleInitErr(message string, err error, uID string) error
	saveHandleLoopErr(message string, err error, uID string) error
}

func CallWasm(gw genericWriter, ws_clientid string) {
	uID := uuid.New().String()
	safeReports.ReportGen(uID)
	defer safeReports.Save(uID)
	code, err := gw.wasmCode()
	wasmLang := gw.wasmLang()
	loop := gw.loop()
	lower := int64(0)
	upper := int64(0)
	lower, upper, err = BoundFromLoop(loop, err)
	if err := gw.saveHandleInitErr("get_uid ", err, uID); err != nil {
		return
	}
	mockTargets := gw.mock_targets()
	mockTargetsList := strings.Split(mockTargets, ",")
	for _, target := range mockTargetsList {
		MockCommandUidMap.Store(target, uID)
	}
	ctx := context.Background()
	engine := wazero.Engine()
	module, err := engine.New(ctx, hostCall, code, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	//module, err := wapc.New(code, hostCall)
	if err := gw.saveHandleInitErr("wapc hostcall init ", err, uID); err != nil {
		return
	}
	// module.SetLogger(wapc.Println) // Send __console_log calls to stardard out
	// module.SetWriter(wapc.Print)   // Send WASI fd_write calls to stardard out
	defer module.Close(ctx)
	var instance wapc.Instance

	if wasmLang == "go" {
		//instance, err = module.Instantiate("wasi_snapshot_preview1")
		instance, err = module.Instantiate(ctx)
	} else {
		//instance, err = module.Instantiate("wasi_unstable")
		instance, err = module.Instantiate(ctx)
	}
	s_instance := util.NewSafeInstance(instance)
	if err := gw.saveHandleInitErr("wapc module instantiate ", err, uID); err != nil {
		return
	}
	defer s_instance.Close()
	s_instance.Invoke(ctx, "save_ws_uid", []byte(ws_clientid))
	_, err = s_instance.Invoke(ctx, "save_uid", []byte(uID))
	if err := gw.saveHandleInitErr("save_uid ", err, uID); err != nil {
		fmt.Println("call save_uid", err)
		return
	}
	wasmUID, err2 := s_instance.Invoke(ctx, "get_uid", []byte(""))
	if err := gw.saveHandleInitErr("get_uid ", err2, uID); err != nil {
		fmt.Println("call get_uid", err)
		return
	}
	s_instance.Invoke(ctx, "add_functions", []byte{})
	_ = wasmUID
	defer resetState(mockTargetsList, uID)

	var errCode uint32 = 1

	for i := int64(0); i < lower; i++ {
		safeReports.AppendUnitTest(uID, i)
	}
	for i := lower; i < upper; i++ {
		tpexRes := Bytes{}
		indexMap[uID] = i
		safeReports.AppendUnitTest(uID, i)
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(i))
		command, err := s_instance.Invoke(ctx, "command", b)
		if err := gw.saveHandleLoopErr("command ", err, uID); err != nil {
			continue
		}
		command_str := string(command)
		iStepCommand := util.Step{Pass: true, Description: "command: " + command_str} //to be added rpc
		safeReports.AppendStep(uID, iStepCommand)
		for _, target := range mockTargetsList {
			indexMap[target] = i
		}
		tpexReq, err := s_instance.Invoke(ctx, "request", Bytes{})
		if err := gw.saveHandleLoopErr("request ", err, uID); err != nil {
			continue
		}
		tpexReq2 := Bytes(tpexReq)
		errCodef := gw.requester(command_str, &tpexReq2, &tpexRes)
		var error_str = ""
		switch v := errCodef.(type) {
		case int:
			errCode = uint32(v)
		case uint32:
			errCode = v
		case error:
			errCode = 1
			error_str = v.Error()
			fmt.Printf("http err %T", v.Error())
		default:
			fmt.Printf("I don't know about type %T!\n", v)
		}
		safeReports.AppendEnd(uID, backgroundMock, hits.Clone())
		tpexReq2, err = s_instance.Invoke(ctx, "request_marshalling", tpexReq2)
		if err := gw.saveHandleLoopErr("request_marshalling ", err, uID); err != nil {
			continue
		}
		safeReports.AppendRequest(uID, string(tpexReq2))
		//	if errCode == 0 {
		if len(tpexRes) > 0 {
			tpexRes, err = s_instance.Invoke(ctx, "response_marshalling", tpexRes)
			if err := gw.saveHandleLoopErr("response_marshalling ", err, uID); err != nil {
			}
		}
		if error_str != "" {
			safeReports.AppendResponse(uID, string(tpexRes)+"<br>"+"errcode: "+error_str)
		} else {
			safeReports.AppendResponse(uID, string(tpexRes)+"<br>"+"errcode: "+strconv.Itoa(int(errCode)))
		}
		// } else {
		// 	safeReports.AppendResponse(uID, "errcode: "+strconv.Itoa(int(errCode)))
		// }
	}

	gw.jsonResponser(uID, err, errCode)
	return
}
