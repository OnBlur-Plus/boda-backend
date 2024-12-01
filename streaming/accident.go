// Copyright (c) 2022-2024 Winlin
//
// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	// From ossrs.
	"github.com/google/uuid"
	"github.com/ossrs/go-oryx-lib/errors"
	ohttp "github.com/ossrs/go-oryx-lib/http"
	"github.com/ossrs/go-oryx-lib/logger"
	"github.com/redis/go-redis/v9"
)
var accidentWorker *AccidentWorker

var categories map[int]string

type AccidentWorker struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Use async goroutine to process on_hls messages.
	msgs chan *AccidentSegmentMsg
	// Got message from SRS, a new TS segment file is generated.
	tsfiles chan *AccidentSegment

	streams sync.Map
}
type AccidentSegmentMsg struct {
	DetectResult *ProcessDetectResult
	TsFile *TsFile
	inputStream *SrsStream
}
type AccidentSegment struct {
	DetectResult *ProcessDetectResult
	TsFile *TsFile
	inputStream *SrsStream
}
func (v *AccidentSegment) String() string {
	return fmt.Sprintf("msg(%v), ts(%v)", v.DetectResult.String(), v.TsFile.String())
}
func NewAccidentWorker() *AccidentWorker {
	categories = map[int]string{
		2: "NON_SAFETY_VEST",
		1: "NON_SAFETY_HELMET",
		7: "FALL",
		8: "USE_PHONE_WHILE_WORKING",
		9: "SOS_REQUEST",
	}

	v := &AccidentWorker{
		// Message on_hls.
		tsfiles: make(chan *AccidentSegment, 1024),
		msgs: make(chan *AccidentSegmentMsg, 1024),
	}
	return v
}
func (v *AccidentWorker) Handle(ctx context.Context, handler *http.ServeMux) error {
	mp4Handler := func(w http.ResponseWriter, r *http.Request) error {
		// Format is :uuid/index.mp4
		filename := r.URL.Path[len("/accident/hls/"):]
		uuid := path.Dir(filename)
		if len(uuid) == 0 {
			return errors.Errorf("invalid uuid %v from %v of %v", uuid, filename, r.URL.Path)
		}

		mp4FilePath := path.Join("accident", uuid, "index.mp4")
		stats, err := os.Stat(mp4FilePath)
		if err != nil {
			return errors.Wrapf(err, "no mp4 file %v", mp4FilePath)
		}

		mp4File, err := os.Open(mp4FilePath)
		if err != nil {
			return errors.Wrapf(err, "open file %v", mp4FilePath)
		}
		defer mp4File.Close()

		// No range request.
		rangeHeader := r.Header.Get("Range")
		if rangeHeader == "" {
			w.Header().Set("Content-Type", "video/mp4")
			io.Copy(w, mp4File)
			logger.T(ctx, "accident serve full mp4=%v", mp4FilePath)
			return nil
		}

		// Support range request.
		var start, end int64
		fmt.Sscanf(rangeHeader, "bytes=%u-%u", &start, &end)
		if end == 0 {
			end = stats.Size() - 1
		}

		if _, err := mp4File.Seek(start, io.SeekStart); err != nil {
			return errors.Wrapf(err, "seek to %v of %v", start, mp4FilePath)
		}

		w.Header().Set("ccept-Ranges", "bytes")
		w.Header().Set("Content-Length", fmt.Sprintf("%v", end+1-start))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %v-%v/%v", start, end, stats.Size()))

		w.WriteHeader(http.StatusPartialContent)
		w.Header().Set("Content-Type", "video/mp4")
		io.CopyN(w, mp4File, end+1-start)

		logger.Tf(ctx, "accident serve partial ok, uuid=%v, mp4=%v", uuid, mp4FilePath)
		return nil
	}

	ep := "/accident/hls/"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			if strings.HasSuffix(r.URL.Path, ".mp4") {
				return mp4Handler(w, r)
			}

			return errors.Errorf("invalid handler for %v", r.URL.Path)
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	return nil
}
func (v *AccidentWorker) OnAccidentAdded(ctx context.Context, result *ProcessDetectResult, _TsFile *TsFile, stream *SrsStream) error {
	select {
	case <-ctx.Done():
	case v.msgs <- &AccidentSegmentMsg{
		DetectResult: result,
		TsFile: _TsFile,
		inputStream: stream,
	}:
	}

	return nil
}
func (v *AccidentWorker) OnAccidentAddedImpl(ctx context.Context, msg *AccidentSegmentMsg) error {
	// Copy the ts file to temporary cache dir.
	tsid := uuid.NewString()
	tsfile := path.Join("accident", fmt.Sprintf("%v.ts", tsid))

	logger.Tf(ctx, "On AccidentAdded %v", msg.TsFile.File)
	// Always use execFile when params contains user inputs, see https://auth0.com/blog/preventing-command-injection-attacks-in-node-js-apps/
	// Note that should never use fs.copyFileSync(file, tsfile, fs.constants.COPYFILE_FICLONE_FORCE) which fails in macOS.
	if err := exec.CommandContext(ctx, "cp", "-f", msg.TsFile.File, tsfile).Run(); err != nil {
		return errors.Wrapf(err, "copy file %v to %v", msg.TsFile.File, tsfile)
	}

	// Get the file size.
	stats, err := os.Stat(msg.TsFile.File)
	if err != nil {
		return errors.Wrapf(err, "stat file %v", msg.TsFile.File)
	}

	// Create a local ts file object.
	tsFile := &TsFile{
		TsID:     tsid,
		URL:      msg.TsFile.URL,
		SeqNo:    msg.TsFile.SeqNo,
		Duration: msg.TsFile.Duration,
		Size:     uint64(stats.Size()),
		File:     tsfile,
	}
	logger.Tf(ctx, "TsFile generated %v", tsFile.File)
	select {
	case <-ctx.Done():
	case v.tsfiles <- &AccidentSegment {
		TsFile: tsFile,
		DetectResult: msg.DetectResult,
		inputStream: msg.inputStream,
	}:
	}

	return nil
}

func (v *AccidentWorker) QueryTask(uuid string) *AccidentM3u8Stream {
	var target *AccidentM3u8Stream
	v.streams.Range(func(key, value interface{}) bool {
		if task := value.(*AccidentM3u8Stream); task.UUID == uuid {
			target = task
			return false
		}
		return true
	})
	return target
}

func (v *AccidentWorker) Close() error {
	if v.cancel != nil {
		v.cancel()
	}

	v.wg.Wait()
	return nil
}

func (v *AccidentWorker) Start(ctx context.Context) error {
	wg := &v.wg

	ctx, cancel := context.WithCancel(ctx)
	v.cancel = cancel

	ctx = logger.WithContext(ctx)
	logger.Tf(ctx, "Accident: start a worker")

	// Create M3u8 object from message.
	buildM3u8Object := func(ctx context.Context, msg *AccidentSegment) error {
		// If glob filters are empty, ignore it, and record all streams.

		// Load stream local object.
		var m3u8LocalObj *AccidentM3u8Stream
		var freshObject bool
		M3u8URL := fmt.Sprintf("%v/%v",msg.inputStream.Stream,msg.DetectResult.Category)
		if obj, loaded := v.streams.LoadOrStore(M3u8URL, &AccidentM3u8Stream{
			M3u8URL: M3u8URL, UUID: uuid.NewString(), AccidentWorker: v,
			Stream: msg.inputStream.Stream,
			Category: msg.DetectResult.Category,
		}); true {
			m3u8LocalObj, freshObject = obj.(*AccidentM3u8Stream), !loaded
		}

		// Initialize the fresh object.
		if freshObject {
			if err := m3u8LocalObj.Initialize(ctx, v); err != nil {
				return errors.Wrapf(err, "init %v", m3u8LocalObj.String())
			}
		}

		// Append new ts file to object.
		m3u8LocalObj.addMessage(ctx, msg)

		// Always save the object to redis, for reloading it when restart.
		if err := m3u8LocalObj.saveObject(ctx); err != nil {
			return errors.Wrapf(err, "save %v", m3u8LocalObj.String())
		}

		// Serve object if fresh one.
		if freshObject {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := m3u8LocalObj.Run(ctx); err != nil {
					logger.Wf(ctx, "serve m3u8 %v err %+v", m3u8LocalObj.String(), err)
				}
			}()
		}

		return nil
	}
	// 복사한 이후 enqueue
	wg.Add(1)
	go func() {
		defer wg.Done()

		for ctx.Err() == nil {
			select {
			case <-ctx.Done():
			case msg := <-v.msgs:
				if err := v.OnAccidentAddedImpl(ctx, msg); err != nil {
					logger.Wf(ctx, "process: task %v on hls ts message %v err %+v", msg.DetectResult.String(), err)
				}
			}
		}
	}()
	
	// Process all messages about HLS ts segments.
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-v.tsfiles:
				if err := buildM3u8Object(ctx, msg); err != nil {
					logger.Wf(ctx, "ignore msg %v ts %v err %+v", msg.DetectResult.String(), msg.TsFile.String(), err)
				}
			}
		}
	}()

	return nil
}

// AccidentM3u8Stream is the current active local object for a HLS stream.
// When recording done, it will generate a M3u8VoDArtifact, which is a HLS VoD object.
type AccidentM3u8Stream struct {
	// The url of m3u8, generated by SRS, such as live/livestream/live.m3u8
	M3u8URL string `json:"m3u8"`
	Stream string `json:"stream"`
	Category int `json:"category"`
	// The uuid of M3u8VoDObject, generated by worker, such as 3ECF0239-708C-42E4-96E1-5AE935C6E6A9
	UUID string `json:"uuid"`

	// Number of local files.
	NN int `json:"nn"`
	// The last update time.
	Update string `json:"update"`
	// The done time.
	Done string `json:"done"`
	// Whether task is set to expire by user.
	Expired bool `json:"expired"`

	AccidentId int `json:"accidentId"`

	// The ts files of this m3u8.
	Messages []*AccidentSegment `json:"msgs"`

	// The worker which owns this object.
	AccidentWorker *AccidentWorker
	// The artifact we're working for.
	artifact *M3u8VoDArtifact
	// To protect the fields.
	lock sync.Mutex
}

func (v AccidentM3u8Stream) String() string {
	return fmt.Sprintf("url=%v, uuid=%v, done=%v, update=%v, messages=%v, expired=%v",
		v.M3u8URL, v.UUID, v.Done, v.Update, len(v.Messages), v.Expired,
	)
}

func (v *AccidentM3u8Stream) deleteObject(ctx context.Context) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	if err := rdb.HDel(ctx, SRS_ACCIDENT_M3U8_WORKING, v.M3u8URL).Err(); err != nil && err != redis.Nil {
		return errors.Wrapf(err, "hdel %v %v", SRS_ACCIDENT_M3U8_WORKING, v.M3u8URL)
	}

	return nil
}

func (v *AccidentM3u8Stream) saveObject(ctx context.Context) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	if b, err := json.Marshal(v); err != nil {
		return errors.Wrapf(err, "marshal object")
	} else if err = rdb.HSet(ctx, SRS_ACCIDENT_M3U8_WORKING, v.M3u8URL, string(b)).Err(); err != nil && err != redis.Nil {
		return errors.Wrapf(err, "hset %v %v %v", SRS_ACCIDENT_M3U8_WORKING, v.M3u8URL, string(b))
	}
	return nil
}

func (v *AccidentM3u8Stream) saveArtifact(ctx context.Context, artifact *M3u8VoDArtifact) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	if b, err := json.Marshal(artifact); err != nil {
		return errors.Wrapf(err, "marshal %v", artifact.String())
	} else if err = rdb.HSet(ctx, SRS_ACCIDENT_M3U8_ARTIFACT, v.UUID, string(b)).Err(); err != nil && err != redis.Nil {
		return errors.Wrapf(err, "hset %v %v %v", SRS_ACCIDENT_M3U8_ARTIFACT, v.UUID, string(b))
	}
	return nil
}

func (v *AccidentM3u8Stream) updateArtifact(ctx context.Context, artifact *M3u8VoDArtifact, msg *AccidentSegment) {
	v.lock.Lock()
	defer v.lock.Unlock()

	artifact.Vhost = msg.inputStream.Vhost
	artifact.App = msg.inputStream.App
	artifact.Stream = msg.inputStream.Stream

	artifact.Files = append(artifact.Files, msg.TsFile)
	artifact.NN = len(artifact.Files)

	artifact.Update = time.Now().Format(time.RFC3339)
}

func (v *AccidentM3u8Stream) finishArtifact(ctx context.Context, artifact *M3u8VoDArtifact) {
	v.lock.Lock()
	defer v.lock.Unlock()

	artifact.Processing = false
	artifact.Update = time.Now().Format(time.RFC3339)
}

func (v *AccidentM3u8Stream) addMessage(ctx context.Context, msg *AccidentSegment) {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.Messages = append(v.Messages, msg)
	v.NN = len(v.Messages)
	v.Update = time.Now().Format(time.RFC3339)
}

func (v *AccidentM3u8Stream) copyMessages() []*AccidentSegment {
	v.lock.Lock()
	defer v.lock.Unlock()

	return append([]*AccidentSegment{}, v.Messages...)
}

func (v *AccidentM3u8Stream) removeMessage(msg *AccidentSegment) {
	v.lock.Lock()
	defer v.lock.Unlock()

	for index, m := range v.Messages {
		if m == msg {
			v.Messages = append(v.Messages[:index], v.Messages[index+1:]...)
			break
		}
	}

	v.NN = len(v.Messages)
	v.Update = time.Now().Format(time.RFC3339)
}

func (v *AccidentM3u8Stream) expired(ctx context.Context) bool {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.Expired {
		return true
	}

	update, err := time.Parse(time.RFC3339, v.Update)
	if err != nil {
		return true
	}

	duration := 30 * time.Second

	if update.Add(duration).Before(time.Now()) {
		return true
	}

	return false
}

// Initialize to load artifact. There is no simultaneously access, so no certFileLock is needed.
func (v *AccidentM3u8Stream) Initialize(ctx context.Context, r *AccidentWorker) error {
	v.AccidentWorker = r
	logger.Tf(ctx, "accident initialize url=%v, uuid=%v", v.M3u8URL, v.UUID)
	// Try to load artifact from redis. The final artifact is VoD HLS object.
	if value, err := rdb.HGet(ctx, SRS_ACCIDENT_M3U8_ARTIFACT, v.UUID).Result(); err != nil && err != redis.Nil {
		return errors.Wrapf(err, "hget %v %v", SRS_ACCIDENT_M3U8_ARTIFACT, v.UUID)
	} else if value != "" {
		artifact := &M3u8VoDArtifact{}
		if err = json.Unmarshal([]byte(value), artifact); err != nil {
			return errors.Wrapf(err, "unmarshal %v", value)
		} else {
			v.artifact = artifact
		}
	}

	// Create a artifact if new.
	if v.artifact == nil {
		v.artifact = &M3u8VoDArtifact{
			UUID:       v.UUID,
			M3u8URL:    v.M3u8URL,
			Processing: true,
			Update:     time.Now().Format(time.RFC3339),
		}
		if err := v.saveArtifact(ctx, v.artifact); err != nil {
			return errors.Wrapf(err, "save artifact %v", v.artifact.String())
		}
	}

	v.callbackBegin(ctx, &v.AccidentId)

	return nil
}

// Run to serve the current recording object.
func (v *AccidentM3u8Stream) Run(ctx context.Context) error {
	parentCtx := logger.WithContext(ctx)
	ctx, cancel := context.WithCancel(parentCtx)
	logger.Tf(ctx, "record run task %v", v.String())

	pfn := func() error {
		// Process message and remove it.
		msgs := v.copyMessages()
		for _, msg := range msgs {
			if err := v.serveMessage(ctx, msg); err != nil {
				logger.Wf(ctx, "ignore %v err %+v", msg.String(), err)
			}
		}

		// Refresh redis if got messages to serve.
		if len(msgs) > 0 {
			if err := v.saveObject(ctx); err != nil {
				return errors.Wrapf(err, "save object %v", v.String())
			}
		}

		// Ignore if still has messages to process.
		if len(v.Messages) > 0 {
			return nil
		}
		// Check whether expired.
		if !v.expired(ctx) {
			return nil
		}

		// Try to finish the object.
		if err := v.finishM3u8(ctx); err != nil {
			return errors.Wrapf(err, "finish m3u8")
		}

		// Now HLS is done
		logger.Tf(ctx, "Record is done, hls is %v, artifact is %v", v.String(), v.artifact.String())
		cancel()

		return nil
	}

	for ctx.Err() == nil {
		if err := pfn(); err != nil {
			logger.Wf(ctx, "ignore %v err %+v", v.String(), err)

			select {
			case <-ctx.Done():
			case <-time.After(10 * time.Second):
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

func (v *AccidentM3u8Stream) serveMessage(ctx context.Context, msg *AccidentSegment) error {
	// We always remove the msg from current object.
	defer v.removeMessage(msg)

	// Ignore file if not exists.
	if _, err := os.Stat(msg.TsFile.File); err != nil {
		return err
	}

	tsDir := path.Join("accident", v.UUID)
	key := path.Join(tsDir, fmt.Sprintf("%v.ts", msg.TsFile.TsID))
	msg.TsFile.Key = key

	if err := os.MkdirAll(tsDir, 0755); err != nil {
		return errors.Wrapf(err, "mkdir %v", tsDir)
	}

	if err := os.Rename(msg.TsFile.File, key); err != nil {
		return errors.Wrapf(err, "rename %v to %v", msg.TsFile.File, key)
	}

	// Update the metadata for m3u8.
	v.updateArtifact(ctx, v.artifact, msg)
	if err := v.saveArtifact(ctx, v.artifact); err != nil {
		return errors.Wrapf(err, "save artifact %v", v.artifact.String())
	}

	logger.Tf(ctx, "accident consume msg %v", msg.String())
	return nil
}

func (v *AccidentM3u8Stream) finishM3u8(ctx context.Context) error {
	parentCtx := logger.WithContext(ctx)
	contentType, m3u8Body, duration, err := buildVodM3u8ForLocal(ctx, v.artifact.Files, false, "")
	if err != nil {
		return errors.Wrapf(err, "build vod")
	}

	hls := path.Join("accident", v.UUID, "index.m3u8")
	if f, err := os.OpenFile(hls, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return errors.Wrapf(err, "open file %v", hls)
	} else {
		defer f.Close()
		if _, err = f.Write([]byte(m3u8Body)); err != nil {
			return errors.Wrapf(err, "write hls %v to %v", m3u8Body, hls)
		}
	}
	logger.Tf(ctx, "accident to %v ok, type=%v, duration=%v", hls, contentType, duration)

	mp4 := path.Join("accident", v.UUID, "index.mp4")
	if b, err := exec.CommandContext(ctx, "ffmpeg", "-i", hls, "-c", "copy", "-y", mp4).Output(); err != nil {
		return errors.Wrapf(err, "covert to mp4 %v err %v", mp4, string(b))
	}
	logger.Tf(ctx, "accident to %v ok", mp4)

	// Remove object from worker.
	v.AccidentWorker.streams.Delete(v.M3u8URL)
	
	if true {
		ctx := parentCtx

		if err := v.callbackEnd(ctx, mp4); err != nil {
			logger.Wf(ctx, "ignore task %v callback end err %+v", v.String(), err)
		}
	}

	// Update artifact after finally.
	v.finishArtifact(ctx, v.artifact)
	r0 := v.saveArtifact(ctx, v.artifact)
	r1 := v.deleteObject(ctx)
	logger.Tf(ctx, "record cleanup ok, r0=%v, r1=%v", r0, r1)

	// Do final cleanup, because new messages might arrive while converting to mp4, which takes a long time.
	files := v.copyMessages()
	for _, file := range files {
		r2 := os.Remove(file.TsFile.File)
		logger.Tf(ctx, "drop %v r2=%v", file.String(), r2)
	}

	return nil
}

func (v *AccidentM3u8Stream) callbackBegin(ctx context.Context, accidentId *int) error {
	logger.Tf(ctx, "callbackBegin called for url=%v", v.M3u8URL)
	pf := func(url string, requestBody interface{}) error {
		logger.Tf(ctx, "callbackBegin called for url=%v b=%v", v.M3u8URL, requestBody)
		b, err := json.Marshal(requestBody)
		if err != nil {
			return errors.Wrapf(err, "marshal req")
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
		if err != nil {
			return errors.Wrapf(err, "new request")
		}
	
		req.Header.Set("Content-Type", "application/json")
	
		var res *http.Response
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrapf(err, "http post")
		}
		defer res.Body.Close()
	
		if res.StatusCode != http.StatusCreated {
			return errors.Errorf("response status %v", res.StatusCode)
		}

		logger.Tf(ctx, "callbackBegin called for url=%v status=%v", url, res.StatusCode)

		b2, err := io.ReadAll(res.Body)
		if err != nil {
			return errors.Wrapf(err, "read body")
		}
		if err := json.Unmarshal(b2, &struct {
			AccidentId *int `json:"accidentId"`
		}{
			AccidentId: accidentId,
		}); err != nil {
			return errors.Wrapf(err, "unmarshal response")
		}
		logger.Tf(ctx, "callbackBegin called for url=%v accidentId=%v", url, accidentId)
		
	
		return nil
	}
	if Type, exists := categories[v.Category]; exists {
		
		err := pf("http://127.0.0.1:5000/accident", &struct {
			StreamKey string `json:"streamKey"`
			Type string `json:"type"`
		}{
			StreamKey: v.Stream,
			Type: Type,
		});
		if err != nil {
			return errors.Wrapf(err, "start Accident with %s", err)
		}
		return nil;
	}

	return errors.Errorf("CategoryId does not exist %v", v.Category)
}

func (v *AccidentM3u8Stream) callbackEnd(ctx context.Context, mp4File string) error {
	logger.Tf(ctx, "callbackbeginend called for url=%v", v.M3u8URL)
	pf := func(url string, requestBody interface{}) error {
		logger.Tf(ctx, "callbackbeginend called for url=%v b=%v", v.M3u8URL, requestBody)
		b, err := json.Marshal(requestBody)
		if err != nil {
			return errors.Wrapf(err, "marshal req")
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
		if err != nil {
			return errors.Wrapf(err, "new request")
		}
	
		req.Header.Set("Content-Type", "application/json")
	
		var res *http.Response
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrapf(err, "http post")
		}
		defer res.Body.Close()
	
		if res.StatusCode != http.StatusCreated {
			return errors.Errorf("response status %v", res.StatusCode)
		}
	
		return nil
	}
		err := pf("http://127.0.0.1:5000/accident/end", &struct {
			AccidentId int `json:"id"`
			MP4 string `json:"videoUrl"`
		}{
			AccidentId: v.AccidentId,
			MP4: mp4File,
		});
		if err != nil {
			return errors.Wrapf(err, "End Accident with %s", err)
		}
		return nil;
}