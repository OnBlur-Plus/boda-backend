package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/ossrs/go-oryx-lib/errors"
	"github.com/redis/go-redis/v9"
)

type Category struct {
	ID     int
	Type string
}

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
	SRS_AUTH_SECRET = "SRS_AUTH_SECRET"
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
	
	RECORD_M3U8_ARTIFACT = "RECORD_M3U8_ARTIFACT"
	PROCESS_TASK = "PROCESS_TASK"
	PROCESS_STREAM_WORKING = "PROCESS_STREAM_WORKING"
	SRS_ACCIDENT_M3U8_WORKING = "SRS_ACCIDENT_M3U8_WORKING"
	SRS_ACCIDENT_M3U8_ARTIFACT = "SRS_ACCIDENT_M3U8_ARTIFACT"
)

// GenerateRoomPublishKey to build the redis hashset key from room stream name.
func GenerateRoomPublishKey(roomStreamName string) string {
	return fmt.Sprintf("room-pub-%v", roomStreamName)
}

func envApiSecret() string {
	return os.Getenv("API_SECRET")
}

func envSource() string {
	return os.Getenv("SOURCE")
}

func envPlatformListen() string {
	return os.Getenv("PLATFORM_LISTEN")
}

func envMgmtListen() string {
	return os.Getenv("MGMT_LISTEN")
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

// buildLiveM3u8ForLocal go generate dynamic m3u8.
func buildLiveM3u8ForLocal(
	ctx context.Context, tsFiles []*TsFile, useKey bool, prefix string, metadata []string,
) (
	contentType, m3u8Body string, duration float64, err error,
) {
	if len(tsFiles) == 0 {
		err = errors.Errorf("no files")
		return
	}

	for _, file := range tsFiles {
		duration = math.Max(duration, file.Duration)
	}

	first := tsFiles[0]

	m3u8 := []string{
		"#EXTM3U",
		"#EXT-X-VERSION:3",
		fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%v", first.SeqNo),
		fmt.Sprintf("#EXT-X-TARGETDURATION:%v", math.Ceil(duration)),
	}
	for index, file := range tsFiles {
		// TODO: FIXME: Identify discontinuity by callback.
		if index < len(tsFiles)-2 {
			next := tsFiles[index+1]
			if file.SeqNo+1 != next.SeqNo {
				m3u8 = append(m3u8, "#EXT-X-DISCONTINUITY")
			}
		}
		m3u8 = append(m3u8, fmt.Sprintf("#BOUNDING-BOX:%v", metadata[index]))

		m3u8 = append(m3u8, fmt.Sprintf("#EXTINF:%.2f, no desc", file.Duration))

		var tsURL string
		if useKey {
			tsURL = fmt.Sprintf("%v%v", prefix, file.Key)
		} else {
			tsURL = fmt.Sprintf("%v%v.ts", prefix, file.TsID)
		}
		m3u8 = append(m3u8, tsURL)
	}

	contentType = "application/vnd.apple.mpegurl"
	m3u8Body = strings.Join(m3u8, "\n")
	return
}

// slicesContains is a function to check whether elem in arr.
func slicesContains(arr []string, elem string) bool {
	for _, e := range arr {
		if e == elem {
			return true
		}
	}
	return false
}

// TsFile is a ts file object.
type TsFile struct {
	// The identify key of TS file, renamed local ts path or COS key, format is record/{m3u8UUID}/{tsID}.ts
	// For example, record/3ECF0239-708C-42E4-96E1-5AE935C6E6A9/5B7B5C03-8DB4-4ABA-AAF3-CB55902CF177.ts
	// For example, 3ECF0239-708C-42E4-96E1-5AE935C6E6A9/5B7B5C03-8DB4-4ABA-AAF3-CB55902CF177.ts
	// Note that for DVR and VoD, the key is key of COS bucket object.
	// Note that for RECORD, the key is the final ts file path.
	// Note that for Transcript, the key is not used.
	Key string `json:"key,omitempty"`
	// The TS local ID, a uuid string, such as 5B7B5C03-8DB4-4ABA-AAF3-CB55902CF177
	TsID string `json:"tsid,omitempty"`
	// The TS local file, format is record/:uuid.ts, such as record/5B7B5C03-8DB4-4ABA-AAF3-CB55902CF177.ts
	File string `json:"tsfile,omitempty"`
	// The TS url, generated by SRS, such as live/livestream/2015-04-23/01/476584165.ts
	URL string `json:"url,omitempty"`
	// The seqno of TS, generated by SRS, such as 100
	SeqNo uint64 `json:"seqno,omitempty"`
	// The duration of TS in seconds, generated by SRS, such as 9.36
	Duration float64 `json:"duration,omitempty"`
	// The size of TS file in bytes, such as 1934897
	Size uint64 `json:"size,omitempty"`
}

func (v *TsFile) String() string {
	return fmt.Sprintf("key=%v, id=%v, url=%v, seqno=%v, duration=%v, size=%v, file=%v",
		v.Key, v.TsID, v.URL, v.SeqNo, v.Duration, v.Size, v.File,
	)
}

// SrsOnHlsMessage is the SRS on_hls callback message.
type SrsOnHlsMessage struct {
	// Must be on_hls
	Action SrsAction `json:"action,omitempty"`
	// The ts file path, such as ./objs/nginx/html/live/livestream/2015-04-23/01/476584165.ts
	File string `json:"file,omitempty"`
	// The duration of ts file, in seconds, such as 9.36
	Duration float64 `json:"duration,omitempty"`
	// The url of m3u8, such as live/livestream/live.m3u8
	M3u8URL string `json:"m3u8_url,omitempty"`
	// The sequence number of ts, such as 100
	SeqNo uint64 `json:"seq_no,omitempty"`

	// The vhost of stream, generated by SRS, such as video.test.com
	Vhost string `json:"vhost,omitempty"`
	// The app of stream, generated by SRS, such as live
	App string `json:"app,omitempty"`
	// The name of stream, generated by SRS, such as livestream
	Stream string `json:"stream,omitempty"`

	// The TS url, generated by SRS, such as live/livestream/2015-04-23/01/476584165.ts
	URL string `json:"url,omitempty"`
}

func (v *SrsOnHlsMessage) String() string {
	return fmt.Sprintf("action=%v, file=%v, duration=%v, seqno=%v, m3u8_url=%v, vhost=%v, "+
		"app=%v, stream=%v, url=%v",
		v.Action, v.File, v.Duration, v.SeqNo, v.M3u8URL, v.Vhost, v.App, v.Stream, v.URL,
	)
}

// SrsOnHlsObject contains a SrsOnHlsMessage and a local TsFile.
type SrsOnHlsObject struct {
	Msg    *SrsOnHlsMessage `json:"msg"`
	TsFile *TsFile          `json:"tsfile"`
}

func (v *SrsOnHlsObject) String() string {
	return fmt.Sprintf("msg(%v), ts(%v)", v.Msg.String(), v.TsFile.String())
}

// VodTranscodeTemplate is the transcode template for VoD.
type VodTranscodeTemplate struct {
	// In query template API, it's string. See https://cloud.tencent.com/document/product/266/33769
	// In remux task API, it's integer. See https://cloud.tencent.com/document/product/266/33427
	Definition string `json:"definition"`
	Name       string `json:"name"`
	Comment    string `json:"comment"`
	Container  string `json:"container"`
	Update     string `json:"update"`
}

func (v *VodTranscodeTemplate) String() string {
	return fmt.Sprintf("definition=%v, name=%v, comment=%v, container=%v, update=%v",
		v.Definition, v.Name, v.Comment, v.Container, v.Update,
	)
}

// SrsStream is a stream in SRS.
type SrsStream struct {
	Vhost  string `json:"vhost,omitempty"`
	App    string `json:"app,omitempty"`
	Stream string `json:"stream,omitempty"`
	Param  string `json:"param,omitempty"`

	Server string `json:"server_id,omitempty"`
	Client string `json:"client_id,omitempty"`

	Update string `json:"update,omitempty"`
}

func (v *SrsStream) String() string {
	return fmt.Sprintf("vhost=%v, app=%v, stream=%v, param=%v, server=%v, client=%v, update=%v",
		v.Vhost, v.App, v.Stream, v.Param, v.Server, v.Client, v.Update,
	)
}

func (v *SrsStream) StreamURL() string {
	streamURL := fmt.Sprintf("%v/%v/%v", v.Vhost, v.App, v.Stream)
	if v.Vhost == "__defaultVhost__" {
		streamURL = fmt.Sprintf("%v/%v", v.App, v.Stream)
	}
	return streamURL
}

func (v *SrsStream) IsSRT() bool {
	return strings.Contains(v.Param, "upstream=srt")
}

func (v *SrsStream) IsRTC() bool {
	return strings.Contains(v.Param, "upstream=rtc")
}

// ParseBody read the body from r, and unmarshal JSON to v.
func ParseBody(ctx context.Context, r io.ReadCloser, v interface{}) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return errors.Wrapf(err, "read body")
	}
	defer r.Close()

	if len(b) == 0 {
		return nil
	}

	if err := json.Unmarshal(b, v); err != nil {
		return errors.Wrapf(err, "json unmarshal %v", string(b))
	}

	return nil
}

// httpAllowCORS allow CORS for HTTP request.
// Note that we always enable CROS because we enable HTTP cache.
func httpAllowCORS(w http.ResponseWriter, r *http.Request) {
	// SRS does not need cookie or credentials, so we disable CORS credentials, and use * for CORS origin,
	// headers, expose headers and methods.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
	w.Header().Set("Access-Control-Allow-Headers", "*")
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
	w.Header().Set("Access-Control-Allow-Methods", "*")
	// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Expose-Headers
	w.Header().Set("Access-Control-Expose-Headers", "*")
}

// httpCreateProxy create a reverse proxy for target URL.
func httpCreateProxy(targetURL string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, errors.Wrapf(err, "parse backend %v", targetURL)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ModifyResponse = func(resp *http.Response) error {
		// We will set the server field.
		resp.Header.Del("Server")

		// We will set the CORS headers.
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Allow-Methods")
		resp.Header.Del("Access-Control-Expose-Headers")
		resp.Header.Del("Access-Control-Allow-Credentials")

		// Not used right now.
		resp.Header.Del("Access-Control-Request-Private-Network")

		return nil
	}

	return proxy, nil
}

// FFprobeFormat is the format object in ffprobe response.
type FFprobeFormat struct {
	// The start time in seconds.
	Starttime string `json:"start_time"`
	// The duration in seconds.
	Duration string `json:"duration"`
	// The bitrate in bps.
	Bitrate string `json:"bit_rate"`
	// The number of streams in file. Note that there might be audio, video, and data stream,
	// so if the streams is 2, it may indicate audio+video, video+data, or audio+data.
	Streams int32 `json:"nb_streams"`
	// The probe score, which indicates the confidence of the format detection.
	Score int32 `json:"probe_score"`
	// Whether has video stream.
	HasVideo bool `json:"has_video"`
	// Whether has audio stream.
	HasAudio bool `json:"has_audio"`
}

func (v *FFprobeFormat) String() string {
	return fmt.Sprintf("starttime=%v, duration=%v, bitrate=%v, streams=%v, score=%v, video=%v, audio=%v",
		v.Starttime, v.Duration, v.Bitrate, v.Streams, v.Score, v.HasVideo, v.HasAudio,
	)
}

// FFprobeVideo is the video object in ffprobe response.
type FFprobeVideo struct {
	// The codec type, should be video.
	CodecType string `json:"codec_type"`
	// The codec name, for example, h264, h265, vp6f, vp8, vp9, av1, or avs3.
	CodecName string `json:"codec_name"`
	// The codec profile, for example, High, Main, Baseline, or Constrained Baseline.
	Profile string `json:"profile"`
	// The width of video.
	Width int32 `json:"width"`
	// The height of video.
	Height int32 `json:"height"`
	// The pixel format, for example, yuv420p, yuv422p, yuv444p, yuv410p, yuv411p, yuvj420p,
	PixFormat string `json:"pix_fmt"`
	// The level of video.
	Level int32 `json:"level"`
	// The bitrate in bps.
	Bitrate string `json:"bit_rate"`
	// The start time in seconds.
	Starttime string `json:"start_time"`
	// The duration in seconds.
	Duration string `json:"duration"`
}

func (v *FFprobeVideo) String() string {
	return fmt.Sprintf("codec=%v, profile=%v, width=%v, height=%v, fmt=%v, level=%v, bitrate=%v",
		v.CodecName, v.Profile, v.Width, v.Height, v.PixFormat, v.Level, v.Bitrate,
	)
}

type M3u8VoDArtifact struct {
	// Number of ts files.
	NN int `json:"nn"`
	// The last update time.
	Update string `json:"update"`
	// The uuid of M3u8VoDArtifact, generated by worker, such as 3ECF0239-708C-42E4-96E1-5AE935C6E6A9
	UUID string `json:"uuid"`
	// The url of m3u8, generated by SRS, such as live/livestream/live.m3u8
	M3u8URL string `json:"m3u8_url"`

	// The vhost of stream, generated by SRS, such as video.test.com
	Vhost string `json:"vhost"`
	// The app of stream, generated by SRS, such as live
	App string `json:"app"`
	// The name of stream, generated by SRS, such as livestream
	Stream string `json:"stream"`

	// TODO: FIXME: It's a typo progress.
	// The Record is processing, use local m3u8 address to preview or download.
	Processing bool `json:"progress"`
	// The done time.
	Done string `json:"done"`
	// The ts files of this m3u8.
	Files []*TsFile `json:"files"`

	// For DVR only.
	// The COS bucket name.
	Bucket string `json:"bucket"`
	// The COS bucket region.
	Region string `json:"region"`

	// For VoD only.
	// The file ID generated by VoD commit.
	FileID   string `json:"fileId"`
	MediaURL string `json:"mediaUrl"`
	// The remux task of VoD.
	Definition uint64 `json:"definition"`
	TaskID     string `json:"taskId"`
}
func (v *M3u8VoDArtifact) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("uuid=%v, done=%v, update=%v, processing=%v, files=%v",
		v.UUID, v.Done, v.Update, v.Processing, len(v.Files),
	))
	if v.Bucket != "" {
		sb.WriteString(fmt.Sprintf(", bucket=%v", v.Bucket))
	}
	if v.Region != "" {
		sb.WriteString(fmt.Sprintf(", region=%v", v.Region))
	}
	if v.FileID != "" {
		sb.WriteString(fmt.Sprintf(", fileId=%v", v.FileID))
	}
	if v.MediaURL != "" {
		sb.WriteString(fmt.Sprintf(", mediaUrl=%v", v.MediaURL))
	}
	if v.Definition != 0 {
		sb.WriteString(fmt.Sprintf(", definition=%v", v.Definition))
	}
	if v.TaskID != "" {
		sb.WriteString(fmt.Sprintf(", taskId=%v", v.TaskID))
	}
	return sb.String()
}

// buildVodM3u8ForLocal go generate dynamic m3u8.
func buildVodM3u8ForLocal(
	ctx context.Context, tsFiles []*TsFile, useKey bool, prefix string,
) (
	contentType, m3u8Body string, duration float64, err error,
) {
	if len(tsFiles) == 0 {
		err = errors.Errorf("no files")
		return
	}

	for _, file := range tsFiles {
		duration += file.Duration
	}

	m3u8 := []string{
		"#EXTM3U",
		"#EXT-X-VERSION:3",
		"#EXT-X-ALLOW-CACHE:YES",
		"#EXT-X-PLAYLIST-TYPE:VOD",
		fmt.Sprintf("#EXT-X-TARGETDURATION:%v", math.Ceil(duration)),
		"#EXT-X-MEDIA-SEQUENCE:0",
	}
	for index, file := range tsFiles {
		// TODO: FIXME: Identify discontinuity by callback.
		if index < len(tsFiles)-2 {
			next := tsFiles[index+1]
			if file.SeqNo+1 != next.SeqNo {
				m3u8 = append(m3u8, "#EXT-X-DISCONTINUITY")
			}
		}

		m3u8 = append(m3u8, fmt.Sprintf("#EXTINF:%.2f, no desc", file.Duration))

		var tsURL string
		if useKey {
			tsURL = fmt.Sprintf("%v%v", prefix, file.Key)
		} else {
			tsURL = fmt.Sprintf("%v%v.ts", prefix, file.TsID)
		}
		m3u8 = append(m3u8, tsURL)
	}
	m3u8 = append(m3u8, "#EXT-X-ENDLIST")

	contentType = "application/vnd.apple.mpegurl"
	m3u8Body = strings.Join(m3u8, "\n")
	return
}