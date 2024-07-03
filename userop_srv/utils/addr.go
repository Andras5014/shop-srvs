package utils

import (
	"net"
)

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := l.Close(); err != nil {
			return
		}
	}()
	return l.Addr().(*net.TCPAddr).Port, nil
}
