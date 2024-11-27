package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/ossrs/go-oryx-lib/errors"
	"github.com/ossrs/go-oryx-lib/logger"
)
var conf *Config

func init() {
    conf = NewConfig()
}
func main() {
    ctx := logger.WithContext(context.Background())
    ctx = logger.WithContext(ctx)

    if err := doMain(ctx); err != nil {
        logger.Tf(ctx, "run err %+v", err)
        return
    }

    logger.Tf(ctx, "run ok")
}

func doMain(ctx context.Context) error {
	flag.Parse()

	// Install signals.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for s := range sc {
			logger.Tf(ctx, "Got signal %v", s)
			cancel()
		}
	}()

	// When cancelled, the program is forced to exit due to a timeout. Normally, this doesn't occur
	// because the main thread exits after the context is cancelled. However, sometimes the main thread
	// may be blocked for some reason, so a forced exit is necessary to ensure the program terminates.
	go func() {
		<-ctx.Done()
		time.Sleep(30 * time.Second)
		logger.Wf(ctx, "Force to exit by timeout")
		os.Exit(1)
	}()

	// Initialize the management password and load the environment without relying on Redis.
	if true {
		if pwd, err := os.Getwd(); err != nil {
			return errors.Wrapf(err, "getpwd")
		} else {
			conf.Pwd = pwd
		}

		// Note that we only use .env in mgmt.
		envFile := path.Join(conf.Pwd, "containers/data/config/.env")
		if _, err := os.Stat(envFile); err == nil {
			if err := godotenv.Overload(envFile); err != nil {
				return errors.Wrapf(err, "load %v", envFile)
			}
		}
	}
	// For platform, default to development for Darwin.
	setEnvDefault("NODE_ENV", "development")
	// For platform, HTTP server listen port.
	setEnvDefault("PLATFORM_LISTEN", "2024")
	// Set the default language, en or zh.
	setEnvDefault("REACT_APP_LOCALE", "en")
	// Whether enable the Go pprof.
	setEnvDefault("GO_PPROF", "")

	// Migrate from mgmt.
	setEnvDefault("REDIS_DATABASE", "0")
	setEnvDefault("REDIS_HOST", "127.0.0.1")
	setEnvDefault("REDIS_PORT", "6379")
	setEnvDefault("MGMT_LISTEN", "2022")

	// For HTTPS.
	setEnvDefault("HTTPS_LISTEN", "2443")
	setEnvDefault("AUTO_SELF_SIGNED_CERTIFICATE", "on")

	// For feature control.
	setEnvDefault("NAME_LOOKUP", "on")
	setEnvDefault("PLATFORM_DOCKER", "off")

	// For multiple ports.
	setEnvDefault("RTMP_PORT", "1935")
	setEnvDefault("HTTP_PORT", "")
	setEnvDefault("SRT_PORT", "10080")
	setEnvDefault("RTC_PORT", "8000")

	// For system limit.
	setEnvDefault("SRS_FORWARD_LIMIT", "10")
	setEnvDefault("SRS_VLIVE_LIMIT", "10")
	setEnvDefault("SRS_CAMERA_LIMIT", "10")
	
	if err := InitRdb(); err != nil {
		return errors.Wrapf(err, "init rdb")
	}
	logger.Tf(ctx, "init rdb(redis client) ok")
    
    httpService := NewHTTPService()
	defer httpService.Close()
	if err := httpService.Run(ctx); err != nil {
		return errors.Wrapf(err, "start http service")
	}

    return nil
}