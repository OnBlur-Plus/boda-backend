// Copyright (c) 2022-2024 Winlin
//
// SPDX-License-Identifier: MIT
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ossrs/go-oryx-lib/errors"
	ohttp "github.com/ossrs/go-oryx-lib/http"
	"github.com/ossrs/go-oryx-lib/logger"
	"github.com/redis/go-redis/v9"
)

type SrsAction string

const (
	// The actions for SRS server and Oryx.
	// The publish action.
	SrsActionOnPublish SrsAction = "on_publish"
	// The unpublish action.
	SrsActionOnUnpublish = "on_unpublish"

	// The hls action, for SRS server only.
	SrsActionOnHls = "on_hls"

	// The on_record_begin action.
	SrsActionOnRecordBegin = "on_record_begin"
	// The on_record_end action.
	SrsActionOnRecordEnd = "on_record_end"

	// The on_ocr action.
	SrsActionOnOcr = "on_ocr"
)

func handleHooksService(ctx context.Context, handler *http.ServeMux) error {
	pf := func(url string, requestBody interface{}) error {
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


	ep := "/hooks/srs/verify"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				return errors.Wrapf(err, "read body")
			}
			// requestBody := string(b)

			var action SrsAction
			var streamObj SrsStream
			if err := json.Unmarshal(b, &struct {
				Action *SrsAction `json:"action"`
				*SrsStream
			}{
				Action: &action, SrsStream: &streamObj,
			}); err != nil {
				return errors.Wrapf(err, "json unmarshal %v", string(b))
			}

			if action == SrsActionOnPublish {
				if err := pf("http://127.0.0.1:5000/stream/verify", &struct {
					StreamKey string `json:"streamKey"`
				}{
					StreamKey: streamObj.Stream,
				}); err != nil {
					return errors.Wrapf(err, "verify with %s", streamObj.Stream)
				}
			}
			streamURL := streamObj.StreamURL()
			if action == SrsActionOnPublish {
				streamObj.Update = time.Now().Format(time.RFC3339)

				b, err := json.Marshal(&streamObj)
				if err != nil {
					return errors.Wrapf(err, "marshal json")
				} else if err = rdb.HSet(ctx, SRS_STREAM_ACTIVE, streamURL, string(b)).Err(); err != nil && err != redis.Nil {
					return errors.Wrapf(err, "hset %v %v %v", SRS_STREAM_ACTIVE, streamURL, string(b))
				}

				if err := rdb.HIncrBy(ctx, SRS_STAT_COUNTER, "publish", 1).Err(); err != nil && err != redis.Nil {
					return errors.Wrapf(err, "hincrby %v publish 1", SRS_STAT_COUNTER)
				}
			} else if action == SrsActionOnUnpublish {
				if err := rdb.HDel(ctx, SRS_STREAM_ACTIVE, streamURL).Err(); err != nil && err != redis.Nil {
					return errors.Wrapf(err, "hset %v %v", SRS_STREAM_ACTIVE, streamURL)
				}

				if err := pf("http://127.0.0.1:5000/stream/end", &struct {
					StreamKey string `json:"streamKey"`
				}{
					StreamKey: streamObj.Stream,
				}); err != nil {
					return errors.Wrapf(err, "unpublish with %s", streamObj.Stream)
				}
			} else if action == "on_play" {
				if err := rdb.HIncrBy(ctx, SRS_STAT_COUNTER, "play", 1).Err(); err != nil && err != redis.Nil {
					return errors.Wrapf(err, "hincrby %v play 1", SRS_STAT_COUNTER)
				}
			}

			ohttp.WriteData(ctx, w, r, nil)
			logger.Tf(ctx, "srs hooks ok, action=%v, %v",
				action, streamObj.String())
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})
	if err := handleOnHls(ctx, handler); err != nil {
		return errors.Wrapf(err, "handle hooks")
	}

	return nil
}

func handleOnHls(ctx context.Context, handler *http.ServeMux) error {
	// TODO: FIXME: Fixed token.
	// See https://github.com/ossrs/srs/wiki/v4_EN_HTTPCallback
	ep := "/hooks/srs/hls"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		if err := func() error {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				return errors.Wrapf(err, "read body")
			}

			var msg SrsOnHlsMessage
			if err := json.Unmarshal(b, &msg); err != nil {
				return errors.Wrapf(err, "json unmarshal %v", string(b))
			}
			if msg.Action != SrsActionOnHls {
				return errors.Errorf("invalid action=%v", msg.Action)
			}
			if _, err := os.Stat(msg.File); err != nil {
				logger.Ef(ctx, conf.Pwd)
				return errors.Wrapf(err, "invalid ts file %v", msg.File)
			}
			logger.Tf(ctx, "on_hls ok, %v", string(b))

			// Handle TS file by Record task if enabled.
			if err = detectWorker.OnHlsTsMessage(ctx, &msg); err != nil {
				return errors.Wrapf(err, "feed %v", msg.String())
			}
			logger.Tf(ctx, "record %v", msg.String())
			
			ohttp.WriteData(ctx, w, r, nil)
			return nil
		}(); err != nil {
			ohttp.WriteError(ctx, w, r, err)
		}
	})

	// if err := recordWorker.Handle(ctx, handler); err != nil {
	// 	return errors.Wrapf(err, "handle record")
	// }

	// if err := dvrWorker.Handle(ctx, handler); err != nil {
	// 	return errors.Wrapf(err, "handle dvr")
	// }

	return nil
}