/* SPDX-License-Identifier: BSD-2-Clause */

package procmap

/*
#include <linux/fs.h>
*/
import "C"

import (
	"fmt"
	"golang.org/x/sys/unix"
	"unsafe"
)

const (
	VMAReadable       = 0x01
	VMAWritable       = 0x02
	VMAExecutable     = 0x04
	VMAShared         = 0x08
	CoveringOrNextVMA = 0x10
	FileBackedVMA     = 0x20

	MaxBuildIDLen    = 64
	MaxVMANamelength = 512
)

var procmapQuery uint

func init() {
	procmapQuery = uint(C.PROCMAP_QUERY)
}

// ProcmapQuery mirrors struct procmap_query from linux/fs.h
type ProcmapQuery struct {
	Size        uint64 // Must be set to sizeof(ProcmapQuery)
	QueryFlags  uint64 // in
	QueryAddr   uint64 // in
	VMAStart    uint64 // out
	VMAEnd      uint64 // out
	VMAFlags    uint64 // out
	VMAPageSize uint64 // out
	VMAOffset   uint64 // out
	Inode       uint64 // out
	DevMajor    uint32 // out
	DevMinor    uint32 // out
	VMANameSize uint32 // in/out
	BuildIDSize uint32 // in/out
	VMANameAddr uint64 // in
	BuildIDAddr uint64 // in
}

// Query performs the PROCMAP_QUERY ioctl on the given file descriptor and address.
func Query(fd int, addr uint64, flags uint64) (*ProcmapQuery, string, []byte, error) {
	var nameBuf [MaxVMANamelength]byte
	var buildID [MaxBuildIDLen]byte

	q := &ProcmapQuery{
		Size:        uint64(unsafe.Sizeof(ProcmapQuery{})),
		QueryFlags:  flags,
		QueryAddr:   addr,
		VMANameSize: MaxVMANamelength,
		BuildIDSize: MaxBuildIDLen,
		VMANameAddr: uint64(uintptr(unsafe.Pointer(&nameBuf[0]))),
		BuildIDAddr: uint64(uintptr(unsafe.Pointer(&buildID[0]))),
	}

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(procmapQuery),
		uintptr(unsafe.Pointer(q)),
	)
	if errno != 0 {
		return nil, "", nil, fmt.Errorf("ioctl PROCMAP_QUERY: %w", errno)
	}

	vmaName := ""
	if q.VMANameSize > 0 {
		vmaName = string(nameBuf[:q.VMANameSize])
	}
	buildIDCopy := make([]byte, q.BuildIDSize)
	copy(buildIDCopy, buildID[:q.BuildIDSize])

	return q, vmaName, buildIDCopy, nil
}
