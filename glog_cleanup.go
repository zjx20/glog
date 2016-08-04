// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
// Copyright 2015 Michael Stapelberg. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// File cleanup for logs.

package glog

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

var logTotalBytes = flag.Int64("log_total_bytes",
	-1,
	"Restrict total size of log files created by this binary to this number of bytes")

// cleanup deletes old log files created by this binary (not restricted to the
// current pid), when their total size exceeds -log_total_bytes
func cleanup() {
	if *logTotalBytes == -1 {
		return
	}

	// Must match how logName() works.
	logRe, err := regexp.Compile(
		fmt.Sprintf("^%s\\.[^.]+\\.%s\\.log\\.(?:INFO|WARNING|ERROR|FATAL)\\.([^.]+)",
			regexp.QuoteMeta(program),
			regexp.QuoteMeta(userName)))
	if err != nil {
		// Print directly to stderr, as we cannot use log (might create a cycle).
		fmt.Fprintf(os.Stderr, "Cannot clean up old logfiles: %v\n", err)
		return
	}

	for _, dir := range logDirs {
		logdir, err := os.Open(dir)
		if err != nil {
			continue
		}
		defer logdir.Close()
		infos, err := logdir.Readdir(-1)
		if err != nil {
			continue
		}

		var total int64
		sorted := make([]string, 0, len(infos))
		sizes := make(map[string]int64)
		paths := make(map[string]string)
		for _, fi := range infos {
			name := fi.Name()
			matches := logRe.FindStringSubmatch(name)
			if matches == nil {
				continue
			}

			timestamp := matches[1]
			paths[timestamp] = name
			sorted = append(sorted, timestamp)
			sizes[timestamp] = fi.Size()
			total += fi.Size()
		}
		// Save the sort if there is nothing to clean up.
		if total < *logTotalBytes {
			continue
		}
		sort.Strings(sorted)

		for total > *logTotalBytes && len(sorted) > 0 {
			oldest := sorted[0]
			name := filepath.Join(dir, paths[sorted[0]])
			if err := os.Remove(name); err != nil {
				// Print directly to stderr, as we cannot use log (might create a cycle).
				fmt.Fprintf(os.Stderr, "Could not clean up old logfile %q: %v\n", name, err)
			}
			total -= sizes[oldest]
			sorted = sorted[1:]
		}
	}
}
