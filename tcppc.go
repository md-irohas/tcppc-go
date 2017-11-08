package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
	"time"
)

const (
	// Version number
	version = "0.1.2"
)

var (
	sessionCount uint
)

func handleRequest(conn net.Conn) {
	sessionCount++

	defer conn.Close()
	defer func() {
		sessionCount--
	}()

	var length int
	var err error

	buf := make([]byte, 2048)
	ftuple := conn.RemoteAddr().String() + "<->" + conn.LocalAddr().String()

	log.Printf("Established: %s (#Sessions: %d)\n", ftuple, sessionCount)

	for {
		length, err = conn.Read(buf)
		if err != nil {
			break
		}

		log.Printf("%s: %q (%d bytes)\n", ftuple, buf[:length], length)
	}

	if length == 0 {
		log.Printf("Closed: %s (#Sessions: %d)\n", ftuple, sessionCount)
	} else {
		log.Printf("Aborted: %s %s (#Sessions: %d)\n", ftuple, err.Error(), sessionCount)
	}
}

func main() {
	var host string
	var port uint
	var tcpTimeout int
	var logFile string
	var maxFdNum uint64
	var showVersion bool

	// Parse params from command-line arguments.
	flag.StringVar(&host, "H", "", "hostname to listen on.")
	flag.UintVar(&port, "p", 12345, "port number to listen on.")
	flag.IntVar(&tcpTimeout, "t", 60, "timeout for TCP connection.")
	flag.StringVar(&logFile, "L", "", "log file.")
	flag.Uint64Var(&maxFdNum, "R", 0, "set maximum number of file descriptors (need root priviledge in some environments).")
	flag.BoolVar(&showVersion, "v", false, "show version and exit.")
	flag.Parse()

	if showVersion {
		fmt.Println("version:", version)
		return
	}

	log.SetFlags(log.LstdFlags)

	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
		if err != nil {
			log.Fatal("Failed to open log file:", err.Error())
		}

		log.SetOutput(f)
		log.Println("Open log file:", logFile)
	}

	var rLimit syscall.Rlimit
	if maxFdNum > 0 {
		rLimit.Max = maxFdNum
		rLimit.Cur = maxFdNum

		err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			log.Printf("Failed to set maximum number of file descriptors:", err)
		}
	}

	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
    if err == nil {
		log.Println("Maximum number of file descriptors:", rLimit.Cur)
    }

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatal("Failed to listen:", err.Error())
	}
	defer l.Close()

	log.Printf("Listen on %s:%d\n", host, port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Failed to accept:", err.Error())
		}

		conn.SetDeadline(time.Now().Add(time.Duration(tcpTimeout) * time.Second))

		go handleRequest(conn)
	}
}
