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

// String нужен для интерфейса flag.Value
func (a AccuralSystemAddress) String() string {
	if a.Port <= 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// Set разбирает строку вида host:port
func (a *AccuralSystemAddress) Set(s string) error {
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

// URL возвращает полный URL с http протоколом, например: http://localhost:8000
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
