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
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"
)

const ipFileName = "ip_info"
const delimStr = "->"

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

// 返回当前进程的home路径
func getExec() string {
	u, err := user.Current()
	if nil == err {
		return u.HomeDir
	}

	// Unix-like system, so just assume Unix
	if p, err := homeUnix(); err != nil {
		return p
	}

	return "~"
}

func getCacheFilePath() string {
	filePath := getExec()
	dst := filePath + "/" + ipFileName

	return dst
}

// return: [time, key, value]
func splitKeyValue(line, delim string) (string, string, string) {
	ret := strings.Split(line, delim)
	if len(ret) == 3 {
		return ret[0], ret[1], ret[2]
	}

	return "", "", ""
}

type scheduleFunc func(context.Context) error

func random(base int64, bias float64) int64 {
	return int64(((rand.Float64()-bias)*bias + 1) * float64(base))
}

func createSchedule(name string, interval time.Duration, f scheduleFunc) {
	for {
		base := interval.Nanoseconds()
		nanoseconds := random(base, 0.5) // 0.75 - 1.25 interval

		timer := time.NewTimer(time.Duration(nanoseconds))
		<-timer.C

		ctx := context.TODO()
		if err := f(ctx); err != nil {
			// todo
			fmt.Printf("err %v\n", err)
		}
	}
}

func stringsEq(a, b []string) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// return: (now.time - stored.time) - duration.time
func ttl(stored string, d time.Duration) int {
	t, _ := strconv.ParseInt(stored, 10, 64)
	return int(time.Now().Unix() - t - int64(d.Seconds()))
}
