package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/md-irohas/tcppc-go/tcppc"
	"github.com/pelletier/go-toml"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const (
	Version = "0.3.0"
)

var (
	host             = flag.String("H", "0.0.0.0", "hostname to listen on.")
	port             = flag.Int("p", 12345, "port number to listen on.")
	timeout          = flag.Int("t", 60, "timeout for TCP/TLS connection.")
	fileNameFmt      = flag.String("w", "", "session file (JSON lines format).")
	rotInt           = flag.Int("T", 0, "rotation interval [sec].")
	rotOffset        = flag.Int("offset", 0, "rotation interval offset [sec].")
	logFileName      = flag.String("L", "", "[deprecated] log file.")
	timezone         = flag.String("z", "Local", "timezone used for session file.")
	maxFdNum         = flag.Uint64("R", 0, "maximum number of file descriptors (need root priviledge).")
	x509Cert         = flag.String("C", "", "TLS certificate file.")
	x509Key          = flag.String("K", "", "TLS key file.")
	cnfFileName      = flag.String("c", "", "configuration file.")
	disableTcpServer = flag.Bool("disable-tcp-server", false, "disable TCP/TLS server.")
	disableUdpServer = flag.Bool("disable-udp-server", false, "disable UDP server.")
	showVersion      = flag.Bool("v", false, "show version and exit.")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Println(Version)
		return
	}

	log.SetFlags(log.LstdFlags)

	// Linux only.
	if runtime.GOOS != "linux" {
		log.Fatalf("This program runs only in Linux.")
	}

	// Parse params from config file.
	// Params in the command-line arguments are ignored.
	if *cnfFileName != "" {
		cnf, err := toml.LoadFile(*cnfFileName)
		if err != nil {
			log.Fatalf("Failed to load configuration file: %s", err)
		}

		*host = cnf.Get("tcppc.host").(string)
		*port = int(cnf.Get("tcppc.port").(int64))
		*timeout = int(cnf.Get("tcppc.timeout").(int64))
		*fileNameFmt = cnf.Get("tcppc.tcpFileFmt").(string)
		*rotInt = int(cnf.Get("tcppc.rotInt").(int64))
		*rotOffset = int(cnf.Get("tcppc.rotOffset").(int64))
		*logFileName = cnf.Get("tcppc.logFile").(string)
		*timezone = cnf.Get("tcppc.timezone").(string)
		*maxFdNum = uint64(cnf.Get("tcppc.maxFdNum").(int64))
		*x509Cert = cnf.Get("tcppc.x509Cert").(string)
		*x509Key = cnf.Get("tcppc.x509Key").(string)
	}

	// This log file is deprecated.
	// The default logging package of goland does not support a function to
	// rotate log files. Therefore, the log file of this process will consume
	// huge diskspace. If your OS uses systemd-journald, it manages the
	// stdout/stderr of this process, so you should use it instead.
	if *logFileName != "" {
		f, err := os.OpenFile(*logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
		if err != nil {
			log.Fatalf("Failed to open log file: %s\n", err)
		}

		log.SetOutput(f)
		log.Printf("Open log file: %s\n", logFileName)
	}

	// Raise the upper limit of the number of file descriptors to handle many
	// requests such as port scannings by attackers.
	var rLimit syscall.Rlimit
	if *maxFdNum > 0 {
		rLimit.Max = *maxFdNum
		rLimit.Cur = *maxFdNum

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

	// Load location from timezone.
	// This location object is used to determine the filename of tcp session
	// files by RotWriter.
	loc, err := time.LoadLocation(*timezone)
	if err != nil {
		log.Fatalf("Failed to load timezone: %s %s\n", *timezone, err)
	}

	log.Printf("Timezone: %s\n", *timezone)
	log.Printf("Timeout: %d\n", *timeout)

	// Select mode of tcppc.
	// When both TLS certificate file and TLS key file are given, this program
	// starts listening as TLS handshaker. When none of them are given, this
	// program starts listening as TCP handshaker. Otherwise, this program
	// fails to start listening.
	var tcppcMode string
	if !*disableTcpServer {
		if *x509Cert != "" && *x509Key != "" {
			tcppcMode = "tls"
		} else if *x509Cert == "" && *x509Key == "" {
			tcppcMode = "tcp"
		} else {
			log.Println("Either TLS cerfiticate or key file is given.")
			log.Println("TCP handshaker: neither TLS certificate nor TLS key files are required.")
			log.Println("TLS handshaker: both TLS certificate and TLS key files are required.")
			log.Fatalln("Abort.")
		}
	}

	var writer *tcppc.RotWriter
	if *fileNameFmt != "" {
		log.Printf("Session data file: %s (Rotate every %d seconds w/ %d seconds offset)\n", *fileNameFmt, *rotInt, *rotOffset)

		writer = tcppc.NewWriter(*fileNameFmt, *rotInt, *rotOffset, loc)
		defer writer.Close()
	} else {
		log.Printf("Session data file: none.\n")
		log.Printf("!!!CAUTION!!! Session data will not be written to files.\n")

		writer = nil
	}

	if !*disableTcpServer {
		switch tcppcMode {
		case "tcp":
			go tcppc.StartTCPServer(*host, *port, writer, *timeout)

		case "tls":
			log.Printf("Certificate: %s, Key: %s\n", *x509Cert, *x509Key)

			cer, err := tls.LoadX509KeyPair(*x509Cert, *x509Key)
			if err != nil {
				log.Fatalf("Failed to load X509 key pair: %s\n", err)
			}

			config := &tls.Config{
				Certificates: []tls.Certificate{cer},
			}

			go tcppc.StartTLSServer(*host, *port, config, writer, *timeout)
		default:
			log.Fatalf("Unknown mode of tcppc: %s\n", tcppcMode)
		}
	}

	// Wait for TCP/TLS server to start, then start UDP server.
	if !*disableUdpServer {
		time.Sleep(100 * time.Millisecond)
		go tcppc.StartUDPServer(*host, *port, writer)
	}

	// Wait for SIGNAL.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigc:
		log.Printf("Exit.")
	}
}
