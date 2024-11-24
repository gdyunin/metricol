package main

import (
	"flag"
)

var (
	flagRunAddr        string
	flagPollInterval   int
	flagReportInterval int
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "server address and port")
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval (sec)")
	flag.IntVar(&flagReportInterval, "r", 10, "report interval (sec")
	flag.Parse()
}
