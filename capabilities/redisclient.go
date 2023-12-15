package capabilities

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type Bytes []byte

func (m *Bytes) Marshal() (dAtA []byte, err error) {
	return *m, nil
}
func (m *Bytes) Unmarshal(dAtA []byte) error {
	*m = dAtA
	return nil
}

func RedisGetKey(addr string, req []byte) ([]byte, error) {
	addrList := strings.Split(addr, ",")
	addrsMap := make(map[string]string)
	for _, item := range addrList {
		if strings.Contains(item, ":") {
			partList := strings.Split(item, ":")
			addrsMap[partList[0]] = ":" + partList[1]
		}
	}
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: addrsMap,
	})

	mycache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
	var wanted = ""

	if err := mycache.Get(context.Background(), string(req), &wanted); err == nil {
		return []byte(wanted), nil
	} else {
		return []byte(wanted), err
	}
}
func RedisExistKey(addr string, req []byte, flag bool) ([]byte, error) {
	addrList := strings.Split(addr, ",")
	addrsMap := make(map[string]string)
	for _, item := range addrList {
		if strings.Contains(item, ":") {
			partList := strings.Split(item, ":")
			addrsMap[partList[0]] = ":" + partList[1]
		}
	}
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: addrsMap,
	})

	mycache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	if mycache.Exists(context.Background(), string(req)) {
		if flag {
			return []byte{}, nil
		}
		return []byte{}, fmt.Errorf("exists")
	}
	if !flag {
		return []byte{}, nil
	}
	return []byte{}, fmt.Errorf("does not exist")
}
func RedisDeleteKey(addr string, req []byte) ([]byte, error) {
	addrList := strings.Split(addr, ",")
	addrsMap := make(map[string]string)
	for _, item := range addrList {
		if strings.Contains(item, ":") {
			partList := strings.Split(item, ":")
			addrsMap[partList[0]] = ":" + partList[1]
		}
	}
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: addrsMap,
	})

	mycache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})
	var wanted = Bytes{}
	if err := mycache.Delete(context.Background(), string(req)); err == nil {
		return wanted, nil
	} else {
		return wanted, err
	}
}
