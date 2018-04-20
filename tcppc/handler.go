package tcppc

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"net"
	"runtime"
	"sync"
	"syscall"
	"github.com/md-irohas/tcppc-go/crypto/tls"
)

const (
	SO_ORIGINAL_DST = 80
)

var (
	counter = NewSessionCounter()
)

type SessionCounter struct {
	Count uint
	mutex sync.RWMutex
}

func NewSessionCounter() *SessionCounter {
	return &SessionCounter{}
}

func (c *SessionCounter) inc() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Count += 1
}

func (c *SessionCounter) dec() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Count -= 1
}

func (c *SessionCounter) count() uint {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Count
}

func getOriginalDst(conn net.Conn) (*net.TCPAddr, error) {
	if runtime.GOOS != "linux" {
		return nil, errors.New("'getOriginalDst' is only supported in Linux.")
	}

	var file *os.File
	var err error

	if _, ok := conn.(*net.TCPConn); ok {
		file, err = conn.(*net.TCPConn).File()
		if err != nil {
			return nil, err
		}
	} else if _, ok := conn.(*tls.Conn); ok {
		file, err = conn.(*tls.Conn).File()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Unknown Conn instance.")
	}

	origDstRaw, err := syscall.GetsockoptIPv6Mreq(int(file.Fd()), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		return nil, err
	}

	ar := origDstRaw.Multiaddr[4:8]
	pr := origDstRaw.Multiaddr[2:4]

	origDst := &net.TCPAddr{}
	origDst.IP = net.IPv4(ar[0], ar[1], ar[2], ar[3])
	origDst.Port = (int(pr[0]) << 8) + int(pr[1])

	return origDst, nil
}

func HandleRequest(conn net.Conn, writer *RotWriter) {
	counter.inc()
	defer counter.dec()
	defer conn.Close()

	var err error

	var src, dst *net.TCPAddr
	src = conn.RemoteAddr().(*net.TCPAddr)

	if runtime.GOOS == "linux" {
		dst, err = getOriginalDst(conn)
		if err != nil {
			log.Printf("Failed to get original dst: %s", err)
		}
	}

	if dst == nil {
		dst = conn.LocalAddr().(*net.TCPAddr)
	}

	flow := NewTCPFlow(src, dst)
	session := NewTCPSession(flow)

	log.Printf("Established: %s (#Sessions: %d)\n", session, counter.count())

	var length uint

	buf := make([]byte, 2048)

	for {
		length, err := conn.Read(buf)
		if err != nil {
			break
		}

		data := make([]byte, length)
		copy(data, buf[:length])

		session.AddPayload(data)

		log.Printf("Received: %s: %q (%d bytes)\n", session, buf[:length], length)
	}

	if writer != nil {
		outputJson, err := json.Marshal(session)
		if err == nil {
			log.Printf("Wrote: %s", session)
			writer.Write(outputJson)
		} else {
			log.Printf("Failed to encode data as json: %s\n", err)
		}
	}

	if length == 0 {
		log.Printf("Closed: %s (#Sessions: %d)\n", session, counter.count())
	} else {
		log.Printf("Aborted: %s %s (#Sessions: %d)\n", session, err, counter.count())
	}
}
