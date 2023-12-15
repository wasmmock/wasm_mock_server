package etc_creator

import (
	"errors"
	"log"
	"os"
)

func CreateReportFolder() {
	if _, err := os.Stat("report"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("report", os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
}
func CreateCertfolder() {
	//"cert_folder/proxy-ca.pem", "cert_folder/proxy-ca.key"
	if _, err := os.Stat("cert_folder"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("cert_folder", os.ModePerm)
		if err != nil {
			log.Println(err)
		} else {
			if f, err := os.Create("cert_folder/proxy-ca.pem"); err == nil {
				defer f.Close()
				f.WriteString(`-----BEGIN CERTIFICATE-----
				MIIDNTCCAh2gAwIBAgIUW8pt3M9cLDFJboyrBaHaZgu9x2UwDQYJKoZIhvcNAQEL
				BQAwKDESMBAGA1UEAwwJbWl0bXByb3h5MRIwEAYDVQQKDAltaXRtcHJveHkwHhcN
				MjMwMTEwMTM1OTUxWhcNMzMwMTA5MTM1OTUxWjAoMRIwEAYDVQQDDAltaXRtcHJv
				eHkxEjAQBgNVBAoMCW1pdG1wcm94eTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
				AQoCggEBAKQxP+v2HmkLKoZ/fB5Hm74wtH17TgtvFp6uSq7cmYW9+BD6l95oUxWG
				hJRqNekW2I5Xi5uz2MKZb7RnOl22MPF51SD/rLDFSVhn33WEvqblcdJ4ehffZI04
				ZaMNZH0rbcQtjkXUhfvgyPUqCImFK1GesRg16W9E6tOT81qEjM1XWF8giFIUn9E+
				uiNAlmk/GWSZueK7eU/+LKEMJjeOLvcUKPi6DDks0WrrEnWqoTlicOtgbmbJTqV/
				fzopRfnAaoL62/MLD+SGHNG4Xnua0NU5xVs4omMel9Q+xbWMRjgG/xRSA7EViJkU
				KZ1HEAD3fgkBuvvBJRIjAtJl5wz50DECAwEAAaNXMFUwDwYDVR0TAQH/BAUwAwEB
				/zATBgNVHSUEDDAKBggrBgEFBQcDATAOBgNVHQ8BAf8EBAMCAQYwHQYDVR0OBBYE
				FFVjL/yGhTnQ17PmWf5tz5XEN7otMA0GCSqGSIb3DQEBCwUAA4IBAQCN22yS3VAs
				a0ideNn3DXlubHVEt2rk+/btMtoXTgWOx5Ew0lI5RaHNyDAZm3vAId3Bs2okD52t
				ZKkcWgjE9MlyoYXgGfRwoy/W+jszkOFvTYbkpvWReA1rMXVBcoQJCTdpNJx4jxAB
				j3zZNpdZADFotAzXYpuGS6N37o95HjipVteCmkmtsAG7RK0kEJAuThJU3+Jc45u6
				CFZ1b0TVImJSjnmcgcxyn0uvP1ZSckKJQde/P2UiVRXD3VRFKl49H2aOf4jcH8II
				ka1BxSDp90uoES2HX95ilq4ehtjUYmBCU7PMNiuW4t4va+AsjyGl4laQf0PiMtkv
				SOfoLSe0vi8+
				-----END CERTIFICATE-----
				`)
			}
			if f, err := os.Create("cert_folder/proxy-ca.pem"); err == nil {
				defer f.Close()
				f.WriteString(`-----BEGIN RSA PRIVATE KEY-----
				MIIEowIBAAKCAQEApDE/6/YeaQsqhn98HkebvjC0fXtOC28Wnq5KrtyZhb34EPqX
				3mhTFYaElGo16RbYjleLm7PYwplvtGc6XbYw8XnVIP+ssMVJWGffdYS+puVx0nh6
				F99kjThlow1kfSttxC2ORdSF++DI9SoIiYUrUZ6xGDXpb0Tq05PzWoSMzVdYXyCI
				UhSf0T66I0CWaT8ZZJm54rt5T/4soQwmN44u9xQo+LoMOSzRausSdaqhOWJw62Bu
				ZslOpX9/OilF+cBqgvrb8wsP5IYc0bhee5rQ1TnFWziiYx6X1D7FtYxGOAb/FFID
				sRWImRQpnUcQAPd+CQG6+8ElEiMC0mXnDPnQMQIDAQABAoIBAADk75OK4hRrSUzb
				vDLHBZGWEgZkvRW7W5zU8RbkOf+iof7OFDTG9GLkejh7uDW/8droN6kX5B+Yadiw
				vwtJHxmfMs25jsV7GzPpah82Y22k+y2s1pwzYFV5+YKKlvt5vwda2/cRGgMd8MJD
				FRi4pxx/i1JI/HxzNeTRBFOMvF6W77LolZX+sSPIdjoicWEUTTvNI0zLrLbhAQ3U
				OMq9kGBYdJow5gm3HODO/FKtEuhzvBfSGVYGWh9PByT4uOPoCo2HbXyrPrP7kIYQ
				hYmqllpN9Iy1B/m0+23sxRuLYmiVDdu22/dcBLTwUflCN5WdO4soC903Xo5UEcis
				u3S+ECUCgYEAtWSu5/Qg3J6q0Bqvvqv247K6iAfwkq7CfqUVi9gKzquGBXI7ZM/B
				ZMY3cu12VZr5s+tvaM+hn3L0DeR8kCPnyyGBdykvT1jfDzUtbR7aWKVQvXb13zDy
				K++ofdWUpilu769FjEWETPE4Eu/Va/gg57wcbZxmmS9OC6EDPOss+/0CgYEA57lx
				YSBBDOgPlftRwZkmRwxv7BWMZBA0YGWM8kFsRUlF9JaF8tBoPNPYyWhKrtvglXFn
				/5W6rIlAo0mBonyR1F3Lba8eyxqu1IClw5QZyA1DrimgZ15h+CJD/v+Qq7w3SFOA
				PAYPe3HDsf/kJQi36/nso39jUOYmCOleRIZWCUUCgYEAh/OZP9HyW0g0X9rQg4jh
				dxE6yr+gqF+A+GiEeJaIqxNVVHmkWE959Cy33FRrz4dixV2c16Je0WHX1x272lB+
				5vkKzqO4iLDkJcLGdDWekrf8hrRFXW2S5CkwUHemfM8rDUuBRbvIh953F4JXpB+J
				kgWkDOce4orY5NBd2+erhIUCgYB5OH2jfagKBGwC3dJLIL7xdAQo0Kz8u61qsDUn
				lin5pPc/mG7CM1wUVg6WbkSDbOrzwrvQ7JcXI0X5Jb73LYtsORTucCn/vhmveQ0+
				Xv+Ns8KwHX6YFLvTfrlrcG5SKMgSwfvXaqQ2w7DIMUE8Tm9Itxmf/kgKThufldWG
				q2/esQKBgBl0+4beg1BNc0c6n3nb4FVVg1JcgGWOawKu+wnwrqFtXa6u/y6ZPUD7
				RQWTzShnhVylxoTOxN1o+LWCy0VEjiPEm5cd20drL7jKom1iJ0LQUNnsx/CsdL31
				KYspUpwitU26M4jrrF+l1qngrs7JTsaHg3VqaOAm3QQX0TrHfwC1
				-----END RSA PRIVATE KEY-----`)
			}
		}
	}
}
