# TCPPC-GO

This is a simple program to capture TCP payloads. This program listens on the
given IP address and the given port, establishes connections from external
hosts, and continues to receive packets until the connections are closed.

I wrote this program to use as a honeypot for monitoring payloads of arbitrary
ports. This may be "Re-inventing the wheel." If you know better utilities,
please let me know.


## Installation

### Precompiled binary

Precompiled binaries for Linux (x86_64) is released. See release page.


### Compile from source

```sh
$ git clone https://github.com/md-irohas/tcppc-go
$ go build tcppc.go
```


### Systemd

I prepared a simple unit file of systemd (see `tcppc.service.orig`).
Edit the file and enable/start tcppc service.

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


## Usage

```
Usage of ./tcppc:
  -H string
    	hostname to listen on.
  -L string
    	log file.
  -R uint
    	set maximum number of file descriptors (need root priviledge in some environments).
  -p uint
    	port number to listen on. (default 12345)
  -t int
    	timeout for TCP connection. (default 60)
  -v	show version and exit.
```

Note:
The log file is not intended to be used for analysis.
It is just for the check of the program.
If you want to analyze packets, you should capture the packets by capturing
utilities such as `tcpdump` and `tshark` (I also wrote a packet-capturing
utility named `rcap-go` ([link](https://github.com/md-irohas/rcap-go)) for the
purpose).


### Listen on all ports?

The easiest way to listen on all ports is to redirect all packets to the
listening port by packet forwarding tools such as `iptables`.

In this case, you should prepare a new (pseudo) network interface and an IP
address (i.e. IP alias) to monitor and capture all the traffic.

```
# !!! DANGER !!!
$ iptables -t nat -A PREROUTING -i <interface> -p tcp -d <listen-ip> -j DNAT --to-destination <listen-ip>:<listen-port>
```

Note that if you redirect all the traffic, you cannot verify destination port
numbers from logs of `tcppc` (Therefore, you should use a packet capture
utility for logging.)


## Alternatives

You can use `nc` or `socat` instead.


## License

MIT License ([link](https://opensource.org/licenses/MIT)).


## Contact

md (E-mail: md.irohas at gmail.com)


