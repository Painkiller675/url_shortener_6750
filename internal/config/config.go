// Config is used for setting initial options in the project (ip,baseURL,log level, storage filename, DSN (for DB).
// It also includes token expiration date and SecretKey
package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// TokenExp - JWT token expiration time
const TokenExp = time.Hour * 3

// SecretKey - secret key for JWT token
const SecretKey = "supersecretkey"

var version = "4.0" +
	""

// Options - basic parameters of the server
type Options struct {
	BaseURL  string
	LogLvl   string // flag
	Filename string
	DBConStr string
	HTTPServer
}

// HTTPServer - embedded basic parameters of the server
type HTTPServer struct {
	Address     string
	Timeout     time.Duration
	IdleTimeout time.Duration
}

// StartOptions - for flags
var StartOptions Options

// var postgreConStr = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
// .
//
//	`localhost`, `postgres`, "12345678", `url_shortener`)
//
// var postgreConStr = "user=postgres password=12345678 dbname=url_shortener sslmode=disable"
// SetConfig sets config via cmline or environment
func SetConfig() {
	//var StartOptions Options
	flag.StringVar(&StartOptions.HTTPServer.Address, "a", "localhost:8080", "HTTP-server address")
	flag.StringVar(&StartOptions.BaseURL, "b", "http://localhost:8080/", "base URL")
	flag.StringVar(&StartOptions.LogLvl, "l", "info", "log level")
	flag.StringVar(&StartOptions.Filename, "f", "", "storage filename")
	flag.StringVar(&StartOptions.DBConStr, "d", "", "DSN (for database)")
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
	if envFilename := os.Getenv("FILE_STORAGE_PATH"); envFilename != "" {
		StartOptions.Filename = envFilename
	}
	if envDsnDB := os.Getenv("DATABASE_DSN"); envDsnDB != "" {
		StartOptions.DBConStr = envDsnDB
	}
}
