// Copyright (c) 2022-2024 Winlin
//
// SPDX-License-Identifier: MIT
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	// From ossrs.
	"github.com/google/uuid"
	"github.com/ossrs/go-oryx-lib/errors"
	"github.com/ossrs/go-oryx-lib/logger"
	"github.com/redis/go-redis/v9"
)

// The total segments in overlay HLS.
const maxFinishSegments = 9

type ProcessWorker struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// The global process task, only support one process task.
	task *ProcessTask

	// Use async goroutine to process on_hls messages.
	msgs chan *SrsOnHlsMessage

	// Got message from SRS, a new TS segment file is generated.
	tsfiles chan *SrsOnHlsObject

	// The worker which owns this object.
	detectWorker *DetectWorker
	UUID string `json:"uuid"`
	Stream string `json:"stream"`
}

func NewProcessWorker(d *DetectWorker) *ProcessWorker {
	v := &ProcessWorker{
		// Message on_hls.
		msgs: make(chan *SrsOnHlsMessage, 1024),
		// TS files.
		tsfiles: make(chan *SrsOnHlsObject, 1024),
		detectWorker: d,
	}
	v.task = NewProcessTask()
	v.task.processWorker = v
	return v
}

func (v *ProcessWorker) Initialize(d *DetectWorker) error {
	v.msgs = make(chan *SrsOnHlsMessage, 1024)
		// TS files.
	v.tsfiles = make(chan *SrsOnHlsObject, 1024)
	v.detectWorker = d
	v.task = NewProcessTask()
	v.task.processWorker = v

	return nil
}
func StartWorker(worker *ProcessWorker, ctx context.Context) (*ProcessWorker, error) {
	dir := fmt.Sprintf("process/%v", worker.Stream)

	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModeDir|os.FileMode(0755)); err != nil {
			return worker, errors.Wrapf(err, "create dir %v", dir)
		}
	}
	go worker.Start(ctx) // 워커 실행.
	logger.Tf(ctx, "Worker %s started", worker.Stream)
	return worker, nil
}

func (v *ProcessWorker)	hlsM3u8Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Format is /webvtt/:uuid.m3u8
	filename := r.URL.Path[len("/detect/hls/")+len(v.Stream)+1:]
	// Format is :uuid.m3u8
	if len(filename) == 0 {
		return errors.Errorf("invalid stream %v from %v of %v", v.Stream, filename, r.URL.Path)
	}

	var metaData []string
	var tsFiles []*TsFile
	segments := v.task.finishSegments()
	for _, segment := range segments {
		tsFiles = append(tsFiles, segment.TsFile)

		if b, err := json.Marshal(segment.BoundingBox); err != nil {
			return errors.Wrapf(err, "marshal %v", segment.BoundingBox)
		} else {
			metaData = append(metaData, string(b))
		}
	}
	contentType, m3u8Body, duration, err := buildLiveM3u8ForLocal(
		ctx, tsFiles, false, fmt.Sprintf("/detect/hls/%v/", v.Stream), metaData,
	)
	if err != nil {
		return errors.Wrapf(err, "build process m3u8 of %v", tsFiles)
	}

	w.Header().Set("Content-Type", contentType)
	w.Write([]byte(m3u8Body))
	logger.Tf(ctx, "process generate m3u8 ok, stream=%v, duration=%v", v.Stream, duration)
	return nil
}

func (v *ProcessWorker) hlsTsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Format is :uuid.ts
	filename := r.URL.Path[len("/detect/hls/")+len(v.Stream)+1:]
	fileBase := path.Base(filename)
	uuid := fileBase[:len(fileBase)-len(path.Ext(fileBase))]
	if len(uuid) == 0 {
		return errors.Errorf("invalid uuid %v from %v of %v", uuid, fileBase, r.URL.Path)
	}

	tsFilePath := path.Join("detect", fmt.Sprintf("%v.ts", uuid))
	if _, err := os.Stat(tsFilePath); err != nil {
		return errors.Wrapf(err, "no ts file %v", tsFilePath)
	}

	if tsFile, err := os.Open(tsFilePath); err != nil {
		return errors.Wrapf(err, "open file %v", tsFilePath)
	} else {
		defer tsFile.Close()
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		io.Copy(w, tsFile)
	}

	logger.Tf(ctx, "process server ts file ok, uuid=%v, ts=%v", uuid, tsFilePath)
	return nil
}

func (v *ProcessWorker) OnHlsTsMessage(ctx context.Context, msg *SrsOnHlsMessage) error {
	select {
	case <-ctx.Done():
	case v.msgs <- msg:
	}

	return nil
}
func (v *ProcessWorker) OnHlsTsObject(ctx context.Context, msg *SrsOnHlsObject) error {
	select {
	case <-ctx.Done():
	case v.tsfiles <- msg:
	}

	return nil
}

// task에 있는 hlstsmessage 처리
// ts파일을 임시 파일로 복사함.
// v.tsfiles에 notify
func (v *ProcessWorker) OnHlsTsMessageImpl(ctx context.Context, msg *SrsOnHlsMessage) error {
	// Ignore if not natch the task config.
	if !v.task.match(msg) {
		return nil
	}

	// Copy the ts file to temporary cache dir.
	tsid := fmt.Sprintf("%v-org-%v", msg.SeqNo, uuid.NewString())
	tsfile := path.Join("process", fmt.Sprintf("%v.ts", tsid))

	if err := exec.CommandContext(ctx, "cp", "-f", msg.File, tsfile).Run(); err != nil {
		return errors.Wrapf(err, "copy file %v to %v", msg.File, tsfile)
	}

	// Get the file size.
	stats, err := os.Stat(msg.File)
	if err != nil {
		return errors.Wrapf(err, "stat file %v", msg.File)
	}

	// Create a local ts file object.
	tsFile := &TsFile{
		TsID:     tsid,
		URL:      msg.URL,
		SeqNo:    msg.SeqNo,
		Duration: msg.Duration,
		Size:     uint64(stats.Size()),
		File:     tsfile,
	}

	// Notify worker asynchronously.
	// TODO: FIXME: Should cleanup the temporary file when restart.
	go func() {
		select {
		case <-ctx.Done():
		case v.tsfiles <- &SrsOnHlsObject{Msg: msg, TsFile: tsFile}:
		}
	}()
	return nil
}

func (v *ProcessWorker) Close() error {
	if v.cancel != nil {
		v.cancel()
	}

	v.wg.Wait()
	return nil
}

func (v *ProcessWorker) Start(ctx context.Context) error {
	wg := &v.wg

	ctx, cancel := context.WithCancel(ctx)
	v.cancel = cancel

	ctx = logger.WithContext(ctx)
	logger.Tf(ctx, "process start a worker")

	// Redis에서 아직 못한 일감 받아와서 v.task에 옮겨놓기
	// if objs, err := rdb.HGetAll(ctx, PROCESS_TASK).Result(); err != nil && err != redis.Nil {
	// 	return errors.Wrapf(err, "hgetall %v", PROCESS_TASK)
	// } else if len(objs) != 1 {
	// 	// Only support one task right now.
	// 	if err = rdb.Del(ctx, PROCESS_TASK).Err(); err != nil && err != redis.Nil {
	// 		return errors.Wrapf(err, "del %v", PROCESS_TASK)
	// 	}
	// } else {
	// 	for uuid, obj := range objs {
	// 		logger.Tf(ctx, "Load task %v object %v", uuid, obj)

	// 		if err = json.Unmarshal([]byte(obj), v.task); err != nil {
	// 			return errors.Wrapf(err, "unmarshal %v %v", uuid, obj)
	// 		}

	// 		break
	// 	}
	// }

	// 일 시작
	wg.Add(1)
	go func() {
		defer wg.Done()

		task := v.task
		for ctx.Err() == nil {
			var duration time.Duration
			if err := task.Run(ctx); err != nil {
				logger.Wf(ctx, "process: run task %v err %+v", task.String(), err)
				duration = 10 * time.Second
			} else {
				duration = 3 * time.Second
			}

			select {
			case <-ctx.Done():
			case <-time.After(duration):
			}
		}
	}()

	// ts file 임시 파일로 복사
	wg.Add(1)
	go func() {
		defer wg.Done()

		for ctx.Err() == nil {
			select {
			case <-ctx.Done():
			case msg := <-v.msgs:
				if err := v.OnHlsTsMessageImpl(ctx, msg); err != nil {
					logger.Wf(ctx, "process: handle on hls message %v err %+v", msg.String(), err)
				}
			}
		}
	}()

	// 복사한 이후 enqueue
	wg.Add(1)
	go func() {
		defer wg.Done()

		task := v.task
		for ctx.Err() == nil {
			select {
			case <-ctx.Done():
			case msg := <-v.tsfiles:
				if err := task.OnTsSegment(ctx, msg); err != nil {
					logger.Wf(ctx, "process: task %v on hls ts message %v err %+v", task.String(), msg.String(), err)
				}
			}
		}
	}()

	// ffmpeg로 이미지로 변환
	wg.Add(1)
	go func() {
		defer wg.Done()

		task := v.task
		for ctx.Err() == nil {
			var duration time.Duration
			if err := task.DriveLiveQueue(ctx); err != nil {
				logger.Wf(ctx, "process: task %v drive live queue err %+v", task.String(), err)
				duration = 10 * time.Second
			} else {
				duration = 200 * time.Millisecond
			}

			select {
			case <-ctx.Done():
			case <-time.After(duration):
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		task := v.task
		for ctx.Err() == nil {
			var duration time.Duration
			if err := task.DriveDetectQueue(ctx); err != nil {
				logger.Wf(ctx, "process: task %v drive asr queue err %+v", task.String(), err)
				duration = 10 * time.Second
			} else {
				duration = 200 * time.Millisecond
			}

			select {
			case <-ctx.Done():
			case <-time.After(duration):
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		task := v.task
		for ctx.Err() == nil {
			var duration time.Duration
			if err := task.DriveFinishQueue(ctx); err != nil {
				logger.Wf(ctx, "process: task %v drive overlay queue err %+v", task.String(), err)
				duration = 10 * time.Second
			} else {
				duration = 200 * time.Millisecond
			}

			select {
			case <-ctx.Done():
			case <-time.After(duration):
			}
		}
	}()

	return nil
}

type ProcessDetectSegment struct {
	ID    int     `json:"id,omitempty"`
	Seek  int     `json:"seek,omitempty"`
	Start float64 `json:"start,omitempty"`
	End   float64 `json:"end,omitempty"`
	Text  string  `json:"text,omitempty"`
}

type ProcessDetectResult struct {
	BBox []float64  `json:"bbox,omitempty"`
	Score float64 `json:"score,omitempty"`
	Category float64 `json:"category_id,omitempty"`
	ImageId string `json:"image_id,omitempty"`
	// The segments of the text.
	Segments []ProcessDetectSegment `json:"segments,omitempty"`
}

func (v ProcessDetectResult) String() string {
	return fmt.Sprintf("label=%v,x=%v,y=%v,width=%v,height=%v,score=%v,segments=%v",
		v.Category, v.BBox, v.Score, len(v.Segments), 
	)
}
type ProcessSegment struct {
	// The SRS callback message msg.
	Msg *SrsOnHlsMessage `json:"msg,omitempty"`
	// The original source TS file.
	TsFile *TsFile `json:"tsfile,omitempty"`
	// The extracted image file.
	ImageFile *TsFile `json:"image,omitempty"`

	BoundingBox []ProcessDetectResult `json:"bounding,omitempty"`
	// The starttime for live stream to adjust the srt.
	StreamStarttime time.Duration `json:"sst,omitempty"`
	// The generated SRT file from ASR result.
	SrtFile string `json:"srt,omitempty"`

	// The cost to transcode the TS file to image file.
	CostExtractImage time.Duration `json:"eic,omitempty"`
	// The cost to do process, converting speech to text.
	CostProcess time.Duration `json:"prc,omitempty"`
	// The cost to callback the process result.
	CostCallback time.Duration `json:"olc,omitempty"`
}

func (v ProcessSegment) String() string {
	var sb strings.Builder
	if v.Msg != nil {
		sb.WriteString(fmt.Sprintf("msg=%v, ", v.Msg.String()))
	}
	if v.TsFile != nil {
		sb.WriteString(fmt.Sprintf("ts=%v, ", v.TsFile.String()))
	}
	if v.ImageFile != nil {
		sb.WriteString(fmt.Sprintf("audio=%v, ", v.ImageFile.String()))
		sb.WriteString(fmt.Sprintf("eac=%v, ", v.CostExtractImage))
	}
	if v.BoundingBox != nil {
		sb.WriteString(fmt.Sprintf("asr=%v, ", v.BoundingBox))
		sb.WriteString(fmt.Sprintf("asrc=%v, ", v.CostProcess))
	}
	return sb.String()
}

func (v *ProcessSegment) Dispose() error {
	// Remove the original ts file.
	if v.TsFile != nil {
		if _, err := os.Stat(v.TsFile.File); err == nil {
			os.Remove(v.TsFile.File)
		}
	}

	// Remove the pure audio mp4 file.
	if v.ImageFile != nil {
		if _, err := os.Stat(v.ImageFile.File); err == nil {
			os.Remove(v.ImageFile.File)
		}
	}

	// Remove the SRT file.
	if v.SrtFile != "" {
		if _, err := os.Stat(v.SrtFile); err == nil {
			os.Remove(v.SrtFile)
		}
	}

	return nil
}

type ProcessQueue struct {
	// The process segments in the queue.
	Segments []*ProcessSegment `json:"segments,omitempty"`

	// To protect the queue.
	lock sync.Mutex
}

func NewProcessQueue() *ProcessQueue {
	return &ProcessQueue{}
}

func (v *ProcessQueue) String() string {
	return fmt.Sprintf("segments=%v", len(v.Segments))
}

func (v *ProcessQueue) count() int {
	v.lock.Lock()
	defer v.lock.Unlock()

	return len(v.Segments)
}

func (v *ProcessQueue) enqueue(segment *ProcessSegment) {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.Segments = append(v.Segments, segment)
}

func (v *ProcessQueue) first() *ProcessSegment {
	v.lock.Lock()
	defer v.lock.Unlock()

	if len(v.Segments) == 0 {
		return nil
	}

	return v.Segments[0]
}

func (v *ProcessQueue) dequeue(segment *ProcessSegment) {
	v.lock.Lock()
	defer v.lock.Unlock()

	for i, s := range v.Segments {
		if s == segment {
			v.Segments = append(v.Segments[:i], v.Segments[i+1:]...)
			return
		}
	}
}

func (v *ProcessQueue) reset(ctx context.Context) error {
	var segments []*ProcessSegment

	func() {
		v.lock.Lock()
		defer v.lock.Unlock()

		segments = v.Segments
		v.Segments = nil
	}()

	for _, segment := range segments {
		segment.Dispose()
	}

	return nil
}

type ProcessTask struct {
	// The ID for task.
	UUID string `json:"uuid,omitempty"`

	// The input url.
	Input string `json:"input,omitempty"`
	// The input stream object, select the active stream.
	inputStream *SrsStream

	// The live queue for the current task. HLS TS segments are copied to the process
	// directory, then a segment is created and added to the live queue for the process
	// task to process, to convert to pure audio mp4 file.
	LiveQueue *ProcessQueue `json:"live,omitempty"`
	// The ASR (Automatic Speech Recognition) queue for the current task. When a pure audio
	// MP4 file is generated, the segment is added to the ASR queue, which then requests the
	// AI server to convert the audio from the MP4 file into text.
	DetectQueue *ProcessQueue `json:"asr,omitempty"`
	// The fix queue for the current task. It allows users to manually fix and correct the
	// ASR-generated text. The overlay task won't start util user fix the ASR text.
	FixQueue *ProcessQueue `json:"fix,omitempty"`
	// The overlay queue for the current task. It involves drawing ASR (Automatic Speech
	// Recognition) text onto the video and encoding it into a new video file.
	FinishQueue *ProcessQueue `json:"overlay,omitempty"`

	// The previous ASR (Automatic Speech Recognition) text, which serves as a prompt for
	// generating the next one. AI services may use this previous ASR text as a prompt to
	// produce more accurate and robust subsequent ASR text.
	PreviousDetectText string `json:"pat,omitempty"`

	// The signal to persistence task.
	signalPersistence chan bool
	// The signal to change the active stream for task.
	signalNewStream chan *SrsStream

	// The process worker.
	processWorker *ProcessWorker

	// The context for current task.
	cancel context.CancelFunc

	// To protect the common fields.
	lock sync.Mutex
}

func NewProcessTask() *ProcessTask {
	return &ProcessTask{
		// Generate a UUID for task.
		UUID: uuid.NewString(),
		// The live queue for current task.
		LiveQueue: NewProcessQueue(),
		// The asr queue for current task.
		DetectQueue: NewProcessQueue(),
		// The fix queue for current task.
		FixQueue: NewProcessQueue(),
		// The overlay queue for current task.
		FinishQueue: NewProcessQueue(),
		// Create persistence signal.
		signalPersistence: make(chan bool, 1),
		// Create new stream signal.
		signalNewStream: make(chan *SrsStream, 1),
	}
}

func (v *ProcessTask) String() string {
	return fmt.Sprintf("uuid=%v, live=%v, asr=%v, fix=%v, pat=%v, overlay=%v",
		v.UUID, v.LiveQueue.String(), v.DetectQueue.String(), v.FixQueue.String(), v.PreviousDetectText,
		v.FinishQueue.String(),
	)
}

func (v *ProcessTask) Run(ctx context.Context) error {
	ctx = logger.WithContext(ctx)
	logger.Tf(ctx, "process run task %v", v.String())

	pfn := func(ctx context.Context) error {
		// Start process task.
		if err := v.doProcess(ctx); err != nil {
			return errors.Wrapf(err, "do process")
		}

		return nil
	}

	for ctx.Err() == nil {
		if err := pfn(ctx); err != nil {
			logger.Wf(ctx, "ignore %v err %+v", v.String(), err)

			select {
			case <-ctx.Done():
			case <-time.After(3500 * time.Millisecond):
			}
			continue
		}

		select {
		case <-ctx.Done():
		case <-time.After(300 * time.Millisecond):
		}
	}

	return nil
}

func (v *ProcessTask) doProcess(ctx context.Context) error {
	// Create context for current task.
	parentCtx := ctx
	ctx, v.cancel = context.WithCancel(ctx)

	// Main loop, process signals from user or system.
	for ctx.Err() == nil {
		select {
		case <-parentCtx.Done():
		case <-ctx.Done():
		case <-v.signalPersistence:
			if err := v.saveTask(ctx); err != nil {
				return errors.Wrapf(err, "save task %v", v.String())
			}
		}
	}

	return nil
}

// 처리완료 한 ts segment Live 큐에 넣기
func (v *ProcessTask) OnTsSegment(ctx context.Context, msg *SrsOnHlsObject) error {
	func() {
		// We must not update the queue, when persistence goroutine is working.
		v.lock.Lock()
		v.lock.Unlock()

		v.LiveQueue.enqueue(&ProcessSegment{
			Msg:    msg.Msg,
			TsFile: msg.TsFile,
		})
	}()

	// Notify the main loop to persistent current task.
	v.notifyPersistence(ctx)
	return nil
}

func (v *ProcessTask) DriveLiveQueue(ctx context.Context) error {
	// Ignore if not enough segments.
	if v.LiveQueue.count() <= 0 {
		return nil
	}

	segment := v.LiveQueue.first()
	starttime := time.Now()

	// Remove segment if file not exists.
	if _, err := os.Stat(segment.TsFile.File); err != nil && os.IsNotExist(err) {
		func() {
			v.lock.Lock()
			defer v.lock.Unlock()
			v.LiveQueue.dequeue(segment)
		}()

		segment.Dispose()
		logger.Tf(ctx, "process: remove not exist ts segment %v", segment.String())
		return nil
	}

	// Wait if ASR queue is full.
	if v.DetectQueue.count() >= maxFinishSegments+1 {
		return nil
	}

	// Transcode to image file, such as jpg.
	imageFile := &TsFile{
		TsID:     fmt.Sprintf("%v/%v-image-%v", v.processWorker.Stream, segment.TsFile.SeqNo, uuid.NewString()),
		URL:      segment.TsFile.URL,
		SeqNo:    segment.TsFile.SeqNo,
		Duration: segment.TsFile.Duration,
	}
	imageFile.File = path.Join("process", fmt.Sprintf("%v.jpg", imageFile.TsID))

	// TODO: FIXME: We should generate a set of images and use the best one.
	args := []string{
		"-i", segment.TsFile.File,
		"-frames:v", "1", "-q:v", "10",
		"-vf", "scale=640:640",
		"-y", imageFile.File,
	}
	if err := exec.CommandContext(ctx, "ffmpeg", args...).Run(); err != nil {
		return errors.Wrapf(err, "transcode %v", args)
	}

	// Update the size of image file.
	stats, err := os.Stat(imageFile.File)
	if err != nil {
		// TODO: FIXME: Cleanup the failed file.
		return errors.Wrapf(err, "stat file %v", imageFile.File)
	}
	imageFile.Size = uint64(stats.Size())

	// Dequeue the segment from live queue and attach to asr queue.
	func() {
		v.lock.Lock()
		defer v.lock.Unlock()

		v.LiveQueue.dequeue(segment)
		segment.ImageFile = imageFile
		segment.CostExtractImage = time.Since(starttime)
		v.DetectQueue.enqueue(segment)
	}()
	logger.Tf(ctx, "process: extract image %v to %v, size=%v, cost=%v",
		segment.TsFile.File, imageFile.File, imageFile.Size, segment.CostExtractImage)

	// Notify the main loop to persistent current task.
	v.notifyPersistence(ctx)
	return nil
}

func (v *ProcessTask) DriveDetectQueue(ctx context.Context) error {
	// Ignore if not enough segments.
	if v.DetectQueue.count() <= 0 {
		return nil
	}

	segment := v.DetectQueue.first()
	starttime := time.Now()

	// Remove segment if file not exists.
	if _, err := os.Stat(segment.ImageFile.File); err != nil && os.IsNotExist(err) {
		func() {
			v.lock.Lock()
			defer v.lock.Unlock()
			v.DetectQueue.dequeue(segment)
		}()
		segment.Dispose()
		logger.Tf(ctx, "process: remove not exist audio segment %v", segment.String())
		return nil
	}

	var imageData string
	if data, err := os.ReadFile(segment.ImageFile.File); err != nil {
		return errors.Wrapf(err, "read image from %v", segment.ImageFile.File)
	} else {
		imageData = base64.StdEncoding.EncodeToString(data)
	}
	
	err := postImageBase64(ctx, "http://121.78.254.27:10080/ai", imageData, &segment.BoundingBox);
	
	if err != nil {
		return errors.Wrapf(err, "post image %v (%v)", segment.ImageFile.File, len(imageData))
	}

	// Discover the starttime of the segment.
	stdout, err := exec.CommandContext(ctx, "ffprobe",
		"-show_error", "-show_private_data", "-v", "quiet", "-find_stream_info", "-print_format", "json",
		"-show_format", "-show_streams", segment.TsFile.File,
	).Output()
	if err != nil {
		return errors.Wrapf(err, "probe %v", segment.TsFile.File)
	}

	format := struct {
		Format FFprobeFormat `json:"format"`
	}{}
	if err = json.Unmarshal([]byte(stdout), &format); err != nil {
		return errors.Wrapf(err, "parse format %v", stdout)
	}

	if stv, err := strconv.ParseFloat(format.Format.Starttime, 10); err == nil {
		segment.StreamStarttime = time.Duration(stv * float64(time.Second))
	}

	// Dequeue the segment from asr queue and attach to correct queue.
	func() {
		v.lock.Lock()
		defer v.lock.Unlock()
		v.DetectQueue.dequeue(segment)
	}()
	// segment.BoundingBox = &ProcessDetectResult{
	// }

	// v.PreviousDetectText = resp.Text
	segment.CostProcess = time.Since(starttime)
	func() {
		v.lock.Lock()
		defer v.lock.Unlock()
		v.FinishQueue.enqueue(segment)
	}()
	logger.Tf(ctx, "process: detect image=%v, cost=%v",
		segment.ImageFile.File, segment.CostProcess)

	// Notify the main loop to persistent current task.
	v.notifyPersistence(ctx)
	return nil
}

func (v *ProcessTask) DriveFinishQueue(ctx context.Context) error {
	// Ignore if not enough segments.
	if v.FinishQueue.count() <= maxFinishSegments {
		select {
		case <-ctx.Done():
		case <-time.After(1 * time.Second):
		}
		return nil
	}

	// Cleanup the old segments.
	segment := v.FinishQueue.first()
	func() {
		v.lock.Lock()
		defer v.lock.Unlock()
		v.FinishQueue.dequeue(segment)
	}()
	logger.Tf(ctx, "dispose %v", segment.TsFile.File)
	defer segment.Dispose()

	// Notify the main loop to persistent current task.
	v.notifyPersistence(ctx)
	return nil
}

// TODO: FIXME: Should restart task when stream unpublish.
func (v *ProcessTask) restart(ctx context.Context) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.cancel != nil {
		v.cancel()
	}

	return nil
}

func (v *ProcessTask) reset(ctx context.Context) error {
	if err := func() error {
		v.lock.Lock()
		defer v.lock.Unlock()

		// Reset all queues.
		v.LiveQueue.reset(ctx)
		v.DetectQueue.reset(ctx)
		v.FixQueue.reset(ctx)
		v.FinishQueue.reset(ctx)

		// Reset all states.
		v.Input = ""
		v.PreviousDetectText = ""

		// Remove previous task from redis.
		if err := rdb.HDel(ctx, PROCESS_TASK, v.UUID).Err(); err != nil && err != redis.Nil {
			return errors.Wrapf(err, "hdel %v %v", PROCESS_TASK, v.UUID)
		}

		// Regenerate new UUID.
		v.UUID = uuid.NewString()

		return nil
	}(); err != nil {
		return errors.Wrapf(err, "reset task")
	}

	// Notify the main loop to persistent current task.
	v.notifyPersistence(ctx)

	// Wait for task to persistence and avoid to reset very fast.
	select {
	case <-ctx.Done():
	case <-time.After(1 * time.Second):
	}
	return nil
}

func (v *ProcessTask) match(msg *SrsOnHlsMessage) bool {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.inputStream == nil {
		return false
	}

	if v.inputStream.App != msg.App || v.inputStream.Stream != msg.Stream {
		return false
	}

	return true
}

func (v *ProcessTask) liveSegments() []*ProcessSegment {
	v.lock.Lock()
	defer v.lock.Unlock()

	return v.LiveQueue.Segments[:]
}

func (v *ProcessTask) detectSegments() []*ProcessSegment {
	v.lock.Lock()
	defer v.lock.Unlock()

	return v.DetectQueue.Segments[:]
}

func (v *ProcessTask) finishSegments() []*ProcessSegment {
	v.lock.Lock()
	defer v.lock.Unlock()

	return v.FinishQueue.Segments[:]
}

func (v *ProcessTask) notifyPersistence(ctx context.Context) {
	select {
	case <-ctx.Done():
	case v.signalPersistence <- true:
	default:
	}
}

func (v *ProcessTask) saveTask(ctx context.Context) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	starttime := time.Now()

	if b, err := json.Marshal(v); err != nil {
		return errors.Wrapf(err, "marshal %v", v.String())
	} else if err = rdb.HSet(ctx, PROCESS_TASK, v.UUID, string(b)).Err(); err != nil && err != redis.Nil {
		return errors.Wrapf(err, "hset %v %v %v", PROCESS_TASK, v.UUID, string(b))
	}

	logger.Tf(ctx, "process persistence ok, cost=%v", time.Since(starttime))

	return nil
}