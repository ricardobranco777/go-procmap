/* SPDX-License-Identifier: BSD-2-Clause */

package procmap

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

func TestQuery(t *testing.T) {
	const pageSize = 4096
	const prot = syscall.PROT_READ | syscall.PROT_EXEC
	const flags = syscall.MAP_PRIVATE | syscall.MAP_ANONYMOUS

	b, err := syscall.Mmap(-1, 0, pageSize, prot, flags)
	if err != nil {
		t.Fatalf("mmap failed: %v", err)
	}
	defer syscall.Munmap(b)

	addr := uintptr(unsafe.Pointer(&b[0]))
	fd, err := os.Open(fmt.Sprintf("/proc/%d/maps", os.Getpid()))
	if err != nil {
		t.Fatalf("open /proc/self/maps failed: %v", err)
	}
	defer fd.Close()

	t.Run("Exact match", func(t *testing.T) {
		q, name, buildID, err := Query(int(fd.Fd()), uint64(addr), 0)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if q.VMAStart != uint64(addr) || q.VMAEnd != uint64(addr+pageSize) {
			t.Errorf("VMA range mismatch: start=0x%x end=0x%x", q.VMAStart, q.VMAEnd)
		}
		if q.VMAPageSize != pageSize {
			t.Errorf("page size mismatch: got %d, want %d", q.VMAPageSize, pageSize)
		}
		if name != "" {
			t.Errorf("expected empty VMA name for anonymous mapping, got %q", name)
		}
		if buildID == nil {
			t.Errorf("expected build ID, got nil")
		}
	})

	t.Run("Address - 1 should not match", func(t *testing.T) {
		_, _, _, err := Query(int(fd.Fd()), uint64(addr-1), 0)
		if err == nil {
			t.Errorf("expected error for addr-1, got nil")
		}
	})

	t.Run("COVERING_OR_NEXT_VMA from addr - 1", func(t *testing.T) {
		q, name, _, err := Query(int(fd.Fd()), uint64(addr-1), CoveringOrNextVMA)
		name = strings.TrimRight(name, "\x00")
		if err != nil {
			if name == "[vsyscall]" {
				t.Skipf("Skipping [vsyscall] region at %#x", q.VMAStart)
			}
			t.Errorf("COVERING_OR_NEXT_VMA failed: %v", err)
			return
		}
		if name == "[vsyscall]" {
			t.Skipf("Skipping [vsyscall] region at %#x", q.VMAStart)
		}
		if q.VMAStart != uint64(addr) {
			t.Errorf("expected VMA start 0x%x, got 0x%x", addr, q.VMAStart)
		}
	})

	t.Run("Writable VMA mismatch", func(t *testing.T) {
		_, _, _, err := Query(int(fd.Fd()), uint64(addr), VMAWritable)
		if err == nil {
			t.Errorf("expected error for non-writable VMA, got nil")
		}
	})
}

func parseHexUint64(s string) (uint64, error) {
	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid hex: %q", s)
	}
	return v, nil
}

func TestProcmapAgainstProcMaps(t *testing.T) {
	pid := os.Getpid()
	path := fmt.Sprintf("/proc/%d/maps", pid)

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open maps file: %v", err)
	}
	defer file.Close()

	fd := int(file.Fd())

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 5 {
			t.Errorf("unexpected line format: %q", line)
			continue
		}

		addrs := strings.Split(fields[0], "-")
		start, err := parseHexUint64(addrs[0])
		if err != nil {
			t.Errorf("invalid start address: %v", err)
			continue
		}
		end, err := parseHexUint64(addrs[1])
		if err != nil {
			t.Errorf("invalid end address: %v", err)
			continue
		}

		// Skip [vsyscall] mapping entirely
		if len(fields) >= 6 && strings.HasPrefix(fields[5], "[vsyscall]") {
			continue
		}

		q, name, _, err := Query(fd, start, CoveringOrNextVMA)
		name = strings.TrimRight(name, "\x00")
		if err != nil {
			t.Errorf("ioctl query failed at %#x: %v", start, err)
			continue
		}
		if q.VMAStart != start || q.VMAEnd != end {
			t.Errorf("mismatch at %#x: got range %#x-%#x, want %#x-%#x", start, q.VMAStart, q.VMAEnd, start, end)
		}
		if len(fields) >= 6 {
			expectedPath := fields[5]
			if name != "" && !bytes.HasSuffix([]byte(name), []byte(expectedPath)) {
				t.Logf("warning: VMA name mismatch: got %q, want suffix %q", name, expectedPath)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}
}
