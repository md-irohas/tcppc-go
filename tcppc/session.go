package tcppc

import (
	"fmt"
	"github.com/jehiah/go-strftime"
	"net"
	"time"
)

const (
	TIME_FMT = "%Y-%m-%dT%H:%M:%S"
)

func formatTimeStr(t *time.Time) string {
	return strftime.Format(TIME_FMT, *t)
}

type TCPFlow struct {
	Proto string `json:"proto"`
	Src   net.IP `json:"src"`
	Sport int    `json:"sport"`
	Dst   net.IP `json:"dst"`
	Dport int    `json:"dport"`
}

func NewTCPFlow(src, dst *net.TCPAddr) *TCPFlow {
	return &TCPFlow{"tcp", src.IP, src.Port, dst.IP, dst.Port}
}

func (f *TCPFlow) String() string {
	return fmt.Sprintf("TCPFlow: %s:%d <-> %s:%d", f.Src.String(), f.Sport, f.Dst.String(), f.Dport)
}

type TCPPayload struct {
	Index     uint      `json:"index"`
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
}

func NewTCPPayload(index uint, timestamp time.Time, data []byte) *TCPPayload {
	return &TCPPayload{index, timestamp, data}
}

func (p *TCPPayload) String() string {
	return fmt.Sprintf("TCPPayload %d: %s: %v", p.Index, formatTimeStr(&p.Timestamp), p.Data)
}

type TCPSession struct {
	Timestamp time.Time     `json:"timestamp"`
	Flow      *TCPFlow      `json:"flow"`
	Payloads  []*TCPPayload `json:"payloads"`
}

func NewTCPSession(flow *TCPFlow) *TCPSession {
	return &TCPSession{Timestamp: time.Now(), Flow: flow}
}

func (s *TCPSession) String() string {
	return fmt.Sprintf("TCPSession: %s: %s (%d payloads)", formatTimeStr(&s.Timestamp), s.Flow, len(s.Payloads))
}

func (s *TCPSession) AddPayload(data []byte) *TCPPayload {
	index := uint(len(s.Payloads))
	ts := time.Now()

	payload := NewTCPPayload(index, ts, data)
	s.Payloads = append(s.Payloads, payload)

	return payload
}
