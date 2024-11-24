package main

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port int
}

func (a NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("Need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}

var (
	flagPollInterval   int
	flagReportInterval int
	flagNetAddress     NetAddress
)

func parseFlags() {
	flagNetAddress = NetAddress{
		Host: "localhost",
		Port: 8080,
	}
	flag.Var(&flagNetAddress, "a", "poll interval (sec)")

	flag.IntVar(&flagPollInterval, "p", 2, "poll interval (sec)")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval (sec")
	flag.Parse()
}
