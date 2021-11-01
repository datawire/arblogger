package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/datawire/dlib/dgroup"
	"github.com/datawire/dlib/dhttp"
	"github.com/datawire/dlib/dlog"
	"github.com/golang/protobuf/jsonpb"

	als_data_v2 "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
)

//////// HTTP Handler
//
// requestHandler is the http.HandlerFunc that handles requests to this service.
// You'll note that it'll basically accept any path and any method (though we check
// the path to figure out what status to return): we're trusting the caller here,
// because this is just a demo.
//
// A very important subtlety: requestHandler gets an HTTP request, which contains a
// body which describes _a different HTTP request_. In general, we want to talk to the
// user about that _second_ request -- the one being described in the body. However,
// very early on we grab some information about the _first_ request, just so we can log
// something with a chance of being helpful if things go wrong.
func requestHandler(w http.ResponseWriter, r *http.Request) {
	// Grab the headers solely so we can look for an X-Request-ID, which we use only
	// for logging if things go wrong.
	hdrs := r.Header
	reqID := hdrs.Get("X-Request-Id")

	if reqID == "" {
		reqID = "no-request-id"
	}

	// Assume that we'll return a 200 for this request...
	status := http.StatusOK

	// ...then look to the path to see if that should be overridden.
	switch r.URL.Path {
	case "/200":
		status = http.StatusOK

	case "/404":
		status = http.StatusNotFound

	case "/501":
		status = http.StatusNotImplemented

	case "/503":
		status = http.StatusServiceUnavailable

	case "/505":
		status = http.StatusHTTPVersionNotSupported

	case "/511":
		status = http.StatusNetworkAuthenticationRequired

	default:
		// We don't default to 500 because this is really an _error_, we default
		// to 500 because it's another useful value to test with in the case of
		// the user not configuring the ARBLogger for anything more specific.
		status = http.StatusInternalServerError
	}

	// Once that's done, grab the request body and parse it.
	//
	// The body must be a JSON-encoded array of HTTPAccessLogEntry messages. This is
	// a little weirder than you might expect because we need to use jsonpb to unmarshal
	// an HTTPAccessLogEntry, but there's no such concept as an array of protobuf messages.
	// So, instead, we use json.NewDecoder.Decode to parse the body into an array of
	// json.RawMessage, then we use jsonpb.Unmarshal to unmarshal each of the array members
	// into an HTTPAccessLogEntry.
	//
	// So. First parse the array itself...
	var allEntries []json.RawMessage
	err := json.NewDecoder(r.Body).Decode(&allEntries)

	// ...and bail if we didn't get a valid array...
	if err != nil {
		dlog.Errorf(r.Context(), "400 for %s %s (%s): error decoding body: %s", r.Method, r.URL.Path, reqID, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ...then iterate over the array and handle one entry at a time.
	for i, rawEntry := range allEntries {
		// If we can't unmarshal the entry here, we don't fail, we just log an error and
		// skip it.
		entry := als_data_v2.HTTPAccessLogEntry{}
		err := jsonpb.Unmarshal(bytes.NewBuffer(rawEntry), &entry)

		if err != nil {
			dlog.Errorf(r.Context(), "--- for %s %s (%s): error decoding entry %s: %s", r.Method, r.URL.Path, reqID, rawEntry, err)
			continue
		}

		// OK, we have a valid entry. Grab the request and response being described...
		req := entry.Request
		resp := entry.Response

		// ...and make sure that they both exist.
		if req == nil {
			dlog.Errorf(r.Context(), "--- for %s %s (%s): error -- no request for %s", r.Method, r.URL.Path, reqID, rawEntry)
			continue
		}

		if resp == nil {
			dlog.Errorf(r.Context(), "--- for %s %s (%s): error -- no response for %s", r.Method, r.URL.Path, reqID, rawEntry)
			continue
		}

		// Given those, grab the method and reqID of the request being described...
		method := req.RequestMethod
		reqID = req.RequestId

		// ...and then the path, which is a little weird. req.Path is the path that Envoy sent
		// upstream... but in the case of a successful request, that's probably been rewritten.
		// req.OriginalPath is the thing the client actually sent, except that it's empty for a
		// rejected request. Sigh. So we'll use req.OriginalPath if it's set, otherwise req.Path.
		path := req.OriginalPath

		if path == "" {
			path = req.Path
		}

		// We assume that this request was accepted in normal processing...
		action := "accept"
		explanation := "(normal processing)"

		// ...and then we go check for rate limiting.
		code := resp.ResponseCode.Value

		if code == 429 {
			// OK, this request was rate-limited. We should have metadata about what rule it
			// hit, etc., in the common properties sent for this request.
			//
			// All the rest of this block is about pulling out that metadata. It should be
			// supplied by "envoy.filters.http.ratelimit", and within that, we should see
			// metadata for various "aes.ratelimit" things which we can understand.
			common := entry.CommonProperties

			limitName := "unknown limit"
			retryAfter := -1

			if common == nil {
				limitName = "unknown limit (no common properties)"
				// Assume that this was rejected, since we can't know, but we did see a 429!
				action = "REJECT?"
			} else {
				metadata := common.GetMetadata()

				if metadata != nil {
					filterdata := metadata.FilterMetadata

					rlmeta, ok := filterdata["envoy.filters.http.ratelimit"]

					if ok {
						fields := rlmeta.Fields

						lnValue, ok := fields["aes.ratelimit.name"]

						if ok {
							limitName = lnValue.GetStringValue()
						}

						actionValue, ok := fields["aes.ratelimit.action"]

						if ok {
							action = actionValue.GetStringValue()
						}

						retryAfterValue, ok := fields["aes.ratelimit.retry_after"]

						if ok {
							retryAfter = int(retryAfterValue.GetNumberValue())
						}
					}
				}
			}

			// Once we've pulled out whatever metadata we can, format it nicely so we
			// can log about it.
			partial := limitName

			if retryAfter >= 0 {
				partial += fmt.Sprintf(", wait %d", retryAfter)
			}

			if action == "Enforce" {
				explanation = fmt.Sprintf("RATELIMITED by %s", partial)
				action = "REJECT"
			} else {
				explanation = fmt.Sprintf("Limiter %s by %s", action, partial)
				action = "accept"
			}
		}

		// OK, we have anything we can find out, so tell the user what's up.
		dlog.Infof(r.Context(), "%03d for entry %02d: %s %s %s (%s): %d %s", status, i, action, method, path, reqID, code, explanation)
	}

	// After processing all the entries, return whatever status we decided on.
	w.WriteHeader(status)
}

//////// MAINLINE

func main() {
	// Start by setting up our context.
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
		// very similarly to *http.Server. Here, we'll set it up to use our requestHandler
		// for everything.
		cfg := &dhttp.ServerConfig{
			Handler: http.HandlerFunc(requestHandler),
		}

		// Next up, figure out if we should use TLS.
		secretDir := os.Getenv("ARB_TLS")

		// Finally, we'll actually start the server, using ListenAndServeTLS or ListenAndServe
		// as appropriate. Since this is dhttp, both handle shutdown for us, so we don't need
		// to worry about calling Shutdown() or Close() ourselves (they don't exist in dhttp,
		// in fact.)
		if secretDir == "" {
			dlog.Infof(ctx, "ARBLOGGER serving HTTP on 8080")
			return cfg.ListenAndServe(ctx, ":8080")
		}

		dlog.Infof(ctx, "ARBLOGGER serving HTTPS on 8443")
		return cfg.ListenAndServeTLS(ctx, ":8443", secretDir+"/tls.crt", secretDir+"/tls.key")
	})

	// After that, we just wait for shutdown.
	err := grp.Wait()

	if err != nil {
		dlog.Errorf(ctx, "finished with error: %v", err)
		os.Exit(1)
	}
}
