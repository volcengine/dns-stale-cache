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

package rocketmq

import (
	"fmt"
	"os"
	"strings"

	. "github.com/volcengine/dns-stale-cache/common"
)

type CacheResolver struct {
	addr []string
}

var resolver *Resolver

func NewCacheResolver(addr []string, opts ...Option) *CacheResolver {
	opts = append(opts, WithAddr(addr))
	resolver = NewResolver(opts...)

	return &CacheResolver{
		addr: addr,
	}
}

func (r *CacheResolver) defaultResolve() []string {
	if v := os.Getenv("NAMESRV_ADDR"); v != "" {
		return strings.Split(v, ";")
	}
	return nil
}

func (r *CacheResolver) Resolve() []string {
	addrList, err := resolver.LookupHost()
	if err != nil {
		return nil
	}

	// Take only one IP address
	for i, ip := range addrList {
		addrList[i] = strings.Split(ip, ",")[0]
	}

	if len(addrList) > 0 {
		return addrList
	}

	return r.defaultResolve()
}

func (r *CacheResolver) Description() string {
	return fmt.Sprintf("resolver with cache of domain:%v ", r.addr)
}
