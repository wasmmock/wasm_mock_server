package tcpproxy

import (
	b64 "encoding/base64"
	"sync"
	"time"

	"github.com/wasmmock/wasm_mock_server/capabilities"
)

type safeIdMapStruct struct {
	mu          sync.Mutex
	outstanding []capabilities.EntityWhole
}

func (s *safeIdMapStruct) append(e capabilities.EntityWhole) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outstanding = append(s.outstanding, e)
}
func (s *safeIdMapStruct) remove(e capabilities.EntityWhole, tcpNewWriteReqResponderChan map[string]chan capabilities.TcpReq) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := e.Time
	i := 0
	not_the_same := false
	for _, o := range s.outstanding {
		if t.Sub(o.Time) < 5*time.Second {
			if o.Id != e.Id { //not the same
				s.outstanding[i] = o
				i++
				not_the_same = true
			} else {
				s.RecordSuccess(tcpNewWriteReqResponderChan, e)
			}
		} else {
			//time_out
			s.RecordFailure(tcpNewWriteReqResponderChan, e)
		}
	}
	s.outstanding = s.outstanding[:i]
	return not_the_same
}
func (s *safeIdMapStruct) remove2(e capabilities.EntityWhole) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := e.Time
	i := 0
	not_the_same := false
	for _, o := range s.outstanding {
		if t.Sub(o.Time) < 5*time.Second {
			if o.Id != e.Id { //not the same
				s.outstanding[i] = o
				i++
				not_the_same = true
			} else {
				if e.Id != "" {
					s.RecordSuccess2(e)
				}
			}
		} else {
			//time_out
			if e.Id != "" {
				s.RecordFailure2(e)
			}
		}
	}
	s.outstanding = s.outstanding[:i]
	return not_the_same
}
func (s *safeIdMapStruct) SelfRemove() {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := time.Now()
	i := 0
	for _, o := range s.outstanding {
		if t.Sub(o.Time) < 5*time.Second {
			s.outstanding[i] = o
			i++
		} else {
			//time_out
			if o.Id != "" {
				s.RecordFailure2(o)
			}
		}
	}
	s.outstanding = s.outstanding[:i]
}
func (s *safeIdMapStruct) len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.outstanding)
}
func (s *safeIdMapStruct) RecordSuccess(tcpNewWriteReqResponderChan map[string]chan capabilities.TcpReq, e capabilities.EntityWhole) {
	if t, ok := tcpNewWriteReqResponderChan[e.Id]; ok {
		sDec := b64.StdEncoding.EncodeToString(e.Payload)
		t <- capabilities.TcpReq{Payload: sDec}
	}
}
func (s *safeIdMapStruct) RecordFailure(tcpNewWriteReqResponderChan map[string]chan capabilities.TcpReq, e capabilities.EntityWhole) {
	if t, ok := tcpNewWriteReqResponderChan[e.Id]; ok {
		t <- capabilities.TcpReq{Timeout: true}
	}
}
func (s *safeIdMapStruct) RecordSuccess2(e capabilities.EntityWhole) {
	if t, ok := TcpReqResponderChan[e.Id]; ok {
		sDec := b64.StdEncoding.EncodeToString(e.Payload)
		t <- capabilities.TcpReq{Payload: sDec}
	}
}
func (s *safeIdMapStruct) RecordFailure2(e capabilities.EntityWhole) {
	if t, ok := TcpReqResponderChan[e.Id]; ok {
		t <- capabilities.TcpReq{Timeout: true}
	}
}
