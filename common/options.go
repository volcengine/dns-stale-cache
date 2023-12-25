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

import "time"

type cacheOptions struct {
	addr           []string
	lookupTimeout  time.Duration
	preferUseCache bool
	preferSaveIP   bool
}

func defaultOption() cacheOptions {
	opts := cacheOptions{
		preferUseCache: false,
		preferSaveIP:   false,
		lookupTimeout:  1 * time.Second,
	}
	return opts
}

type Option func(*cacheOptions)

// WithCacheFirst If false, initiate DNS requests first each time and then flush the cache;
// If true, return the cache content first each time and asynchronously initiate DNS requests to refresh the cache.
// Default is false.
func WithCacheFirst(preferUse bool) Option {
	return func(options *cacheOptions) {
		options.preferUseCache = preferUse
	}
}

// WithIPConsistance Persist cached content or not.
// Default is false.
func WithIPConsistance(preferUse bool) Option {
	return func(options *cacheOptions) {
		options.preferSaveIP = preferUse
	}
}

// WithDnsTimeout Lookup timeout for resolving domain names.
// Default is 1 second.
func WithDnsTimeout(timeout time.Duration) Option {
	return func(options *cacheOptions) {
		options.lookupTimeout = timeout
	}
}

// WithAddr domain or host:port address.
func WithAddr(addr []string) Option {
	return func(options *cacheOptions) {
		options.addr = addr
	}
}
