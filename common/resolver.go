/*
 * Copyright 2023 ByteDance and/or its affiliates.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	staleUpto     = 24 * time.Hour
	writeInterval = 3 * time.Second
)

type Resolver struct {
	context   context.Context
	staleUpTo time.Duration
	opt       *cacheOptions
	hostMap   sync.Map
	resolver  *net.Resolver
	goPool    gopool.Pool
}

type item struct {
	stored   int64
	info     []string
	hasWrite bool
}

func (r *Resolver) readFile() {
	dst := getCacheFilePath()

	dstFile, err := os.Open(dst)
	if err != nil {
		return
	}
	defer dstFile.Close()

	reader := bufio.NewReader(dstFile)
	for {
		line, err := reader.ReadString('\n')

		if err == io.EOF && line != "" {
			t, k, v := splitKeyValue(line[:len(line)-1], delimStr)
			if ttl(t, r.staleUpTo) < 0 {
				ipList := strings.Split(v, ",")
				r.setMap(k, *newItem(&ipList, true))
			}
			break
		}

		if err != nil {
			return
		}

		t, k, v := splitKeyValue(line[:len(line)-1], delimStr)
		if ttl(t, r.staleUpTo) < 0 {
			ipList := strings.Split(v, ",")
			r.setMap(k, *newItem(&ipList, true))
		}
	}
}

func (r *Resolver) cleanFileCache() error {
	r.hostMap.Range(func(key, value any) bool {
		if ttl(strconv.FormatInt(value.(item).stored, 10), r.staleUpTo) >= 0 {
			r.hostMap.Delete(key)
		}
		return true
	})

	dst := getCacheFilePath()
	err := os.Remove(dst)
	if err != nil {
		// todo
	}

	return nil
}

func (r *Resolver) writeFile() error {
	dst := getCacheFilePath()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil
	}
	defer dstFile.Close()

	w := bufio.NewWriter(dstFile)

	r.hostMap.Range(func(key, value any) bool {
		if value.(item).info == nil || value.(item).hasWrite {
			return true
		}

		content := strconv.FormatInt(value.(item).stored, 10) + delimStr + key.(string) + delimStr + strings.Join(value.(item).info, ",")
		r.setMap(key.(string), item{value.(item).stored, value.(item).info, true})

		_, err = fmt.Fprintln(w, content)
		if err != nil {
			return true
		}
		_ = w.Flush()

		return true
	})

	return nil
}

func (r *Resolver) start() {
	if r.opt.preferSaveIP == true {
		go createSchedule("write file", writeInterval, func(context.Context) error {
			return r.writeFile()
		})
	}

	go createSchedule("clean cache", r.staleUpTo, func(context.Context) error {
		return r.cleanFileCache()
	})
}

func NewResolver(opts ...Option) *Resolver {
	defaultOpts := defaultOption()
	for _, apply := range opts {
		apply(&defaultOpts)
	}

	r := &Resolver{
		context:   context.Background(),
		opt:       &defaultOpts,
		staleUpTo: staleUpto,
		hostMap:   sync.Map{},
		goPool:    gopool.NewPool("DnsHandler", int32(10*runtime.GOMAXPROCS(0)), gopool.NewConfig()),
	}

	r.readFile()
	r.start()

	return r
}

func newItem(info *[]string, hasWrite bool) *item {
	i := new(item)

	i.info = *info
	i.stored = time.Now().Unix()
	i.hasWrite = hasWrite

	return i
}

func (r *Resolver) setMap(k string, v item) {
	r.hostMap.Store(k, v)
}

func (r *Resolver) getMap(k string) []string {
	if v, ok := r.hostMap.Load(k); ok {
		return v.(item).info
	}

	return nil
}

func (r *Resolver) requestAndRefresh(addr string) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.opt.lookupTimeout)
	defer cancel()
	ipList, err := r.resolver.LookupHost(ctx, host)
	if err != nil {
		return
	}

	// Synthesize host:port，[host]:port
	for i, h := range ipList {
		ipList[i] = net.JoinHostPort(h, port)
	}

	if stringsEq(r.getMap(addr), ipList) {
		return
	}

	r.setMap(addr, *newItem(&ipList, false))
}

func (r *Resolver) LookupHost() ([]string, error) {
	var addrList []string

	for _, addr := range r.opt.addr {
		// perhaps url
		if strings.Contains(addr, "//") {
			u, err := url.Parse(addr)
			if err != nil {
				return nil, err
			}
			addr = u.Host
		}

		// prefer use cache
		if r.opt.preferUseCache {
			if addrs := r.getMap(addr); addrs != nil {
				r.goPool.Go(func() {
					r.requestAndRefresh(addr)
				})
				addrList = append(addrList, strings.Join(addrs, ","))
				continue
			}
		}

		r.requestAndRefresh(addr)
		tmp := r.getMap(addr)
		if tmp == nil {
			continue
		}
		addrList = append(addrList, strings.Join(tmp, ","))
	}

	return addrList, nil
}
