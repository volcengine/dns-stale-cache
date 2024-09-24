package v7

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	. "github.com/volcengine/dns-stale-cache/common"
)

var defaultResolver *Resolver

// NewDialerWithCache returns a function that will be used as the default dialer
// when none is specified in Options.Dialer.
func NewDialerWithCache(opt *redis.Options, cacheOpts ...Option) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		var firstErr error
		var netDialer = &net.Dialer{
			Timeout:   opt.DialTimeout,
			KeepAlive: 5 * time.Minute,
		}
		var f = func(addr string) (net.Conn, error) {
			if opt.TLSConfig == nil {
				return netDialer.DialContext(ctx, network, addr)
			}
			return tls.DialWithDialer(netDialer, network, addr, opt.TLSConfig)
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

		return f(addr)
	}
}
