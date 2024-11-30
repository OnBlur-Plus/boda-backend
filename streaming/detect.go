// Copyright (c) 2022-2024 Winlin
//
// SPDX-License-Identifier: MIT
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"github.com/ossrs/go-oryx-lib/errors"
	ohttp "github.com/ossrs/go-oryx-lib/http"
	"github.com/ossrs/go-oryx-lib/logger"

	"github.com/google/uuid"
)

type RecordPostProcess string

var detectWorker *DetectWorker

type DetectWorker struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Use async goroutine to process on_hls messages.
	msgs chan *SrsOnHlsMessage

	// Got message from SRS, a new TS segment file is generated.
	tsfiles chan *SrsOnHlsObject

	// The streams we're detecting, key is m3u8 URL in string, value is m3u8 object *DetectM3u8Stream.
	workers sync.Map
}

func NewDetectWorker() *DetectWorker {
	return &DetectWorker{
		// Message on_hls.
		msgs: make(chan *SrsOnHlsMessage, 1024),
		// TS files.
		tsfiles: make(chan *SrsOnHlsObject, 1024),
	}
}

func (v *DetectWorker) Handle(ctx context.Context, handler *http.ServeMux) error {
	ep := "/detect/hls/"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			prefix := "/detect/hls/"
			filename := r.URL.Path[len(prefix):]
			// Format is :uuid/index.m3u8
			splits := strings.Split(filename, "/")
			if len(splits) < 2 {
				return errors.Errorf("invalid url %v", filename)
			}
			stream := splits[0]

			var processWorker *ProcessWorker

			obj, ok := v.workers.Load(stream);

			// Initialize the fresh object.
			if !ok {
				return errors.Errorf("invalid url %v", stream)
			}

			processWorker = obj.(*ProcessWorker)
			
			if strings.HasSuffix(r.URL.Path, ".m3u8") {
				return processWorker.hlsM3u8Handler(ctx, w, r)
			} else if strings.HasSuffix(r.URL.Path, ".ts") {
				return processWorker.hlsTsHandler(ctx, w, r)
			}

			return errors.Errorf("invalid handler for %v", r.URL.Path)
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	return nil
}
func (v *DetectWorker) OnHlsTsMessage(ctx context.Context, msg *SrsOnHlsMessage) error {
	select {
	case <-ctx.Done():
	case v.msgs <- msg:
	}

	return nil
}

func (v *DetectWorker) OnHlsTsMessageImpl(ctx context.Context, msg *SrsOnHlsMessage) error {
	// Copy the ts file to temporary cache dir.
	tsid := uuid.NewString()
	tsfile := path.Join("detect", fmt.Sprintf("%v.ts", tsid))

	// Always use execFile when params contains user inputs, see https://auth0.com/blog/preventing-command-injection-attacks-in-node-js-apps/
	// Note that should never use fs.copyFileSync(file, tsfile, fs.constants.COPYFILE_FICLONE_FORCE) which fails in macOS.
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
	go func() {
		select {
		case <-ctx.Done():
		case v.tsfiles <- &SrsOnHlsObject{Msg: msg, TsFile: tsFile}:
		}
	}()
	return nil
}

func (v *DetectWorker) Close() error {
	if v.cancel != nil {
		v.cancel()
	}
	v.wg.Wait()
	return nil
}

func (v *DetectWorker) QueryTask(uuid string) *ProcessWorker {
	var target *ProcessWorker
	v.workers.Range(func(key, value interface{}) bool {
		if task := value.(*ProcessWorker); task.UUID == uuid {
			target = task
			return false
		}
		return true
	})
	return target
}

func (v *DetectWorker) Start(ctx context.Context) error {
	wg := &v.wg

	ctx, cancel := context.WithCancel(ctx)
	v.cancel = cancel

	ctx = logger.WithContext(ctx)
	logger.Tf(ctx, "Record: start a worker")

	// Create M3u8 object from message.
	createWorker := func(ctx context.Context, msg *SrsOnHlsObject) error {
		// Load stream local object.
		var processWorker *ProcessWorker
		var freshObject bool
		if obj, loaded := v.workers.LoadOrStore(msg.Msg.Stream, &ProcessWorker{
			Stream: msg.Msg.Stream, UUID: uuid.NewString(), detectWorker: v,
		}); true {
			processWorker, freshObject = obj.(*ProcessWorker), !loaded
		}

		// Initialize the fresh object.
		if freshObject {
			if err := processWorker.Initialize(v); err != nil {
				return errors.Wrapf(err, "init process worker")
			}
		}

		logger.Wf(ctx, "process worker initialize2 %v, %v", msg.Msg.Stream, freshObject)

		if freshObject {
			if _, err := StartWorker(processWorker, ctx); err != nil {
				logger.Wf(ctx, "process worker error!!", msg.Msg.Stream, err)
			}
		}
		
		processWorker.OnHlsTsObject(ctx, msg)

		return nil
	}
		// Consume all on_hls messages.
	wg.Add(1)
	go func() {
		defer wg.Done()
	
		for ctx.Err() == nil {
			select {
			case <-ctx.Done():
			case msg := <-v.msgs:
				if err := v.OnHlsTsMessageImpl(ctx, msg); err != nil {
					logger.Wf(ctx, "transcript: handle on hls message %v err %+v", msg.String(), err)
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
			case tsflie := <-v.tsfiles:
				if err := createWorker(ctx, tsflie); err != nil {
					logger.Wf(ctx, "ignore msg %v ts %v err %+v", tsflie.Msg.String(), tsflie.TsFile.String(), err)
				}
			}
		}
	}()

	return nil
}