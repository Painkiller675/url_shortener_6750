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
	Filename string
	BaseURL  string
	LogLvl   string // flag
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
	flag.StringVar(&StartOptions.LogLvl, "l", "info", "log level")
	flag.StringVar(&StartOptions.Filename, "f", "./stor.json", "path to the file to save the storage")
	// set version in usage output
	flag.Usage = func() {
		// TODO: How should I handle this error the best???
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %v\nUsage of %s:\n", version, os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	// TODO: How do that using caarlos0/env in the best way?

	//ENV values (if set => use them else use   flags)
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		StartOptions.HTTPServer.Address = envRunAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		StartOptions.BaseURL = envBaseURL
	}
	if envLogLvl := os.Getenv("LOG_LEVEL"); envLogLvl != "" {
		StartOptions.LogLvl = envLogLvl
	}
	if envFile := os.Getenv("FILE_STORAGE_PATH"); envFile != "" {
		StartOptions.Filename = envFile
	}
}
