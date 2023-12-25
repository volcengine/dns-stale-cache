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

package v7

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
	. "github.com/volcengine/dns-stale-cache/common"
)

func ExampleClient() {
	opt := &redis.Options{
		Addr: "localhost:6379",
	}
	opt.Dialer = NewDialerWithCache(opt,
		WithCacheFirst(true),
		WithIpStorageFirst(true),
		WithDnsTimeout(2*time.Second),
	)

	rdb := redis.NewClient(opt)

	err := rdb.Set("key1", "value1", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := rdb.Get("key1").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := rdb.Get("key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	// Output: key value
	// key2 does not exist
}
