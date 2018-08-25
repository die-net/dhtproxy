package main

import (
	"flag"
	"log"
	"syscall"
)

var maxConnections = flag.Int("maxConnections", getRlimitMax(syscall.RLIMIT_NOFILE), "The number of incoming connections to allow.")

func getRlimitMax(resource int) int {
	var rlimit syscall.Rlimit

	if err := syscall.Getrlimit(resource, &rlimit); err != nil {
		return 0
	}

	return int(rlimit.Max)
}

func setRlimit(resource, value int) {
	rlimit := &syscall.Rlimit{Cur: uint64(value), Max: uint64(value)}

	err := syscall.Setrlimit(resource, rlimit)
	if err != nil {
		log.Fatalln("Error Setting Rlimit ", err)
	}
}

func setRlimitFromFlags() {
	setRlimit(syscall.RLIMIT_NOFILE, *maxConnections)
}
