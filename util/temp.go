package util

import (
	"io/ioutil"
	"net"
)

func TmpDir() string {
	tmp, _ := ioutil.TempDir("", "")
	return tmp
}

func TmpPort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return -1
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return -1
	}

	defer func() {
		_ = l.Close()
	}()

	return l.Addr().(*net.TCPAddr).Port
}
