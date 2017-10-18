# TCPPC-GO

This is a simple program to capture TCP payloads. This program listens on the
given IP address and the given port, establishes connections from external
hosts, and continues to receive packets until the connections are closed.

I wrote this program to use as a honeypot for monitoring the payload of the
arbitrary ports. This may be "Re-inventing the wheel."
If you know better utilities, please let me know.


## Installation

### Precompiled binary

A precompiled binary for Linux is released. See release page.


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
  -p uint
    	port number to listen on. (default 12345)
  -v	show version and exit.
```

Note: The log file is not intended to be used for analysis.
It is just for the check of the program.
If you want to analyze packets, you should capture the packets by capturing
utilities such as `tcpdump` and `tshark` (I also wrote a packet-capturing
utility named `rcap-go` ([link](https://github.com/md-irohas/rcap-gp)) for the
purpose).


### Listen on all ports?

Redirect all packets to the listening port by `iptables`.

If you want to redirect all packets, you should prepare a new network interface
and an IP address to capture the whole traffic.

```
# !!! DANGER !!!
$ iptables -t nat -A PREROUTING -i <interface> -p tcp -d <listen-ip> -j DNAT --to-destination <listen-ip>:<listen-port>
```


## Alternatives

You can use `nc` or `socat` instead.


## License

MIT License ([link](https://opensource.org/licenses/MIT)).


## Contact

md (E-mail: md.irohas at gmail.com)


