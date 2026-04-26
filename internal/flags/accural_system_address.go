package flags

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type AccuralSystemAddress struct {
	Host string
	Port int
}

func (a AccuralSystemAddress) String() string {
	if a.Port <= 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

func (a *AccuralSystemAddress) Set(s string) error {

	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		u, err := url.Parse(s)
		if err != nil {
			return err
		}
		host := u.Hostname()
		portStr := u.Port()
		if portStr == "" {
			return errors.New("port is required")
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return err
		}
		a.Host = host
		a.Port = port
		return nil
	}

	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need service address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}

func (a AccuralSystemAddress) URL() string {
	if a.Host == "" || a.Port <= 0 {
		return ""
	}
	u := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", a.Host, a.Port),
	}
	return u.String()
}
