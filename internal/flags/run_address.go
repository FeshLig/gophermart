package flags

import (
	"net"
	"strconv"
)

type RunAddress struct {
	Host string
	Port int
}

func (a RunAddress) String() string {
	return a.Host + ":" + strconv.FormatInt(int64(a.Port), 10)
}

func (a *RunAddress) Set(s string) error {
	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	if host == "" {
		host = "localhost"
	}

	a.Host = host
	a.Port = port

	return nil
}
