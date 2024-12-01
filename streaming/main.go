// Copyright (c) 2022-2024 Winlin
//
// SPDX-License-Identifier: MIT
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/ossrs/go-oryx-lib/errors"
	"github.com/ossrs/go-oryx-lib/logger"
)

// conf is a global config object.
var conf *Config

func init() {
	// certManager = NewCertManager()
	conf = NewConfig()

	// We use polling to update some fast cache, for example, LLHLS config.
	fastCache = NewFastCache()
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
		time.Sleep(15 * time.Second)
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

	// For platform, HTTP server listen port.
	setEnvDefault("PLATFORM_LISTEN", "2024")
	// Set the default language, en or zh.
	setEnvDefault("GO_PPROF", "")

	// Migrate from mgmt.
	setEnvDefault("REDIS_DATABASE", "0")
	setEnvDefault("REDIS_HOST", "127.0.0.1")
	setEnvDefault("REDIS_PORT", "6379")
	setEnvDefault("MGMT_LISTEN", "2022")

	// For HTTPS.
	setEnvDefault("HTTPS_LISTEN", "2443")
	setEnvDefault("AUTO_SELF_SIGNED_CERTIFICATE", "on")

	// For multiple ports.
	setEnvDefault("RTMP_PORT", "1935")
	setEnvDefault("HTTP_PORT", "")

	logger.Tf(ctx, "load .env as GO_PPROF=%v, API_SECRET=%vB, SOURCE=%v, REDIS_DATABASE=%v, REDIS_HOST=%v, REDIS_PASSWORD=%vB, REDIS_PORT=%v, "+
		"RTMP_PORT=%v, PUBLIC_URL=%v, BUILD_PATH=%v, PLATFORM_LISTEN=%v, HTTP_PORT=%v, HTTPS_LISTEN=%v, MGMT_LISTEN=%v",
		envGoPprof(), len(envApiSecret()), envSource(), envRedisDatabase(), envRedisHost(), len(envRedisPassword()), envRedisPort(),
		envRtmpPort(), envPublicUrl(), envBuildPath(), envPlatformListen(), envHttpPort(), envHttpListen(), envMgmtListen(),
	)

	// Start the Go pprof if enabled.
	if addr := envGoPprof(); addr != "" {
		go func() {
			logger.Tf(ctx, "Start Go pprof at %v", addr)
			http.ListenAndServe(addr, nil)
		}()
	}

	// Initialize global rdb, the redis client.
	if err := InitRdb(); err != nil {
		return errors.Wrapf(err, "init rdb")
	}
	logger.Tf(ctx, "init rdb(redis client) ok")

	// For platform, we should initOS after redis.
	// Setup the OS for redis, which should never depends on redis.
	if err := initialize(ctx); err != nil {
		return errors.Wrapf(err, "init os")
	}

	accidentWorker = NewAccidentWorker()
	defer accidentWorker.Close()
	if err := accidentWorker.Start(ctx); err != nil {
		return errors.Wrapf(err, "start accident worker")
	}

	detectWorker = NewDetectWorker()
	defer detectWorker.Close()
	if err := detectWorker.Start(ctx); err != nil {
		return errors.Wrapf(err, "start detect worker")
	}

	// Run HTTP service.
	httpService := NewHTTPService()
	defer httpService.Close()
	if err := httpService.Run(ctx); err != nil {
		return errors.Wrapf(err, "start http service")
	}

	return nil
}
// Initialize before thread run.
func initialize(ctx context.Context) error {
	// For Darwin, append the search PATH for docker.
	// Note that we should set the PATH env, not the exec.Cmd.Env.
	// Note that it depends on conf.IsDarwin, so it's unavailable util initOS.
	if conf.IsDarwin && !strings.Contains(envPath(), "/usr/local/bin") {
		os.Setenv("PATH", fmt.Sprintf("%v:/usr/local/bin", envPath()))
	}

	// Create directories for data, allow user to link it.
	// Keep in mind that the containers/data/srs-s3-bucket maybe mount by user, because user should generate
	// and mount it if they wish to save recordings to cloud storage.
	for _, dir := range []string{
		"containers/data/record", "containers/data/config", "containers/data/accident", "containers/data/detect", "containers/data/process",
		// "containers/data/dvr", "containers/data/vod",
		// "containers/data/upload", "containers/data/vlive", "containers/data/signals",
		// "containers/data/lego", "containers/data/.well-known",
		// "containers/data/transcript", "containers/data/srs-s3-bucket", "containers/data/ai-talk",
		// "containers/data/dubbing", "containers/data/ocr",
	} {
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModeDir|os.FileMode(0755)); err != nil {
				return errors.Wrapf(err, "create dir %v", dir)
			}
		}
	}

	// Migrate from previous versions.
	// for _, migrate := range []struct {
	// 	PVK string
	// 	CVK string
	// }{
	// 	{"SRS_RECORD_M3U8_METADATA", SRS_RECORD_M3U8_ARTIFACT},
	// 	{"SRS_DVR_M3U8_METADATA", SRS_DVR_M3U8_ARTIFACT},
	// 	{"SRS_VOD_M3U8_METADATA", SRS_VOD_M3U8_ARTIFACT},
	// } {
	// 	pv, _ := rdb.HLen(ctx, migrate.PVK).Result()
	// 	cv, _ := rdb.HLen(ctx, migrate.CVK).Result()
	// 	if pv > 0 && cv == 0 {
	// 		if vs, err := rdb.HGetAll(ctx, migrate.PVK).Result(); err == nil {
	// 			for k, v := range vs {
	// 				_ = rdb.HSet(ctx, migrate.CVK, k, v)
	// 			}
	// 			logger.Tf(ctx, "migrate %v to %v with %v keys", migrate.PVK, migrate.CVK, len(vs))
	// 		}
	// 	}
	// }

	return nil
}