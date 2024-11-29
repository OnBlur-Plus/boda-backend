package main

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/ossrs/go-oryx-lib/errors"
	ohttp "github.com/ossrs/go-oryx-lib/http"
	"github.com/ossrs/go-oryx-lib/logger"
)

type HttpService interface {
	Close() error
	Run(ctx context.Context) error
}

func NewHTTPService() HttpService {
	return &httpService{}
}

type httpService struct {
	servers []*http.Server
}

func (v *httpService) Close() error {
	v.servers = nil

	var wg sync.WaitGroup
	defer wg.Wait()

	return nil
}

func (v *httpService) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	// For debugging server, listen at 127.0.0.1:22022
	go func() {
		dh := http.NewServeMux()
		handleDebuggingGoroutines(context.Background(), dh)
		server := &http.Server{Addr: "127.0.0.1:22022", Handler: dh}
		server.ListenAndServe()
	}()

	ctx, cancel := context.WithCancel(ctx)

	handler := http.NewServeMux()
	if true {
		serviceHandler := http.NewServeMux()
		if err := handleHTTPService(ctx, serviceHandler); err != nil {
			return errors.Wrapf(err, "handle service")
		}

		handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Set common header.
			ohttp.SetHeader(w)

			// Always allow CORS.
			httpAllowCORS(w, r)

			// Allow OPTIONS for CORS.
			if r.Method == http.MethodOptions {
				w.Write(nil)
				return
			}

			// Handle by service handler.
			serviceHandler.ServeHTTP(w, r)
		})
	}

	// hls, ts파일 서빙
	var r0 error
	if true {
		addr := envPlatformListen()
		if !strings.HasPrefix(addr, ":") {
			addr = fmt.Sprintf(":%v", addr)
		}
		logger.Tf(ctx, "HTTP listen at %v", addr)

		server := &http.Server{Addr: addr, Handler: handler}
		v.servers = append(v.servers, server)

		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ctx.Done()
			logger.Tf(ctx, "shutting down HTTP server, addr=%v", addr)
			v.Close()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer cancel()
			if err := server.ListenAndServe(); err != nil && ctx.Err() != context.Canceled {
				r0 = errors.Wrapf(err, "listen %v", addr)
			}
			logger.Tf(ctx, "HTTP server is done, addr=%v", addr)
		}()
	}

	// srs hook용 서버
	var r1 error
	if true {
		addr := envMgmtListen()
		if !strings.HasPrefix(addr, ":") {
			addr = fmt.Sprintf(":%v", addr)
		}
		logger.Tf(ctx, "HTTP listen at %v", addr)

		server := &http.Server{Addr: addr, Handler: handler}
		v.servers = append(v.servers, server)

		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ctx.Done()
			logger.Tf(ctx, "shutting down HTTP server, addr=%v", addr)
			v.Close()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer cancel()
			if err := server.ListenAndServe(); err != nil && ctx.Err() != context.Canceled {
				r1 = errors.Wrapf(err, "listen %v", addr)
			}
			logger.Tf(ctx, "HTTP server is done, addr=%v", addr)
		}()
	}

	// var r2 error
	// if true {
	// 	addr := envHttpListen()
	// 	if !strings.HasPrefix(addr, ":") {
	// 		addr = fmt.Sprintf(":%v", addr)
	// 	}
	// 	logger.Tf(ctx, "HTTPS listen at %v", addr)

	// 	server := &http.Server{
	// 		Addr:    addr,
	// 		Handler: handler,
	// 		TLSConfig: &tls.Config{
	// 			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	// 				return certManager.httpsCertificate, nil
	// 			},
	// 		},
	// 	}
	// 	v.servers = append(v.servers, server)

	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		<-ctx.Done()
	// 		logger.Tf(ctx, "shutting down HTTPS server, addr=%v", addr)
	// 		v.Close()
	// 	}()

	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		defer cancel()
	// 		if err := server.ListenAndServeTLS("", ""); err != nil && ctx.Err() != context.Canceled {
	// 			r2 = errors.Wrapf(err, "listen %v", addr)
	// 		}
	// 		logger.Tf(ctx, "HTTPS server is done, addr=%v", addr)
	// 	}()
	// }

	wg.Wait()
	for _, r := range []error{r0, r1} {
		if r != nil {
			return r
		}
	}
	return nil
}

func handleHTTPService(ctx context.Context, handler *http.ServeMux) error {
	ohttp.Server = fmt.Sprintf("boda")

	if err := handleHooksService(ctx, handler); err != nil {
		return errors.Wrapf(err, "handle hooks")
	}
	if err := detectWorker.Handle(ctx, handler); err != nil {
		return errors.Wrapf(err, "handle detects")
	}

	var ep string

	proxy1985, err := httpCreateProxy("http://127.0.0.1:1985")
	if err != nil {
		return err
	}

	proxy8080, err := httpCreateProxy("http://127.0.0.1:8080")
	if err != nil {
		return err
	}

	// wellKnownFileServer := http.FileServer(http.Dir(path.Join(conf.Pwd, "containers/data")))
	hlsFileServer := http.FileServer(http.Dir(path.Join(conf.Pwd, "containers/objs/nginx/html")))

	ep = "/"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		// For HTTPS management.
		// if strings.HasPrefix(r.URL.Path, "/.well-known/") {
		// 	w.Header().Set("Cache-Control", "no-cache, max-age=0")
		// 	wellKnownFileServer.ServeHTTP(w, r)
		// 	return
		// }

		// Proxy to SRS HTTP API, for console, by /api/ prefix.
		if strings.HasPrefix(r.URL.Path, "/api/") {

			logger.Tf(ctx, "Proxy %v to backend 1985", r.URL.Path)
			proxy1985.ServeHTTP(w, r)
			return
		}

		// Always directly serve the HLS ts files.
		if fastCache.HLSHighPerformance && strings.HasSuffix(r.URL.Path, ".m3u8") {
			var m3u8ExpireInSeconds int = 10
			if fastCache.HLSLowLatency {
				m3u8ExpireInSeconds = 1 // Note that we use smaller expire time that fragment duration.
			}

			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%v", m3u8ExpireInSeconds))
			hlsFileServer.ServeHTTP(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, ".ts") {
			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%v", 600))
			hlsFileServer.ServeHTTP(w, r)
			return
		}

		if strings.HasSuffix(r.URL.Path, ".m3u8") ||
			strings.HasSuffix(r.URL.Path, ".ts") {
			logger.Tf(ctx, "Proxy %v to backend 8080", r.URL.Path)
			proxy8080.ServeHTTP(w, r)
			return
		}

		http.Redirect(w, r, "/mgmt", http.StatusFound)
	})

	return nil
}

func handleDebuggingGoroutines(ctx context.Context, handler *http.ServeMux) {
	ep := "/debug/goroutines"
	logger.Tf(ctx, "Handle %v", ep)
	handler.HandleFunc(ep, func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1<<16)
		stacklen := runtime.Stack(buf, true)
		fmt.Fprintf(w, "%s", buf[:stacklen])
	})
}