package mikrotik

import (
	"crypto/tls"
	"time"
)

//Client
type Client interface {
	DialTLSTimeout(address, username, password string, tlsConfig *tls.Config, timeout time.Duration) (*Client, error)
	Close()
}
