package util

import (
	"database/sql"
	"fmt"

	"github.com/wasmmock/wasm_mock_server/myerror"
)

type DB_Op struct {
}

func (d *DB_Op) SetError(err error) {

}
func (d *DB_Op) GetError() error {
	return nil
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
