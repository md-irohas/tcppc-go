package tcppc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"net"
	"syscall"
	"unsafe"
)

func getOrigDst(oob []byte, oobn int) (*net.UDPAddr, error) {
	msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return nil, err
	}

	var origDst *net.UDPAddr

	for _, msg := range msgs {
		if msg.Header.Level == syscall.SOL_IP && msg.Header.Type == syscall.IP_RECVORIGDSTADDR {
			origDstRaw := &syscall.RawSockaddrInet4{}
			if err := binary.Read(bytes.NewReader(msg.Data), binary.LittleEndian, origDstRaw); err != nil {
				return nil, err
			}

			switch origDstRaw.Family {
			case syscall.AF_INET:
				pp := (*syscall.RawSockaddrInet4)(unsafe.Pointer(origDstRaw))
				p := (*[2]byte)(unsafe.Pointer(&pp.Port))

				origDst = &net.UDPAddr{
					IP:   net.IPv4(pp.Addr[0], pp.Addr[1], pp.Addr[2], pp.Addr[3]),
					Port: int(p[0])<<8 + int(p[1]),
				}

			default:
				return nil, errors.New("Unsupported network family.")
			}
		}
	}

	return origDst, err
}

func HandleUDPSession(src, dst *net.UDPAddr, buf []byte, length int, writer *RotWriter) {
	flow := NewUDPFlow(src, dst)
	session := NewSession(flow)

	data := make([]byte, length)
	copy(data, buf[:length])

	session.AddPayload(data)

	log.Printf("UDP: Received: %s: %q (%d bytes)\n", session, buf[:length], length)

	if writer != nil {
		outputJson, err := json.Marshal(session)
		if err == nil {
			log.Printf("Wrote data: %s", session)
			writer.Write(outputJson)
		} else {
			log.Printf("Failed to encode data as json: %s\n", err)
		}
	}
}

func StartUDPServer(host string, port int, writer *RotWriter) {
	log.Printf("Server Mode: UDP\n")
	log.Printf("Listen: %s:%d\n", host, port)

	addr := &net.UDPAddr{
		IP:   net.ParseIP(host),
		Port: int(port),
	}

	ln, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen UDP socket: %s\n", err)
	}
	defer ln.Close()

	file, err := ln.File()
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

	log.Printf("Start UDP server.\n")

	for {
		buf := make([]byte, 2048)
		oob := make([]byte, 1024)

		length, oobn, _, src, err := ln.ReadMsgUDP(buf, oob)
		if err != nil {
			log.Printf("Failed to read UDP message: %s\n", err)
			continue
		}

		origDst, err := getOrigDst(oob, oobn)
		if err != nil {
			log.Printf("Failed to get the original destination: %s\n", err)
			continue
		}

		go HandleUDPSession(src, origDst, buf, length, writer)
	}
}
