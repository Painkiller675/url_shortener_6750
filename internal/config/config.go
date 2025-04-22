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
	CertFile     string
	KeyFile      string
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

// UnmOptions is used to unmarshal config file
var UnmOptions ummarshalOptions

// var postgreConStr = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
// .
//
//	`localhost`, `postgres`, "12345678", `url_shortener`)
//
// var postgreConStr = "user=postgres password=12345678 dbname=url_shortener sslmode=disable"

// SetConfig sets config via cmline or environment
func SetConfig() error {
	//var StartOptions Options
	flag.StringVar(&StartOptions.HTTPServer.Address, "a", ":8080", "HTTP-server address")
	flag.StringVar(&StartOptions.BaseURL, "b", "", "base URL")
	flag.StringVar(&StartOptions.LogLvl, "l", "info", "log level")
	flag.StringVar(&StartOptions.Filename, "f", "", "storage filename")
	flag.StringVar(&StartOptions.DBConStr, "d", "", "DSN (for database)")
	flag.BoolVar(&StartOptions.HTTPSEnabled, "s", false, "to deactivate https mode use -s false ")
	flag.StringVar(&StartOptions.JSONConfig, "c", "", "path to a json config")
	flag.StringVar(&StartOptions.CertFile, "certFile", "../../internal/cert/localhost.pem", "tls certificate file path")
	flag.StringVar(&StartOptions.KeyFile, "keyFile", "../../internal/cert/localhost-key.pem", "tls key file path")
	// set version in usage output
	flag.Usage = func() {
		// TODO: How should I handle this error the best???
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %v\nUsage of %s:\n", version, os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	// TODO: How do that using caarlos0/env in the best way?
	// try to figure out if we have a json config or not
	if envJSONConfig := os.Getenv("CONFIG"); envJSONConfig != "" {
		StartOptions.JSONConfig = envJSONConfig
	}
	// if we have the flag with json config path
	if StartOptions.JSONConfig != "" {
		// try to read config data from json config file

		file, err := os.ReadFile(StartOptions.JSONConfig)
		if err != nil {
			return err
		}
		err = json.Unmarshal(file, &UnmOptions)
		if err != nil {
			return err
		}

	}

	if envLogLvl := os.Getenv("LOG_LEVEL"); envLogLvl != "" {
		StartOptions.LogLvl = envLogLvl
	}
	if envCertFile := os.Getenv("CERT_FILE"); envCertFile != "" {
		StartOptions.CertFile = envCertFile
	}
	if envKeyFile := os.Getenv("KEY_FILE"); envKeyFile != "" {
		StartOptions.KeyFile = envKeyFile
	}

	//ENV values (if set => use them else use   flags)
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		StartOptions.HTTPServer.Address = envRunAddr
	} else {
		// if flags are set => assigning
		if isFlagPassed("a") {
			// assigning set value
		} else {
			// assign config parameters (from json)
			if StartOptions.JSONConfig != "" {
				// if we have smth in config file
				if UnmOptions.ServerAddress != "" {
					StartOptions.HTTPServer.Address = UnmOptions.ServerAddress
				}

			} // else ==> DEFAULT values will be set

		}
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		StartOptions.BaseURL = envBaseURL
	} else {
		// if flags are set => assigning
		if isFlagPassed("b") {
			// assigning set value
		} else {
			// assign config parameters (from json)
			if StartOptions.JSONConfig != "" {
				if UnmOptions.BaseURL != "" {
					StartOptions.BaseURL = UnmOptions.BaseURL
				}
			} // else ==> DEFAULT values will be set

		}
	}

	if envFilename := os.Getenv("FILE_STORAGE_PATH"); envFilename != "" {
		StartOptions.Filename = envFilename
	} else {
		// if flags are set => assigning
		if isFlagPassed("f") {
			// assigning set value
		} else {
			// assign config parameters (from json)
			if StartOptions.JSONConfig != "" {
				if UnmOptions.Filename != "" {
					StartOptions.Filename = UnmOptions.Filename
				}
			} // else ==> DEFAULT values will be set

		}
	}

	if envDsnDB := os.Getenv("DATABASE_DSN"); envDsnDB != "" {
		StartOptions.DBConStr = envDsnDB
	} else {
		// if flags are set => assigning
		if isFlagPassed("d") {
			// assigning set value
		} else {
			// assign config parameters (from json)
			if StartOptions.JSONConfig != "" {
				if UnmOptions.DBConStr != "" {
					StartOptions.DBConStr = UnmOptions.DBConStr
				}
			} // else ==> DEFAULT values will be set

		}
	}

	if envHTTPSEnabled := os.Getenv("ENABLE_HTTPS"); envHTTPSEnabled != "" { // TODO mb use  os.LookupEnv() use bool value
		var err error
		StartOptions.HTTPSEnabled, err = strconv.ParseBool(envHTTPSEnabled) //
		if err != nil {
			return err
		}
	} else {
		// if flags are set => assigning
		if isFlagPassed("s") {
			// assigning set value
		} else {
			// assign config parameters (from json)
			if StartOptions.JSONConfig != "" { // TODO [MENTOR]: should I somehow validate this data from config?
				StartOptions.HTTPSEnabled = UnmOptions.HTTPSEnabled
			} // else ==> DEFAULT values will be set

		}
	}
	return nil
}

// isFlagPassed is used to figure out if console flag was set by user
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
