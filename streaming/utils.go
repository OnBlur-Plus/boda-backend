package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/ossrs/go-oryx-lib/errors"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	IsDarwin bool
	// Current working directory, at xxx/oryx/platform.
	Pwd string
}

func NewConfig() *Config {
	return &Config{
		IsDarwin: runtime.GOOS == "darwin",
	}
}

func (v *Config) String() string {
	return fmt.Sprintf("darwin=%v, pwd=%v", v.IsDarwin, v.Pwd)
}

const (
	// For SRS stream status.
	SRS_HP_HLS = "SRS_HP_HLS"
	SRS_LL_HLS = "SRS_LL_HLS"
	// For SRS stream status.
	SRS_STREAM_ACTIVE     = "SRS_STREAM_ACTIVE"
	SRS_STREAM_SRT_ACTIVE = "SRS_STREAM_SRT_ACTIVE"
	SRS_STREAM_RTC_ACTIVE = "SRS_STREAM_RTC_ACTIVE"
	// For feature statistics.
	SRS_STAT_COUNTER = "SRS_STAT_COUNTER"
	// For container and images.
	SRS_CONTAINER_DISABLED = "SRS_CONTAINER_DISABLED"
	// For live stream and rooms.
	SRS_LIVE_ROOM = "SRS_LIVE_ROOM"
	// For system settings.
	SRS_FIRST_BOOT      = "SRS_FIRST_BOOT"
	SRS_UPGRADING       = "SRS_UPGRADING"
	SRS_UPGRADE_WINDOW  = "SRS_UPGRADE_WINDOW"
	// SRS_HTTPS           = "SRS_HTTPS"
	// SRS_HTTPS_DOMAIN    = "SRS_HTTPS_DOMAIN"
	SRS_HOOKS           = "SRS_HOOKS"
	SRS_SYS_LIMITS      = "SRS_SYS_LIMITS"
)

func envPlatformListen() string {
	return os.Getenv("PLATFORM_LISTEN")
}
func envMgmtListen() string {
	return os.Getenv("MGMT_LISTEN")
}

func envRegion() string {
	return os.Getenv("REGION")
}

func envSource() string {
	return os.Getenv("SOURCE")
}

func envHttpListen() string {
	return os.Getenv("HTTPS_LISTEN")
}

func envRedisPassword() string {
	return os.Getenv("REDIS_PASSWORD")
}

func envRedisPort() string {
	return os.Getenv("REDIS_PORT")
}

func envRedisHost() string {
	return os.Getenv("REDIS_HOST")
}

func envRedisDatabase() string {
	return os.Getenv("REDIS_DATABASE")
}

func envRtmpPort() string {
	return os.Getenv("RTMP_PORT")
}

func envPublicUrl() string {
	return os.Getenv("PUBLIC_URL")
}

func envBuildPath() string {
	return os.Getenv("BUILD_PATH")
}

func envHttpPort() string {
	return os.Getenv("HTTP_PORT")
}

func envPath() string {
	return os.Getenv("PATH")
}

func envGoPprof() string {
	return os.Getenv("GO_PPROF")
}

// setEnvDefault set env key=value if not set.
func setEnvDefault(key, value string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}


// rdb is a global redis client object.
var rdb *redis.Client

// InitRdb create and init global rdb, which is a redis client.
func InitRdb() error {
	redisDatabase, err := strconv.Atoi(envRedisDatabase())
	if err != nil {
		return errors.Wrapf(err, "invalid REDIS_DATABASE %v", envRedisDatabase())
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", envRedisHost(), envRedisPort()),
		Password: envRedisPassword(),
		DB:       redisDatabase,
	})
	return nil
}
