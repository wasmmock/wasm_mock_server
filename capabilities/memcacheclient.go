package capabilities

import (
	"fmt"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

func MemcacheGetKey(addr string, req []byte) ([]byte, error) {
	addrList := strings.Split(addr, ",")
	mc := memcache.New(addrList...)
	it, err := mc.Get(string(req))
	if err != nil {
		return []byte{}, err
	}
	return it.Value, nil
}
func MemcacheExistKey(addr string, req []byte, flag bool) ([]byte, error) {
	addrList := strings.Split(addr, ",")
	mc := memcache.New(addrList...)
	_, err := mc.Get(string(req))
	if err == nil {
		if flag {
			return []byte{}, nil
		}
		return []byte{}, fmt.Errorf("exists")
	}
	if !flag {
		return []byte{}, nil
	}
	return []byte{}, fmt.Errorf("does not exists")
}
func MemcacheDeleteKey(addr string, req []byte) ([]byte, error) {
	addrList := strings.Split(addr, ",")
	mc := memcache.New(addrList...)
	err := mc.Delete(string(req))
	return []byte{}, err
}
