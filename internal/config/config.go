package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var version = "4.0" +
	""

type Options struct {
	BaseURL string
	HTTPServer
}

type HTTPServer struct {
	Address     string
	Timeout     time.Duration
	IdleTimeout time.Duration
}

var StartOptions Options

func SetConfig() {
	//var StartOptions Options
	flag.StringVar(&StartOptions.HTTPServer.Address, "a", "localhost:8080", "HTTP-server address")
	flag.StringVar(&StartOptions.BaseURL, "b", "http://localhost:8080/", "base URL")

	flag.Usage = func() {
		// TODO: How should I handle this error the best???
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %v\nUsage of %s:\n", version, os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
}
