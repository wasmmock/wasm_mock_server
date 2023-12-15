package Security

import (
	"fmt"
	"testing"
)

func UserPermissionJWT(email string, secret string) (string, error) {
	var permission = ""
	if email == "test.email" {
		permission = "ok"
	}
	return GenNewBusinessJWT(email, permission, secret)
}
func Test_UserPermissionJWT(t *testing.T) {
	email := "test.email"
	secret := "aaas"
	t.Logf("Say hi2 %s", "token")
	fmt.Println("jj")
	permission_jwt, er := UserPermissionJWT(email, secret)
	if er == nil {
		if user_email, er := ParseUserPermissionJWT(permission_jwt, secret); er == nil {
			_ = user_email
			t.Logf("Say user_email %s", user_email)
		}
	} else {
		t.Logf("Say er %s", er.Error())
	}
}
