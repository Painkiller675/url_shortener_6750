// Config is used for setting initial options in the project (ip,baseURL,log level, storage filename, DSN (for DB).
// It also includes token expiration date and SecretKey
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
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
	BaseURL      string
	LogLvl       string // flag
	Filename     string
	DBConStr     string
	HTTPSEnabled bool
	JSONConfig   string
	HTTPServer
}

// ummarshalOptions - is used to unmarshal json config file
type ummarshalOptions struct {
	ServerAddress string `json:"server_address"`
	BaseURL       string `json:"base_url"`
	Filename      string `json:"file_storage_path"`
	DBConStr      string `json:"database_dsn"`
	HTTPSEnabled  bool   `json:"enable_https"`
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
	flag.StringVar(&StartOptions.HTTPServer.Address, "a", "", "HTTP-server address")
	flag.StringVar(&StartOptions.BaseURL, "b", "", "base URL")
	flag.StringVar(&StartOptions.LogLvl, "l", "info", "log level")
	flag.StringVar(&StartOptions.Filename, "f", "", "storage filename")
	flag.StringVar(&StartOptions.DBConStr, "d", "", "DSN (for database)")
	flag.BoolVar(&StartOptions.HTTPSEnabled, "s", false, "to deactivate https mode use -s false ")
	flag.StringVar(&StartOptions.JSONConfig, "c", "config.json", "path to a json config")
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
	if envHTTPSEnabled := os.Getenv("ENABLE_HTTPS"); envHTTPSEnabled != "" { // TODO use  os.LookupEnv() use bool value
		var err error
		StartOptions.HTTPSEnabled, err = strconv.ParseBool(envHTTPSEnabled) // TODO:
		if err != nil {
			panic(err)
		} // TODO: лучше вернуть ошибку и сделать лог фатал в main
	}
	if envJSONConfig := os.Getenv("CONFIG"); envJSONConfig != "" {
		StartOptions.JSONConfig = envJSONConfig
	}

	if StartOptions.JSONConfig != "" {
		// try to read config data from json config file
		var unmOptions ummarshalOptions
		file, err := os.ReadFile(StartOptions.JSONConfig)
		if err != nil {
			panic(err)
			return
		}
		err = json.Unmarshal(file, &unmOptions)
		if err != nil {
			panic(err)
			return
		}
		// try to reassign config parameters (from json)
		if StartOptions.HTTPServer.Address == "" {
			StartOptions.HTTPServer.Address = unmOptions.ServerAddress
		}
		if StartOptions.BaseURL == "" { // TODO: is it ok?
			StartOptions.BaseURL = unmOptions.BaseURL
		}
		if StartOptions.Filename == "" {
			StartOptions.Filename = unmOptions.Filename
		}
		if StartOptions.DBConStr == "" {
			StartOptions.DBConStr = unmOptions.DBConStr
		}
		if StartOptions.HTTPSEnabled { // TODO: HOW TO HANDLE IT ????!
			// описать как ссылку на бул HTTPSEnabled (в енв ничего нет не заполняю то, есди в флагах ничего нет не заполняю)
			//или сразу заполнить из json а потом сверху накладывать
			StartOptions.HTTPSEnabled = unmOptions.HTTPSEnabled
		}

	}

}
