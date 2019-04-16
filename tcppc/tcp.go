package tcppc

import (
	"encoding/json"
	"log"
	"net"
	"syscall"
	"time"
)

func HandleTCPSession(conn *net.TCPConn, writer *RotWriter, timeout int) {
	defer conn.Close()
	defer counter.dec()
	counter.inc()

	var src, dst *net.TCPAddr
	src = conn.RemoteAddr().(*net.TCPAddr)
	dst = conn.RemoteAddr().(*net.TCPAddr)

	flow := NewTCPFlow(src, dst)
	session := NewSession(flow)

	log.Printf("TCP: Established: %s (#Sessions: %d)\n", session, counter.count())

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

		log.Printf("TCP: Received: %s: %q (%d bytes)\n", session, buf[:length], length)
	}

	if writer != nil {
		outputJson, err := json.Marshal(session)
		if err == nil {
			log.Printf("Wrote data: %s\n", session)
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

func StartTCPServer(host string, port int, writer *RotWriter, timeout int) {
	log.Printf("Server Mode: TCP\n")
	log.Printf("Listen: %s:%d\n", host, port)

	addr := &net.TCPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen TCP socket: %s\n", err)
	}
	defer ln.Close()

	file, err := ln.File()
	if err != nil {
		log.Fatalf("Failed to get a file descriptor of the listener: %s\n", err)
	}
	defer file.Close()

	fd := int(file.Fd())
	if err := syscall.SetsockoptInt(fd, syscall.SOL_IP, syscall.IP_TRANSPARENT, 1); err != nil {
		log.Fatalf("Failed to set socket option (IP_TRANSPARENT): %s\n", err)
	}
	if err := syscall.SetsockoptInt(fd, syscall.SOL_IP, syscall.IP_RECVORIGDSTADDR, 1); err != nil {
		log.Fatalf("Failed to set socket option (IP_RECVORIGDSTADDR): %s\n", err)
	}

	log.Printf("Start TCP server.\n")

	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			log.Fatalf("Failed to accept a new connection: %s\n", err)
		}

		go HandleTCPSession(conn, writer, timeout)
	}
}
