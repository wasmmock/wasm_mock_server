package util

import (
	"bytes"
	"compress/gzip"
	b64 "encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

func HttpContentEncodingDecode(header http.Header, body io.ReadCloser, len int64, typ string) (string, error) {
	if len > 100000 {
		return "", fmt.Errorf("Len larger than 100000")
	}

	//defer body.Close()
	gzipBody, err := ioutil.ReadAll(body)
	var b bytes.Buffer
	b.Write(gzipBody)
	body = ioutil.NopCloser(&b)
	if err != nil {
		fmt.Println(typ+".Body err is not nil", err, body == nil)
		return "", nil
	}
	var reader io.ReadCloser
	if strings.Contains(header.Get("Content-Encoding"), "gzip") {
		reader, err = gzip.NewReader(bytes.NewBuffer(gzipBody))
	} else if strings.Contains(header.Get("Content-Encoding"), "br") {
		var b = make([]byte, 60000)
		n, nerr := brotli.NewReader(bytes.NewBuffer(gzipBody)).Read(b)
		if nerr != nil {
			return "", nerr
		}
		if n < 60000 {
			data := b[:n]
			reader = io.NopCloser(bytes.NewReader(data))
		} else {
			fmt.Println("n > 60000")
			return "", nil
		}

	} else {
		reader = ioutil.NopCloser(bytes.NewReader(gzipBody))
	}
	if err != nil {
		fmt.Println("err____", err)
		return "", err
	}
	body_new, read_err := ioutil.ReadAll(reader)
	//resp.Body = ioutil.NopCloser(bytes.NewBuffer(gzipBody))
	if read_err != nil {
		fmt.Println("err____", err)
		return "", err
	}
	return b64.StdEncoding.EncodeToString(body_new), nil
}
