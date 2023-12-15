package logger

import (
	"fmt"
	"log"
)

func CheckError(err error, info string) (res bool) {
	if err != nil {
		Errorf("%s: %s", info, err.Error())
		return false
	}
	return true
}

func Infof(info string, v ...interface{}) {
	s := fmt.Sprintf(info, v...)
	log.Printf("INFO|%s\n", s)
}

func Warnf(warn string, v ...interface{}) {
	s := fmt.Sprintf(warn, v...)
	log.Printf("WARN|%s\n", s)
}

func Errorf(error string, v ...interface{}) {
	s := fmt.Sprintf(error, v...)
	log.Printf("ERROR|%s\n", s)
}
