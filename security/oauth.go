package Security

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func Oauth() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		response_type := ""
		client_id := ""
		state := ""
		redirect_uri := ""
		location := fmt.Sprintf("/o/oauth2/v2/auth?response_type=%s&client_id=%s&state=%s&redirect_uri=%s", response_type, client_id, state, redirect_uri)
		rw.Header().Add("Location", location)
		rw.WriteHeader(302)
	}
}

type OauthLoginRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
}
type OauthLoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int32  `json:"expires_in"`
	Scope       string `json:"scope"`
}
type OauthLoginFEResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int32  `json:"expires_in"`
	Scope       string `json:"scope"`
}
type OauthUserInfoResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int32  `json:"expires_in"`
	Scope       string `json:"scope"`
}

//https://jgurda.medium.com/how-to-mock-google-oauth2-api-in-three-easy-steps-28a908563c1e
var (
	verifyKey *rsa.PublicKey

	serverPort int
)

func IsJWTValid(credential string) (string, string, bool) {
	//g_csrf_token := req.Form.Get("g_csrf_token")
	signature, msg_hash, req_kid, email, exptime := parseJWT(credential)
	req_pubkey := check_publickeys(current_keys, req_kid)
	time_now := time.Now().Unix()
	time_expiry, _ := strconv.ParseInt(exptime, 10, 64)
	var validToken = false
	if time_now < time_expiry {
		err := rsa.VerifyPKCS1v15(&req_pubkey, crypto.SHA256, msg_hash, signature)
		if err != nil {
			// w.WriteHeader(400)
			// fmt.Fprintf(w, "Error : Invalid Token")
		} else {
			validToken = true
			if v, ok := SecurityMap["Remote-Verification"]; ok {
				var rc_config = RemoteVerificationConfig{}
				if err := json.Unmarshal([]byte(v.ConfigValue), &rc_config); err != nil {
					fmt.Println("Security.SecurityMap[Remote-Verification] error: " + v.ConfigValue)
					return "", "", false
				}
				if rc_config.Enabled && rc_config.Secret != "" {
					return "", rc_config.Secret, false
				}
			}
			//fmt.Fprintf(w, email)
		}
	} else {
		//w.WriteHeader(400)
		//fmt.Fprintf(w, "Error : expired token")
	}
	fmt.Println("validToken", validToken, email)
	return email, "", validToken

}

var StoredJWTToken string = ""

func GiveToken(token string, rw http.ResponseWriter, req *http.Request) {
	StoredJWTToken = token
}

var keys map[string]rsa.PublicKey
var req_pubkey rsa.PublicKey
var current_keys map[string]rsa.PublicKey

func check_publickeys(keys map[string]rsa.PublicKey, req_kid string) rsa.PublicKey {
	for k, v := range keys {
		if req_kid == k {
			req_pubkey = v
			return req_pubkey
		}
	}

	// get new keys and check again
	new_keys := get_publickeys()
	for k, v := range new_keys {
		if req_kid == k {
			req_pubkey = v
			return req_pubkey
		}
	}
	return req_pubkey
}

func parseJWT(token_raw string) ([]byte, []byte, string, string, string) {

	token_split := strings.Split(token_raw, ".")

	var header = urlsafeB64decode(token_split[0])
	var payload = urlsafeB64decode(token_split[1])
	var signature = urlsafeB64decode(token_split[2])

	msg_byte := calcSum(token_split[0] + "." + token_split[1])

	regex_kid := regexp.MustCompile(`.kid...([\w]+).`)
	regex_email := regexp.MustCompile(`\w+@\w+.\w+`)
	regex_exptime := regexp.MustCompile(`.exp.:(\d+)`)

	email := regex_email.FindString(string(payload))
	exptime := regex_exptime.FindStringSubmatch(string(payload))

	header_kid := regex_kid.FindStringSubmatch(string(header))
	req_kid := header_kid[1]

	return signature, msg_byte, req_kid, email, exptime[1]

}

func get_publickeys() map[string]rsa.PublicKey {
	jwk_get_resp, _ := http.Get("https://www.googleapis.com/oauth2/v3/certs")
	jwk_cont, _ := ioutil.ReadAll(jwk_get_resp.Body)

	regex_kid := regexp.MustCompile(`"kid": "(\S+)"`)
	regex_n := regexp.MustCompile(`"n": "(\S+)"`)
	regex_e := regexp.MustCompile(`"e": "(\S+)"`)

	kid := regex_kid.FindAllSubmatch(jwk_cont, -1)
	n := regex_n.FindAllSubmatch(jwk_cont, -1)
	e := regex_e.FindAllSubmatch(jwk_cont, -1)

	n1 := byteToInt(urlsafeB64decode(string(n[0][1])))
	n2 := byteToInt(urlsafeB64decode(string(n[1][1])))
	e1 := btrToInt(byteToBtr(urlsafeB64decode(string(e[0][1]))))
	e2 := btrToInt(byteToBtr(urlsafeB64decode(string(e[1][1]))))

	public_key1 := rsa.PublicKey{N: n1, E: e1}
	public_key2 := rsa.PublicKey{N: n2, E: e2}

	keys := map[string]rsa.PublicKey{
		string(kid[0][1]): public_key1,
		string(kid[1][1]): public_key2,
	}
	return keys
}

//////////////////////// Helper Functions ////////////////////////////////////
func byteToBtr(bt0 []byte) *bytes.Reader {
	var bt1 []byte
	if len(bt0) < 8 {
		bt1 = make([]byte, 8-len(bt0), 8)
		bt1 = append(bt1, bt0...)
	} else {
		bt1 = bt0
	}
	return bytes.NewReader(bt1)
}
func btrToInt(a io.Reader) int {
	var e uint64
	binary.Read(a, binary.BigEndian, &e)
	return int(e)
}
func urlsafeB64decode(str string) []byte {
	if m := len(str) % 4; m != 0 {
		str += strings.Repeat("=", 4-m)
	}
	bt, _ := base64.URLEncoding.DecodeString(str)
	return bt
}
func calcSum(str string) []byte {
	a := sha256.New()
	a.Write([]byte(str))
	return a.Sum(nil)
}
func byteToInt(bt []byte) *big.Int {
	a := big.NewInt(0)
	a.SetBytes(bt)
	return a
}
