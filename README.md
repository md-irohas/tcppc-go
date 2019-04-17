# tcppc

`tcppc` is a simple honeypot program to capture TCP/TLS payloads.  This program
listens on the given IP address and the given port, establishes connections
from external hosts, and continues to receive packets until the connections are
closed or timeouted. Using `tcppc`, you can get payloads of arbitrary ports. I
am developing this program to use as a honeypot for monitoring payloads.


Main functions:

* Establish TCP & TLS handshake and continue to receive packets.
* Save received data (session data) as JSON lines format.
* Rotate the data files in the given interval.


## Installation

### Precompiled binary

Precompiled binaries for Linux (x86_64) are released.
See [release](https://github.com/md-irohas/tcppc-go/releases) page.

### Compile from source

`tcppc` is written in Go. So, if you want to build its binary, you need to
prepare the development environment for Go.

If you are ready for building Go, type the following commands.

```sh
$ go get github.com/md-irohas/tcppc-go
```

Note that the version of go compiler must be 1.11.0 or newer to enable timeout
of connection.


## Usage

The followings are the options of `tcppc`.
You can also use configuration files instead of using these options (See
'Configuration' section).

```sh
Usage of ./tcppc-go:
  -C string
        TLS certificate file.
  -H string
        hostname to listen on. (default "0.0.0.0")
  -K string
        TLS key file.
  -L string
        [deprecated] log file.
  -R uint
        maximum number of file descriptors (need root priviledge).
  -T int
        rotation interval [sec].
  -c string
        configuration file.
  -offset int
        rotation interval offset [sec].
  -p int
        port number to listen on. (default 12345)
  -t int
        timeout for TCP/TLS connection. (default 60)
  -v    show version and exit.
  -w string
        session file (JSON lines format).
  -z string
        timezone used for session file. (default "Local")
```


### Example-1: Basics

Run tcppc-go program.

```sh
$ sudo ./tcppc-go
2019/04/16 23:42:32 Maximum number of file descriptors: 1024
2019/04/16 23:42:32 Timezone: Local
2019/04/16 23:42:32 Timeout: 60
2019/04/16 23:42:32 Session data file: none.
2019/04/16 23:42:32 !!!CAUTION!!! Session data will not be written to files.
2019/04/16 23:42:32 Server Mode: TCP
2019/04/16 23:42:32 Listen: 0.0.0.0:12345
2019/04/16 23:42:32 Start TCP server.
2019/04/16 23:42:32 Server Mode: UDP
2019/04/16 23:42:32 Listen: 0.0.0.0:12345
2019/04/16 23:42:32 Start UDP server.
```

Connect to the server from another terminal.

```sh
$ echo "Hello, TCPPC" | nc 127.0.0.1 12345
```

The tcppc-go gets the following logs.

```sh
$ ./tcppc-go
...
2019/04/16 23:44:00 TCP: Established: Session: 2019-04-16T23:44:00: Flow: tcp 127.0.0.1:60998 <-> 127.0.0.1:12345 (0 payloads) (#Sessions: 1)
2019/04/16 23:44:00 TCP: Received: Session: 2019-04-16T23:44:00: Flow: tcp 127.0.0.1:60998 <-> 127.0.0.1:12345 (1 payloads): "Hello, TCPPC\n" (13 bytes)
2019/04/16 23:44:00 Closed: Session: 2019-04-16T23:44:00: Flow: tcp 127.0.0.1:60998 <-> 127.0.0.1:12345 (1 payloads) (#Sessions: 1)
```

Send UDP packets from the terminal.

```sh
$ echo "Hello, TCPPC" | nc -u 127.0.0.1 12345
```

The tcppc-go gets the following logs.

```sh
...
2019/04/16 23:45:20 UDP: Received: Session: 2019-04-16T23:45:20: Flow: udp 127.0.0.1:49616 <-> 127.0.0.1:12345 (1 payloads): "Hello, TCPPC\n" (13 bytes)
```

Type Ctrl+C to stop this program.

### Example-2: Save session data

You can save session data to files as [JSON lines](http://jsonlines.org/) format.

When `-w` option is specified, the data will be written to the given file.
You can use datetime format in `-w` option (See `man strftime` for more
details). When `-T` option is specified, data files will be rotated every
given seconds.

Run tcppc-go program.

```sh
$ ./tcppc-go -T 86400 -w log/tcppc-%Y%m%d.jsonl
```

Connect to the server from another terminal.

```sh
$ echo "Hello, TCPPC" | nc 127.0.0.1 12345
```

The results of the data are the following.
Note that data in payloads are encoded in base64.

```
# jq is a command for formatting JSON.
# I tested this on April 18th, 2018.

$ jq . log/tcppc-20180418.jsonl
{
  "timestamp": "2018-04-18T09:51:34.689896842+09:00",
  "flow": {
    "proto": "tcp",
    "src": "127.0.0.1",
    "sport": 53484,
    "dst": "127.0.0.1",
    "dport": 12345
  },
  "payloads": [
    {
      "index": 0,
      "timestamp": "2018-04-18T09:51:34.690076698+09:00",
      "data": "SGVsbG8sIFRDUFBDCg=="
    }
  ]
}
```

### Example-3: TLS handshaker

`tcppc` supports not only TCP handshake but also TLS handshake.

When both `-C` and `-K` options are specified, this program works as TLS
handshaker. You need to prepare for TLS certificate (in many cases,
self-signed) and key files (See 'Configuration' section).

Run tcppc-go program.

```sh
$ ./tcppc-go -T 86400 -C server.crt -K server.key -w log/tcppc-%Y%m%d.jsonl
```

Connect to the server from another terminal.

```sh
$ wget --no-check-certificate https://127.0.0.1:12345/index
```

The results of session data are the following (formatted by `jq` command).

```sh
$ jq . log/tcppc-20180418.jsonl
{
  "timestamp": "2018-04-18T10:06:08.104667676+09:00",
  "flow": {
    "proto": "tcp",
    "src": "127.0.0.1",
    "sport": 53635,
    "dst": "127.0.0.1",
    "dport": 12345
  },
  "payloads": [
    {
      "index": 0,
      "timestamp": "2018-04-18T10:06:08.111138967+09:00",
      "data": "R0VUIC9pbmRleCBIVFRQLzEuMQ0KVXNlci1BZ2VudDogV2dldC8xLjE5LjIgKGRhcndpbjE3LjMuMCkNCkFjY2VwdDogKi8qDQpBY2NlcHQtRW5jb2Rpbmc6IGd6aXANCkhvc3Q6IDEyNy4wLjAuMToxMjM0NQ0KQ29ubmVjdGlvbjogS2VlcC1BbGl2ZQ0KDQo="
    }
  ]
}

# decode "data" as base64.
$ echo "R0VUIC9pbmRleCBIVFRQLzEuMQ0KVXNlci1BZ2VudDogV2dldC8xLjE5LjIgKGRhcndpbjE3LjMuMCkNCkFjY2VwdDogKi8qDQpBY2NlcHQtRW5jb2Rpbmc6IGd6aXANCkhvc3Q6IDEyNy4wLjAuMToxMjM0NQ0KQ29ubmVjdGlvbjogS2VlcC1BbGl2ZQ0KDQo=" | base64 -D
GET /index HTTP/1.1
User-Agent: Wget/1.19.2 (darwin17.3.0)
Accept: */*
Accept-Encoding: gzip
Host: 127.0.0.1:12345
Connection: Keep-Alive

```


## Configuration

### Configuration file

The template of configuration file in [TOML](https://github.com/toml-lang/toml) format is ready.
See tcppc.toml.orig.

```sh
$ cp tcppc.toml.orig /etc/tcppc.toml
$ vim /etc/tcppc.toml

... (edit) ...
```

### TLS certificate/key files

If you want to use `tcppc` as TLS handshaker, you need to prepare TLS
certificate file and TLS key file.

You can create these files by the following commands.

```sh
$ openssl genrsa 2048 > server.key
$ openssl req -new -key server.key > server.csr
$ openssl x509 -days 36500 -req -signkey server.key < server.csr > server.crt
```

Note that these commands create not a valid certificate file but a
self-signed certificate file.

### Systemd

A simple unit file of systemd is ready (`tcppc.service.orig`)
Edit it and enable/start `tcppc` service.

```sh
# copy this file to systemd's directory.
cp -v tcppc.service.orig /etc/systemd/system/

# edit this file.
vim /etc/systemd/system/tcppc.service

# reload unit files.
systemctl daemon-reload

# start tcppc service.
systemctl start tcppc

# (optional) autostart tcppc service
systemctl enable tcppc
```

### Listen on all ports

The easiest way to listen on all ports is to use TPROXY function of `iptables`.

In this case, you should prepare a new (pseudo) network interface and an IP
address (i.e. IP alias) to monitor and capture all the traffic.

```sh
# !!! DANGER !!!
$ iptables -t mangle -A PREROUTING -i <interface> -p tcp -j TPROXY --tproxy-mark 0x1/0x1 --on-port 12345
$ iptables -t mangle -A PREROUTING -i <interface> -p udp -j TPROXY --tproxy-mark 0x1/0x1 --on-port 12345
```


## Session data format

Session data file is [JSON Lines](http://jsonlines.org/) format (i.e. each
line holds a JSON string.)
Each line represents each session data.

The following shows the data example with some comments.

```
{
  // Time when the session is accepted.
  // i.e.
  //   tcp/tls: time when the handshake is finished.
  //   udp: time when the UDP packet is received.
  "timestamp": "2018-04-18T11:06:09.419437117+09:00",

  // Flow (protocol, source IP address, source port, local address, local port)
  "flow": {
    "proto": "tcp",
    "src": "127.0.0.1",
    "sport": 54167,
    "dst": "127.0.0.1",
    "dport": 12345
  },

  // List of payloads
  "payloads": [
    {
      // Index of payloads
      "index": 0,

      // Time when this payload was received.
      "timestamp": "2018-04-18T11:06:13.830444868+09:00",

      // Data encoded in base64
      "data": "Rmlyc3QgcGF5bG9hZAo="
    },
    {
      "index": 1,
      "timestamp": "2018-04-18T11:06:18.015019663+09:00",
      "data": "U2Vjb25kIHBheWxvYWQK"
    }
  ]
}
```


## Alternatives

You might be able to use `nc` or `socat` instead.


## License

MIT License ([link](https://opensource.org/licenses/MIT)).


## Contact

md (E-mail: md.irohas at gmail.com)


