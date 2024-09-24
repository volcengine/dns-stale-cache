package v6

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"time"

	"github.com/go-redis/redis"
	. "github.com/volcengine/dns-stale-cache/common"
)

var defaultResolver *Resolver

// NewDialerWithCache returns a function that will be used as the default dialer
// when none is specified in Options.Dialer.
func NewDialerWithCache(opt *redis.Options, cacheOpts ...Option) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		var firstErr error
		var netDialer = &net.Dialer{
			Timeout:   opt.DialTimeout,
			KeepAlive: 5 * time.Minute,
		}
		var f = func(addr string) (net.Conn, error) {
			if opt.TLSConfig == nil {
				return netDialer.DialContext(context.Background(), opt.Network, addr)
			}
			return tls.DialWithDialer(netDialer, opt.Network, addr, opt.TLSConfig)
		}

		if defaultResolver == nil {
			cacheOpts = append(cacheOpts, WithAddr([]string{opt.Addr}))
			defaultResolver = NewResolver(cacheOpts...)
		}
		addrList, err := defaultResolver.LookupHost()
		if err == nil && len(addrList) == 1 && addrList[0] != "" {
			addrs := strings.Split(addrList[0], ",")

			for _, ip := range addrs {
				c, ret := f(ip)
				if ret == nil {
					return c, nil
				}

				if firstErr == nil {
					firstErr = ret
				}
			}

			// Attempt to connect to all IPs failed
			return nil, firstErr
		}

		return f(opt.Addr)
	}
}
