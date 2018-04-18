# tcpPC

`tcpPC` is a simple program to capture TCP payloads.  This program listens
on the given IP address and the given port, establishes connections from
external hosts, and continues to receive packets until the connections are
closed or timeouted.  This program supports not only TCP handshake but also
TLS(SSL) handshake.  Using `tcpPC`, you can get payloads of arbitrary ports.
I am developing this program to use as a honeypot for monitoring payloads of
arbitrary ports.


The followings are the main functions of `tcpPC`

* Establish TCP & SSL handshake and continue to receive packets.
* Save received data (session data) as JSON lines format.
* Rotate the data files in the given interval.


## Installation

### Precompiled binary

Precompiled binaries for Linux (x86_64) is released. See release page.

### Compile from source

`tcpPC` is written in Go. So, if you want to build its binary, you need to
prepare the development environment for Go.

If you have got ready for building Go, type the following commands.

```sh
$ git clone https://github.com/md-irohas/tcppc-go

$ cd path/to/tcppc-go

$ go get
$ go build
```

### Move the binary

Move the binary to `/usr/local/bin`.

```
$ cp -v tcppc-go /usr/local/bin/tcppc
```


## Usage

The followings are the options of `tcpPC`. You can also use configuration
files instead of using these options (See 'Configuration' section).

```
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
    	set maximum number of file descriptors (need root priviledge in some environments).
  -T int
    	rotation interval [sec]. (default 0)
  -c string
    	configuration file.
  -offset int
    	rotation interval offset [sec].
  -p uint
    	port number to listen on. (default 12345)
  -t int
    	timeout for TCP connection. (default 60)
  -v	show version and exit.
  -w string
    	tcp session file (JSON lines format).
  -z string
    	timezone used for tcp session file. (default "Local")
```


I will show you three basic usages of `tcpPC`.


### Example-1: Basics

Run tcppc-go program.

```sh
$ ./tcppc-go
2018/04/18 09:08:51 Start TCPPC program.
2018/04/18 09:08:51 Maximum number of file descriptors: 256
2018/04/18 09:08:51 Timezone: Local
2018/04/18 09:08:51 Server Mode: TCP handshaker.
2018/04/18 09:08:51 Listen: 0.0.0.0:12345
2018/04/18 09:08:51 TCP session data: none.
2018/04/18 09:08:51 !!!CAUTION!!! TCP session data will not be written to files.
```

Connect to the server from another terminal.

```sh
$ echo "Hello, TCPPC" | nc 127.0.0.1 12345
```

The tcppc-go will get the following logs.

```sh
$ ./tcppc-go
...
2018/04/18 09:10:13 Established: TCPSession: 2018-04-18T09:10:13: TCPFlow: 127.0.0.1:53073 <-> 127.0.0.1:12345 (0 payloads) (#Sessions: 1)
2018/04/18 09:10:13 Received: TCPSession: 2018-04-18T09:10:13: TCPFlow: 127.0.0.1:53073 <-> 127.0.0.1:12345 (1 payloads): "Hello, TCPPC\n" (13 bytes)
2018/04/18 09:10:13 Closed: TCPSession: 2018-04-18T09:10:13: TCPFlow: 127.0.0.1:53073 <-> 127.0.0.1:12345 (1 payloads) (#Sessions: 1)
```

Type Ctrl+C to stop this program.

### Example-2: Save session data

You can save session data to files as [JSON lines](http://jsonlines.org/) format.

When `-w` option is specified, the data will be written to the given file.
You can use datetime format in `-w` option (See `man strftime` for more
details).

When `-T` option is specified, data files will be rotated every given seconds.

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

`tcpPC` supports not only TCP handshake but also TLS handshake.

When both `-C` and `-K` options are specified, this program works as TLS handshaker.
You need to prepare for TLS certificate (in many cases, self-signed) and key
files (See 'Configuration' section).

```sh
$ ./tcppc-go -T 86400 -C server.crt -K server.key -w log/tcppc-%Y%m%d.jsonl
```

Connect to the server from another terminal.

```sh
$ wget --no-check-certificate https://127.0.0.1:12345/index
```

The results of session data are the following (formatted by `jq` command).

```
$ jq . log/tcppc-20180418.jsonl
{
  "timestamp": "2018-04-18T10:06:08.104667676+09:00",
  "flow": {
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

```
$ cp tcppc.toml.orig /etc/tcppc.toml
$ vim /etc/tcppc.toml

... (edit) ...
```


### TLS certificate/key files

If you want to use `tcpPC` as TLS handshaker, you need to prepare TLS
certificate file and TLS key file.

You can create these files by the following commands.

```
$ openssl genrsa 2048 > server.key
$ openssl req -new -key server.key > server.csr
$ openssl x509 -days 36500 -req -signkey server.key < server.csr > server.crt
```

Note that these commands create not a valid certificate file but a
self-signed certificate file.


### Systemd

A simple unit file of systemd is ready (`tcppc.service.orig`)
Edit it and enable/start `tcpPC` service.

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

The easiest way to listen on all ports is to redirect all packets to the
listening port by a packet forwarding tool such as `iptables`.

In this case, you should prepare a new (pseudo) network interface and an IP
address (i.e. IP alias) to monitor and capture all the traffic.

```
# !!! DANGER !!!
$ iptables -t nat -A PREROUTING -i <interface> -p tcp -d <listen-ip> -j DNAT --to-destination <listen-ip>:<listen-port>
```

Although the destination address and port are converted by NAT, recent linux
provides an `ORIGINAL_DST` function to lookup the original destination
address and port. Therefore, `tcpPC` can record original destination address
and port in linux environment.


## Session data format

Session data file is [JSON Lines](http://jsonlines.org/) format (i.e. each
line holds a JSON string.)
Each line represents each session data.

The following shows the data example with some comments.

```
{
  // Time when the session is accepted (i.e. the time when SYN packet was received).
  "timestamp": "2018-04-18T11:06:09.419437117+09:00",

  // TCP flow (source IP address, source port, local address, local port)
  "flow": {
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


