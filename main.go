package main

import (
	"flag"
	"fmt"
	"github.com/md-irohas/tcppc-go/tcppc"
	"github.com/md-irohas/tcppc-go/crypto/tls"
	"github.com/pelletier/go-toml"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// Version number
	version = "0.2.1"
)


func main() {
	var host string
	var port uint
	var timeout int
	var tcpFileNameFmt string
	var rotInt int64
	var rotOffset int64
	var logFileName string
	var timezone string
	var maxFdNum uint64
	var x509Cert string
	var x509Key string
	var cnfFileName string
	var showVersion bool

	var listener net.Listener
	var writer *tcppc.RotWriter

	// Parse params from command-line arguments.
	flag.StringVar(&host, "H", "0.0.0.0", "hostname to listen on.")
	flag.UintVar(&port, "p", 12345, "port number to listen on.")
	flag.IntVar(&timeout, "t", 60, "timeout for TCP connection.")
	flag.StringVar(&tcpFileNameFmt, "w", "", "tcp session file (JSON lines format).")
	flag.Int64Var(&rotInt, "T", 0, "rotation interval [sec].")
	flag.Int64Var(&rotOffset, "offset", 0, "rotation interval offset [sec].")
	flag.StringVar(&logFileName, "L", "", "[deprecated] log file.")
	flag.StringVar(&timezone, "z", "Local", "timezone used for tcp session file.")
	flag.Uint64Var(&maxFdNum, "R", 0, "set maximum number of file descriptors (need root priviledge in some environments).")
	flag.StringVar(&x509Cert, "C", "", "TLS certificate file.")
	flag.StringVar(&x509Key, "K", "", "TLS key file.")
	flag.StringVar(&cnfFileName, "c", "", "configuration file.")
	flag.BoolVar(&showVersion, "v", false, "show version and exit.")
	flag.Parse()

	if showVersion {
		fmt.Println("version:", version)
		return
	}

	// Parse params from config file. Params in the command-line arguments are
	// ignored.
	if cnfFileName != "" {
		cnf, err := toml.LoadFile(cnfFileName)
		if err != nil {
			log.Fatal(err)
		}

		host = cnf.Get("tcppc.host").(string)
		port = uint(cnf.Get("tcppc.port").(int64))
		timeout = int(cnf.Get("tcppc.timeout").(int64))
		tcpFileNameFmt = cnf.Get("tcppc.tcpFileFmt").(string)
		rotInt = cnf.Get("tcppc.rotInt").(int64)
		rotOffset = cnf.Get("tcppc.rotOffset").(int64)
		logFileName = cnf.Get("tcppc.logFile").(string)
		timezone = cnf.Get("tcppc.timezone").(string)
		maxFdNum = uint64(cnf.Get("tcppc.maxFdNum").(int64))
		x509Cert = cnf.Get("tcppc.x509Cert").(string)
		x509Key = cnf.Get("tcppc.x509Key").(string)
	}

	log.SetFlags(log.LstdFlags)

	// This log file is deprecated. The default logging package of goland does
	// not support a function to rotate log files. Therefore, the log file of
	// this process will consume huge diskspace. If your OS uses
	// systemd-journald, it manages the stdout/stderr of this process, so you
	// should use it instead.
	if logFileName != "" {
		f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
		if err != nil {
			log.Fatalf("Failed to open log file: %s\n", err)
		}

		log.SetOutput(f)
		log.Printf("Open log file: %s\n", logFileName)
	}

	log.Printf("Start TCPPC program.\n")

	// Raise the upper limit of the number of file descriptors to handle many
	// requests such as port scannings by attackers.
	var rLimit syscall.Rlimit
	if maxFdNum > 0 {
		rLimit.Max = maxFdNum
		rLimit.Cur = maxFdNum

		err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			log.Fatalf("Failed to set maximum number of file descriptors: %s\n", err)
		}
	}

	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Fatalf("Failed to get maximum number of file descriptos.\n")
	}

	log.Printf("Maximum number of file descriptors: %d\n", rLimit.Cur)

	// Load location from timezone. This location object is used to determine
	// the filename of tcp session files by RotWriter.
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Fatalf("Failed to load timezone: %s %s\n", timezone, err)
	}

	log.Printf("Timezone: %s\n", timezone)
	log.Printf("Timeout: %d\n", timeout)

	// Init signal handling.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-sigc

		log.Printf("SIGNAL: %s\n", s)
		if writer != nil {
			writer.Close()
		}
		os.Exit(0)
	}()

	// Prepare for TCP/TLL handshake listener.  When both TLS certificate file
	// and TLS key file are given, start listening as TLS handshaker. When none
	// of them are given, start listening as TCP handshaker. Otherwise, failed
	// to start listening.

	if x509Cert != "" && x509Key != "" {
		log.Printf("Server Mode: TLS handshaker.\n")
		log.Printf("Certificate: %s, Key: %s\n", x509Cert, x509Key)

		cer, err := tls.LoadX509KeyPair(x509Cert, x509Key)
		if err != nil {
			log.Fatalf("Failed to load X509 key pair: %s\n", err)
		}

		config := &tls.Config{
			Certificates: []tls.Certificate{cer},
		}

		log.Printf("Listen: %s:%d\n", host, port)

		listener, err = tls.Listen("tcp", fmt.Sprintf("%s:%d", host, port), config)
		if err != nil {
			log.Fatalf("Failed to listen: %s\n", err)
		}
	} else if x509Cert == "" && x509Key == "" {
		log.Printf("Server Mode: TCP handshaker.\n")
		log.Printf("Listen: %s:%d\n", host, port)

		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			log.Fatalf("Failed to listen: %s\n", err)
		}
	} else {
		log.Println("Either TLS cerfiticate or key file is given.")
		log.Println("TCP handshaker: neither TLS certificate nor TLS key files are required.")
		log.Println("TLS handshaker: both TLS certificate and TLS key files are required.")
		log.Fatalln("Exit.")
	}

	defer listener.Close()

	if tcpFileNameFmt != "" {
		log.Printf("TCP session data: %s (Rotate every %d seconds w/ %d seconds offset)", tcpFileNameFmt, rotInt, rotOffset)
		writer = tcppc.NewWriter(tcpFileNameFmt, rotInt, rotOffset, loc)
	} else {
		log.Printf("TCP session data: none.\n")
		log.Printf("!!!CAUTION!!! TCP session data will not be written to files.\n")
		writer = nil
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Failed to accept a new connection: %s\n", err)
		}

		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		go tcppc.HandleRequest(conn, writer)
	}
}
