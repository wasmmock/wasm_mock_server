package Security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/golang-jwt/jwt"
)

type RemoteVerification struct {
	Addr          string
	Secret        string
	Last_verified time.Time
	Ok            bool
	CheckTime     time.Duration
}
type RemoteVerificationConfig struct {
	Enabled            bool   `json:"enabled"`
	Secret             string `json:"secret"`
	ServeEndpointLocal bool   `json:"serve_endpoint_local"`
}
type resVerification struct {
	JWT string `json:"jwt"`
}
type LoginCred struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

var RemoteVerificationState = RemoteVerification{

	//	Last_verified: time.Time(*time.UTC),
}
var Login = LoginCred{}

func GenNewBusinessJWT(email string, permission string, secret string) (string, error) {
	currentTime := time.Now()
	// find email in database first
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":      email,
		"permission": permission,
		"expiry":     currentTime.Add(time.Hour * 36).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(secret))
	fmt.Println("tokenString signed", tokenString)
	return tokenString, err
}
func ParseUserPermissionJWT(j_token string, secret string) (string, error) {
	fmt.Println("secret", secret)
	token, err := jwt.Parse(j_token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err == nil {
		fmt.Println("token", token)
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			fmt.Println("email", claims["email"], "expiry", claims["expiry"])
			exptime := claims["expiry"]
			email := claims["email"]
			permission := claims["permission"]
			fmt.Println("type of exptime", reflect.TypeOf(exptime).String())

			if exptime, ok := exptime.(float64); ok && permission == "ok" {
				if email, ok := email.(string); ok {
					time_now := time.Now().Unix()
					if err == nil {
						//time_expiry, _ := strconv.ParseInt(exptime, 10, 64)
						time_expiry := int64(exptime)
						if time_now < time_expiry {
							fmt.Println(": not email", email)
							return email, nil
						} else {
							//w.WriteHeader(400)
							fmt.Println("Error : expired token", time_expiry, time_now)
						}
					}

				}
			}

		} else {
			fmt.Println("err", err)
		}
	}
	return "", nil
}
func Remote_Verify_Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if RemoteVerificationState.Last_verified == nil {
		// 	return
		// }
		requestBody, err := json.Marshal(Login)
		if err != nil {
			return
		}
		var authenticated = false
		if time.Now().Sub(RemoteVerificationState.Last_verified) > RemoteVerificationState.CheckTime*time.Second {
			resp, err := http.Post(RemoteVerificationState.Addr+"/remote_verify", "application/json", bytes.NewBuffer(requestBody))

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {

			}
			var resKey = resVerification{}
			if err = json.Unmarshal(body, &resKey); err == nil {
				// token, err := jwt.Parse(resKey.JWT, func(token *jwt.Token) (interface{}, error) {
				// 	// Don't forget to validate the alg is what you expect:
				// 	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				// 		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				// 	}
				// 	// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
				// 	return []byte(hmacSampleSecret), nil
				// })
				// if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// 	fmt.Println(claims["foo"], claims["nbf"])
				// 	authenticated = true
				// } else {
				// 	fmt.Println(err)
				// }
			}
		}
		if !authenticated {
			http.Redirect(w, r, "/error", http.StatusForbidden)

		} else {
			h.ServeHTTP(w, r)
		}
	})
}
