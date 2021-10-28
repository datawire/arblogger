package main

import (
	"context"
	"net/http"
	"os"

	"github.com/datawire/dlib/dgroup"
	"github.com/datawire/dlib/dhttp"
	"github.com/datawire/dlib/dlog"
	"github.com/golang/protobuf/jsonpb"

	als_data_v2 "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

func main() {
	ctx := context.Background()

	// We're going to be running tasks in parallel, so we'll use dgroup to manage
	// a group of goroutines to do so.
	grp := dgroup.NewGroup(ctx, dgroup.GroupConfig{
		// Enable signal handling so that SIGINT can start a graceful shutdown,
		// and a second SIGINT will force a not-so-graceful shutdown. This shutdown
		// will be signaled to the worker goroutines through the Context that gets
		// passed to them.
		EnableSignalHandling: true,
	})

	// One of those tasks will be running an HTTP server.
	grp.Go("http", func(ctx context.Context) error {
		// We'll be using a *dhttp.ServerConfig instead of an *http.Server, but it works
		// very similarly to *http.Server, everything else in the stdlib net/http package is
		// still valid; we'll still be using plain-old http.ResponseWriter and *http.Request
		// and http.HandlerFunc.
		cfg := &dhttp.ServerConfig{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hdrs := r.Header
				reqID := hdrs.Get("X-Request-Id")

				if reqID == "" {
					reqID = "no-request-id"
				}

				entry := als_data_v2.HTTPAccessLogEntry{}
				err := jsonpb.Unmarshal(r.Body, &entry)

				if err != nil {
					dlog.Errorf(r.Context(), "%s %s (%s): error decoding body: %s", r.Method, r.URL.Path, reqID, err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				// common := entry.CommonProperties
				req := entry.Request
				resp := entry.Response

				if req == nil {
					dlog.Errorf(r.Context(), "%s %s (%s): error -- no request", r.Method, r.URL.Path, reqID)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				if resp == nil {
					dlog.Errorf(r.Context(), "%s %s (%s): error -- no response", r.Method, r.URL.Path, reqID)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				method := req.RequestMethod
				path := req.Path
				reqID = req.RequestId
				code := resp.ResponseCode.Value

				dlog.Infof(r.Context(), "%s %s (%s): %d %s", method, path, reqID, code, resp.ResponseCodeDetails)
				dlog.Infof(r.Context(), "  headers %v", hdrs)

				// if resp != nil {
				// 	code := resp.ResponseCode

				// 	if code.Value == 429 {
				// 		method := req.RequestMethod
				// 		path := req.Path
				// 		reqID := req.RequestId

				// 		metadata := common.GetMetadata().FilterMetadata

				// 		rlmeta, ok := metadata["envoy.filters.http.ratelimit"]

				// 		if ok {
				// 			metadata := rlmeta.Fields

				// 			lnValue, ok := metadata["aes.ratelimit.name"]
				// 			limitName := "unknown"

				// 			if ok {
				// 				limitName = lnValue.GetStringValue()
				// 			}

				// 			actionValue, ok := metadata["aes.ratelimit.action"]
				// 			action := "unknown"

				// 			if ok {
				// 				action = actionValue.GetStringValue()
				// 			}

				// 			retryAfterValue, ok := metadata["aes.ratelimit.retry_after"]
				// 			retryAfter := -1

				// 			if ok {
				// 				retryAfter = int(retryAfterValue.GetNumberValue())
				// 			}

				// 			if action == "Enforce" {
				// 				log.Printf("---- RATELIMITED: %s (%s %s) by %s, wait %ds", reqID, method, path, limitName, retryAfter)
				// 			} else {
				// 				log.Printf("---- Limiter %s: %s (%s %s) by %s, wait %ds", action, reqID, method, path, limitName, retryAfter)
				// 			}

				// 		}
				// 	}
				// }
			}),
		}

		// ListenAndServe will gracefully shut down according to ctx; we don't need to worry
		// about separately calling .Shutdown() or .Close() like we would for *http.Server
		// (those methods don't even exist on dhttp.ServerConfig).  During a graceful
		// shutdown, it will stop listening and close idle connections, but will wait on any
		// active connections; during a not-so-graceful shutdown it will forcefully close
		// any active connections.
		//
		// If the server itself needs to log anything, it will use dlog according to ctx.
		// The Request.Context() passed to the Handler function will inherit from ctx, and
		// so the Handler will also log according to ctx.
		//
		// And, on the end-user-facing side of things, this supports HTTP/2, where
		// *http.Server.ListenAndServe wouldn't.
		return cfg.ListenAndServe(ctx, ":8080")
	})

	err := grp.Wait()

	if err != nil {
		dlog.Errorf(ctx, "finished with error: %v", err)
		os.Exit(1)
	}
}
