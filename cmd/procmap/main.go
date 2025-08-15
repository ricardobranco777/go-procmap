/* SPDX-License-Identifier: BSD-2-Clause */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	procmap "github.com/ricardobranco777/go-procmap"
)

func formatFlags(flags uint64) string {
	var s [4]byte
	if flags&procmap.VMAReadable != 0 {
		s[0] = 'r'
	} else {
		s[0] = '-'
	}
	if flags&procmap.VMAWritable != 0 {
		s[1] = 'w'
	} else {
		s[1] = '-'
	}
	if flags&procmap.VMAExecutable != 0 {
		s[2] = 'x'
	} else {
		s[2] = '-'
	}
	if flags&procmap.VMAShared != 0 {
		s[3] = 's'
	} else {
		s[3] = 'p'
	}
	return string(s[:])
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s PID\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	pid, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		log.Fatalf("Invalid PID: %v", err)
	}

	path := fmt.Sprintf("/proc/%d/maps", pid)
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open %s: %v", path, err)
	}
	defer file.Close()

	addr := uint64(0)
	for {
		q, name, buildID, err := procmap.Query(int(file.Fd()), addr, procmap.CoveringOrNextVMA|procmap.FileBackedVMA)
		if err != nil {
			break
		}

		flags := formatFlags(q.VMAFlags)
		offset := q.VMAOffset
		dev := fmt.Sprintf("%02x:%02x", q.DevMajor, q.DevMinor)
		inode := q.Inode

		fmt.Printf("%08x-%08x %s %08x %s %d",
			q.VMAStart, q.VMAEnd,
			flags,
			offset,
			dev,
			inode,
		)

		if name != "" {
			fmt.Printf("\t%s", name)
		}
		if len(buildID) > 0 {
			fmt.Printf("\t%x", buildID)
		}
		fmt.Println()

		addr = q.VMAEnd
	}
}
