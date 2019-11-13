package tcppc

import (
	"fmt"
	"github.com/jehiah/go-strftime"
	"net"
	"time"
)

const (
	TIME_FMT = "%Y-%m-%dT%H:%M:%S%z"
)

func formatTimeStr(t *time.Time) string {
	return strftime.Format(TIME_FMT, *t)
}

type Flow struct {
	Proto string `json:"proto"`
	Src   net.IP `json:"src"`
	Sport int    `json:"sport"`
	Dst   net.IP `json:"dst"`
	Dport int    `json:"dport"`
}

func NewTCPFlow(src, dst *net.TCPAddr) *Flow {
	return &Flow{"tcp", src.IP, src.Port, dst.IP, dst.Port}
}

func NewTLSFlow(src, dst *net.TCPAddr) *Flow {
	return &Flow{"tls", src.IP, src.Port, dst.IP, dst.Port}
}

func NewUDPFlow(src, dst *net.UDPAddr) *Flow {
	return &Flow{"udp", src.IP, src.Port, dst.IP, dst.Port}
}

func (f *Flow) String() string {
	return fmt.Sprintf("Flow: %s %s:%d <-> %s:%d", f.Proto, f.Src.String(), f.Sport, f.Dst.String(), f.Dport)
}

type Payload struct {
	Index     uint      `json:"index"`
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
}

func NewPayload(index uint, timestamp time.Time, data []byte) *Payload {
	return &Payload{index, timestamp, data}
}

func (p *Payload) String() string {
	return fmt.Sprintf("Payload %d: %s: %v", p.Index, formatTimeStr(&p.Timestamp), p.Data)
}

type Session struct {
	Timestamp time.Time  `json:"timestamp"`
	Flow      *Flow      `json:"flow"`
	Payloads  []*Payload `json:"payloads"`
}

func NewSession(flow *Flow) *Session {
	return &Session{Timestamp: time.Now(), Flow: flow}
}

func (s *Session) String() string {
	return fmt.Sprintf("Session: %s: %s (%d payloads)", formatTimeStr(&s.Timestamp), s.Flow, len(s.Payloads))
}

func (s *Session) AddPayload(data []byte) *Payload {
	index := uint(len(s.Payloads))
	ts := time.Now()

	payload := NewPayload(index, ts, data)
	s.Payloads = append(s.Payloads, payload)

	return payload
}
