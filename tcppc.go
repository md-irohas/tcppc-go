package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	// Version number
	version = "0.1.0"
)

func handleRequest(conn net.Conn) {
	var length int
	var err error

	buf := make([]byte, 2048)
	ftuple := conn.RemoteAddr().String() + "<->" + conn.LocalAddr().String()

	log.Printf("Established: %s\n", ftuple)

	for {
		length, err = conn.Read(buf)
		if err != nil {
			break
		}

		log.Printf("%s: %q (%d bytes)\n", ftuple, buf[:length], length)
	}

	if length == 0 {
		log.Printf("Closed: %s\n", ftuple)
	} else {
		log.Printf("Aborted: %s %s\n", ftuple, err.Error())
	}

	conn.Close()
}

func main() {
	var host string
	var port uint
	var logFile string
	var showVersion bool

	// Parse params from command-line arguments.
	flag.StringVar(&host, "H", "", "hostname to listen on.")
	flag.UintVar(&port, "p", 12345, "port number to listen on.")
	flag.StringVar(&logFile, "L", "", "log file.")
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

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatal("Failed to listen:", err.Error())
	}
	defer l.Close()

	log.Printf("Listen on %s:%d\n", host, port)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Failed to accept:", err.Error())
			continue
		}

		go handleRequest(conn)
	}
}
