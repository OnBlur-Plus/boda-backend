// Copyright (c) 2022-2024 Winlin
//
// SPDX-License-Identifier: MIT
package main

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ossrs/go-oryx-lib/errors"
	ohttp "github.com/ossrs/go-oryx-lib/http"
	"github.com/ossrs/go-oryx-lib/logger"
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

			// verifiedBy := "noVerify"
			// if action == SrsActionOnPublish {
			// 	isSecretOK := func(publish, stream, param string) bool {
			// 		return publish == "" || strings.Contains(param, publish) || strings.Contains(stream, publish)
			// 	}

			// 	// Use live room secret to verify if stream name matches.
			// 	roomPublishAuthKey := GenerateRoomPublishKey(streamObj.Stream)
			// 	publish, err := rdb.HGet(ctx, SRS_AUTH_SECRET, roomPublishAuthKey).Result()
			// 	verifiedBy = "room"
			// 	if publish == "" {
			// 		// Use global publish secret to verify
			// 		publish, err = rdb.HGet(ctx, SRS_AUTH_SECRET, "pubSecret").Result()
			// 		verifiedBy = "global"
			// 	}
			// 	if err != nil && err != redis.Nil {
			// 		return errors.Wrapf(err, "hget %v pubSecret", SRS_AUTH_SECRET)
			// 	}
			// 	if !isSecretOK(publish, streamObj.Stream, streamObj.Param) {
			// 		return errors.Errorf("invalid normal stream=%v, param=%v, action=%v", streamObj.Stream, streamObj.Param, action)
			// 	}
			// }
			// streamURL := streamObj.StreamURL()
			// if action == SrsActionOnPublish {
			// 	streamObj.Update = time.Now().Format(time.RFC3339)

			// 	b, err := json.Marshal(&streamObj)
			// 	if err != nil {
			// 		return errors.Wrapf(err, "marshal json")
			// 	} else if err = rdb.HSet(ctx, SRS_STREAM_ACTIVE, streamURL, string(b)).Err(); err != nil && err != redis.Nil {
			// 		return errors.Wrapf(err, "hset %v %v %v", SRS_STREAM_ACTIVE, streamURL, string(b))
			// 	}

			// 	if err := rdb.HIncrBy(ctx, SRS_STAT_COUNTER, "publish", 1).Err(); err != nil && err != redis.Nil {
			// 		return errors.Wrapf(err, "hincrby %v publish 1", SRS_STAT_COUNTER)
			// 	}
			// } else if action == SrsActionOnUnpublish {
			// 	if err := rdb.HDel(ctx, SRS_STREAM_ACTIVE, streamURL).Err(); err != nil && err != redis.Nil {
			// 		return errors.Wrapf(err, "hset %v %v", SRS_STREAM_ACTIVE, streamURL)
			// 	}
			// } else if action == "on_play" {
			// 	if err := rdb.HIncrBy(ctx, SRS_STAT_COUNTER, "play", 1).Err(); err != nil && err != redis.Nil {
			// 		return errors.Wrapf(err, "hincrby %v play 1", SRS_STAT_COUNTER)
			// 	}
			// }

			ohttp.WriteData(ctx, w, r, nil)
			// logger.Tf(ctx, "srs hooks ok, action=%v, verifiedBy=%v, %v, %v",
			// 	action, verifiedBy, streamObj.String(), requestBody)
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
			b, err := ioutil.ReadAll(r.Body)
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
				return errors.Wrapf(err, "invalid ts file %v", msg.File)
			}
			logger.Tf(ctx, "on_hls ok, %v", string(b))

			// Handle TS file by Record task if enabled.
			// if recordAll, err := rdb.HGet(ctx, SRS_RECORD_PATTERNS, "all").Result(); err != nil && err != redis.Nil {
			// 	return errors.Wrapf(err, "hget %v all", SRS_RECORD_PATTERNS)
			// } else if recordAll == "true" {
			// 	if err = recordWorker.OnHlsTsMessage(ctx, &msg); err != nil {
			// 		return errors.Wrapf(err, "feed %v", msg.String())
			// 	}
			// 	logger.Tf(ctx, "record %v", msg.String())
			// }

			// Handle TS file by DVR task if enabled.
			// if dvrAll, err := rdb.HGet(ctx, SRS_DVR_PATTERNS, "all").Result(); err != nil && err != redis.Nil {
			// 	return errors.Wrapf(err, "hget %v all", SRS_DVR_PATTERNS)
			// } else if dvrAll == "true" {
			// 	if err = dvrWorker.OnHlsTsMessage(ctx, &msg); err != nil {
			// 		return errors.Wrapf(err, "feed %v", msg.String())
			// 	}
			// 	logger.Tf(ctx, "dvr %v", msg.String())
			// }

			// Handle TS file by Transcript task if enabled.
			// if transcriptWorker.Enabled() {
			// 	if err = transcriptWorker.OnHlsTsMessage(ctx, &msg); err != nil {
			// 		return errors.Wrapf(err, "feed %v", msg.String())
			// 	}
			// 	logger.Tf(ctx, "transcript %v", msg.String())
			// }
			
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