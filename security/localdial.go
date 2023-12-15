package Security

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/audiolion/ipip"
	"github.com/google/uuid"
)

type SecurityResult struct {
	ConfigValue string
	Bool        bool
	ErrorMsg    string
}
type SecurityIpRange = map[string][]string

var SecurityMap = make(map[string]SecurityResult)
var X_Api_Key_Map = sync.Map{}

const UIDKey string = "uID"

func X_Api_Key(r *http.Request) bool {
	if v, ok := SecurityMap["X-Api-Key"]; ok {
		var ip_range = SecurityIpRange{}
		if err := json.Unmarshal([]byte(v.ConfigValue), &ip_range); err != nil {
			fmt.Println("Security.SecurityMap[X-Api-Key] error: " + v.ConfigValue)
		}
		api_key := r.Header.Get("X-Api-Key")

		//uIDContext := context.Get(r, UIDKey)
		uIDContext := r.Context().Value(UIDKey)
		if uIDContext == nil {
			fmt.Println("X_Api_Key uID error")
		}
		if uID, ok := uIDContext.(string); ok {
			X_Api_Key_Map.Store(uID, api_key)
		}
		var found = false
		for x, ip_ranges := range ip_range {
			if x == api_key && len(ip_ranges) == 0 {
				found = true
				break
			} else {
				for _, network := range ip_ranges {
					_, subnet, er := net.ParseCIDR(network)
					if er != nil {
						return true
					}
					ad, e := net.ResolveTCPAddr("tcp", r.RemoteAddr)
					if e == nil {
						if ipip.IsPrivate(ad.IP) {
							found = true
							break
						}
						if ad.IP.String() == "::1" {
							found = true
							break
						}
						if subnet.Contains(ad.IP) {
							found = true
							break
						}
					} else {
						fmt.Println("ResolveIPAddr", r.RemoteAddr, e.Error())
					}

				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func X_Api_Key_Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uID := uuid.New().String()
		ctx := r.Context()
		req := r.WithContext(context.WithValue(ctx, UIDKey, uID))
		*r = *req
		//context.Set(r, UIDKey, uID)
		if !X_Api_Key(r) {
			http.Redirect(w, r, "/error", http.StatusForbidden)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
