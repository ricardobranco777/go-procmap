
![Build Status](https://github.com/ricardobranco777/go-procmap/actions/workflows/ci.yml/badge.svg)

Call **PROCMAP_QUERY** ioctl from Golang on Linux.

From the [documentation](https://docs.kernel.org/filesystems/proc.html):

> Starting with 6.11 kernel, /proc/PID/maps provides an alternative ioctl()-based API
> that gives ability to flexibly and efficiently query and filter individual VMAs.
> This interface is binary and is meant for more efficient and easy programmatic use.
> struct procmap_query, defined in [linux/fs.h](https://github.com/torvalds/linux/blob/master/include/uapi/linux/fs.h)
> UAPI header, serves as an input/output argument to the PROCMAP_QUERY ioctl() command.

I saw this ioctl in this [LWN article](https://lwn.net/Articles/1026749/)

The example code is a command to dump the information (plus the ELF build-ID) in a
format similar to `/proc/<pid>/maps` from a given PID.

Similar project in Python:
- https://github.com/geofft/pycon2025/tree/main/procmapquery

More information at:
- https://github.com/torvalds/linux/blob/master/include/uapi/linux/fs.h
