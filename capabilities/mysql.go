package capabilities

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/wasmmock/wasm_mock_server/myerror"
)

type MysqlRequest struct {
	Query       string `json:"query"`
	Execute     string `json:"execute"`
	QueryColumn string `json:query_column`
}
type MysqlResponse struct {
	data [][][]byte `json:"data"`
}

func MysqlStatement(addr string, req []byte) ([]byte, error) {

	wasmReq := MysqlRequest{}
	err1 := json.Unmarshal(req, &wasmReq)

	if wasmReq.Query != "" {

	}
	if wasmReq.Execute != "" {

	}
	return []byte{}, err1
}

func Column3(rows *sql.Rows, row int) [][]byte {
	var a = [][]byte{[]byte{}, []byte{}, []byte{}}
	rows.Scan(&a[0], &a[1], &a[2])
	return a
}
func Column2(rows *sql.Rows, row int) [][]byte {
	var a = [][]byte{[]byte{}, []byte{}}
	rows.Scan(&a[0], &a[1])
	return a
}

//handleGet used for Database get row
func HandleGet(rows *sql.Rows, err error, closure func(*sql.Rows, int) [][]byte) ([][][]byte, myerror.Code) {
	var rowLen = 0
	var b = [][][]byte{}
	if err != nil {
		fmt.Println("db..", err.Error())
		return b, myerror.DBError
	}

	for rows.Next() {
		r := closure(rows, rowLen)
		b = append(b, r)
		rowLen++
	}
	if rowLen == 0 {
		return b, myerror.NoRecords
	}
	return b, myerror.Nil
}
func HandleUpdate(res sql.Result, err error) myerror.Code {
	if err != nil {
		fmt.Println("db..", err.Error())
		return myerror.DeleteRecordFail
	}
	aRow, _ := res.RowsAffected()
	if aRow == 0 {
		return myerror.DeleteRecordFail
	}
	return myerror.Nil
}
