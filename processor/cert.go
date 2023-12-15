package processor

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"
	"sync"

	"github.com/wasmmock/wasm_mock_server/model"
	"github.com/wasmmock/wasm_mock_server/util"
)

var certFile map[string][]byte = make(map[string][]byte)
var fiddleCertCache = CertCache{}

func SetCertFile(rw http.ResponseWriter, req *http.Request) {
	response := new(model.CallApiResponse)
	response.Header.Message = "ok"
	code, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	host := req.URL.Query()["targets"][0] //host
	certFile[host] = code
}

type CertCache struct {
	m sync.Map
}

func (c *CertCache) Set(host string, cert *tls.Certificate) {
	c.m.Store(host, cert)
}
func (c *CertCache) Get(host string) *tls.Certificate {
	v, ok := c.m.Load(host)
	if !ok {
		return nil
	}

	return v.(*tls.Certificate)
}
func UseKeyFile(rw http.ResponseWriter, req *http.Request) {
	response := new(model.CallApiResponse)
	response.Header.Message = "ok"
	host := req.URL.Query()["targets"][0] //host
	sudo_password := req.URL.Query()["sudo_password"][0]
	if certPem, ok := certFile[host]; ok {
		keyPem, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			response.Header.Message = err.Error()
			util.JsonResponse(rw, response)
			return
		}
		certPair, err := tls.X509KeyPair(certPem, keyPem)
		if err != nil {
			response.Header.Message = "Cert X509KeyPair error"
			util.JsonResponse(rw, response)
			return

		}

		fiddleCertCache.Set(host, &certPair)
		if runtime.GOOS == "linux" {
			fmt.Println("Hello from linux")
			err := ioutil.WriteFile("/usr/local/share/ca-certificates/"+host+".crt", certPem, 0644)
			if err != nil {
				response.Header.Message = "save to ca-certificates err"
				util.JsonResponse(rw, response)
				return
			}
			cmd := exec.Command("echo", sudo_password, "|", "sudo", "-S", "update-ca-certificates")
			stdout, err := cmd.Output()
			if err != nil {
				response.Header.Message = "save to ca-certificates err"
				util.JsonResponse(rw, response)
				return
			} else {
				response.Header.Message = string(stdout)
				util.JsonResponse(rw, response)
				return
			}
		} else if runtime.GOOS == "darwin" {
			fmt.Println("Hello from darwin")
			err := ioutil.WriteFile(host+".crt", certPem, 0644)
			cmd := exec.Command("echo", sudo_password, "|", "sudo", "-S", "security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", host+".crt")
			stdout, err := cmd.Output()
			if err != nil {
				response.Header.Message = "save to ca-certificates err"
				util.JsonResponse(rw, response)
				return
			} else {
				response.Header.Message = string(stdout)
				util.JsonResponse(rw, response)
				return
			}
		}
		// fiddlerProxy.Cert = cert.NewCertificate(&c)
		// fiddlerProxy.DecryptHTTPS = true
	} else {
		response.Header.Message = "Set Cert file first"
		util.JsonResponse(rw, response)
		return
	}

}
