Files in crypto directory are copied or modifed version of the original
golang repository. See LICENSE file for more details.


Original Repository:
  https://github.com/golang/go


Differences between original source codes and modified source codes are as
follows:

crypto/cipher_suites.go

```
18,19c18
< 	// "golang_org/x/crypto/chacha20poly1305"
< 	"golang.org/x/crypto/chacha20poly1305"
---
> 	"golang_org/x/crypto/chacha20poly1305"
```

crypto/common.go

```
10c10
< 	// "crypto/internal/cipherhw"
---
> 	"crypto/internal/cipherhw"
22,24d21
< 
< 	// internal package is not allowed to import from external package.
< 	"github.com/md-irohas/tcppc-go/crypto/internal/cipherhw"
```

conn.go

```
17d16
< 	"os"	// appended by mkt
1357,1369d1355
< 
< 
< // The following method is appended by mkt to provide the os.File object of tls.Conn.
< // File descriptors of os.File objects are required to call getsockopt systemc call.
< 
< func (c *Conn) File() (*os.File, error) {
< 	// Assume that Conn is an instance of net.TCPConn.
< 	if tcpConn, ok := c.conn.(*net.TCPConn); ok {
< 		return tcpConn.File()
< 	} else {
< 		return nil, errors.New("TLS is not over TCP.")
< 	}
< }
```

key_agreement.go

```
20,21c20
< 	// "golang_org/x/crypto/curve25519"
< 	"golang.org/x/crypto/curve25519"
---
> 	"golang_org/x/crypto/curve25519"
```
