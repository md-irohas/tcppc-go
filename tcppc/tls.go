package tcppc

import (
	"log"
	"net"
	"time"
	"encoding/json"
	"syscall"
	"github.com/md-irohas/tcppc-go/crypto/tls"
)

func HandleTLSSession(conn *tls.Conn, writer *RotWriter, timeout int) {
	defer conn.Close()

	var src, dst *net.TCPAddr
	src = conn.RemoteAddr().(*net.TCPAddr)
	dst = conn.LocalAddr().(*net.TCPAddr)

	flow := NewTCPFlow(src, dst)
	session := NewSession(flow)

	log.Printf("TLS: Established: %s (#Sessions: %d)\n", session, counter.count())

	var length uint
	var err error

	buf := make([]byte, 4096)

	for {
		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))

		length, err := conn.Read(buf)
		if err != nil {
			break
		}

		data := make([]byte, length)
		copy(data, buf[:length])

		session.AddPayload(data)

		log.Printf("TLS: Received: %s: %q (%d bytes)\n", session, buf[:length], length)
	}

	if writer != nil {
		outputJson, err := json.Marshal(session)
		if err == nil {
			log.Printf("Wrote data: %s", session)
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

func StartTLSServer(host string, port int, config *tls.Config, writer *RotWriter, timeout int) {
	log.Printf("Server Mode: TLS\n")
	log.Printf("Listen: %s:%d\n", host, port)

	addr := &net.TCPAddr {
		IP: net.ParseIP(host),
		Port: port,
	}

	tcpLn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen TCP socket: %s\n", err)
	}
	defer tcpLn.Close()

	file, err := tcpLn.File()
	if err != nil {
		log.Fatalf("Failed to get a file descriptor of the listener: %s", err)
	}
	defer file.Close()

	fd := int(file.Fd())
	if err := syscall.SetsockoptInt(fd, syscall.SOL_IP, syscall.IP_TRANSPARENT, 1); err != nil {
		log.Fatalf("Failed to set socket option (IP_TRANSPARENT): %s\n", err)
	}
	if err := syscall.SetsockoptInt(fd, syscall.SOL_IP, syscall.IP_RECVORIGDSTADDR, 1); err != nil {
		log.Fatalf("Failed to set socket option (IP_RECVORIGDSTADDR): %s\n", err)
	}

	ln := tls.NewListener(tcpLn, config)

	log.Printf("Start TLS server.\n")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("Failed to accept a new connection: %s\n", err)
		}

		go HandleTLSSession(conn.(*tls.Conn), writer, timeout)
	}
}
