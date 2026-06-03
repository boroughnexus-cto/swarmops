// quota-proxy is a transparent pass-through HTTP proxy that captures
// Anthropic rate-limit headers from API responses and exposes them via /quota.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	anthropicAPI = "https://api.anthropic.com"
	headerPrefix = "anthropic-ratelimit-unified"
)

type quotaData struct {
	CapturedAt    int64   `json:"captured_at"`
	OverallStatus string  `json:"overall_status"`
	BindingWindow string  `json:"binding_window"`
	Session5h     *window `json:"session_5h,omitempty"`
	Weekly7d      *window `json:"weekly_7d,omitempty"`
}

type window struct {
	Utilization float64 `json:"utilization"`
	PercentLeft float64 `json:"percent_left"`
	Status      string  `json:"status"`
	ResetEpoch  int64   `json:"reset_epoch"`
}

type proxyState struct {
	mu    sync.RWMutex
	data  quotaData
	valid bool
}

var state proxyState

// QuotaTransport is an http.RoundTripper that captures response headers.
type QuotaTransport struct {
	target *url.URL
}

func (t *QuotaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read body so it can be replayed to the target (may have been consumed by Director)
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}

	// Build fresh request for the target, preserving method/headers/body
	targetURL := t.target.ResolveReference(req.URL)
	req2 := &http.Request{
		Method:     req.Method,
		URL:        targetURL,
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Header:     req.Header.Clone(),
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		GetBody:    func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewBuffer(body)), nil },
		Host:       t.target.Host,
		RemoteAddr: req.RemoteAddr,
	}

	resp, err := http.DefaultTransport.RoundTrip(req2)
	if err != nil {
		return resp, err
	}

	parseHeaders(resp.Header, resp.StatusCode)
	return resp, nil
}

func main() {
	port := flag.Int("port", 8082, "Port to listen on")
	flag.Parse()

	targetURL, _ := url.Parse(anthropicAPI)

	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			// The incoming request has a request-target URL (just the path, no scheme/host).
			// Read and restore body (consumed by httputil before Director runs for some requests).
			if r.Body != nil {
				b, _ := io.ReadAll(r.Body)
				r.Body.Close()
				r.Body = io.NopCloser(bytes.NewBuffer(b))
			}
			// Rewrite to target API
			r.URL.Scheme = targetURL.Scheme
			r.URL.Host = targetURL.Host
			r.Host = targetURL.Host
		},
		Transport: &QuotaTransport{target: targetURL},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/quota", handleQuota)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.Handle("/", proxy)

	addr := fmt.Sprintf("localhost:%d", *port)
	log.Printf("quota-proxy listening on %s → %s", addr, anthropicAPI)

	// No ReadTimeout/WriteTimeout: this proxies Anthropic streaming responses,
	// which can run for many minutes. A WriteTimeout severs the stream
	// mid-response — the client sees "socket connection was closed
	// unexpectedly" and every long Claude call fails. ReadHeaderTimeout guards
	// against slow-header (slowloris) clients; IdleTimeout reaps idle keep-alives.
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 30 * time.Second, IdleTimeout: 120 * time.Second}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}

func handleQuota(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	state.mu.RLock()
	data := state.data
	valid := state.valid
	state.mu.RUnlock()
	if !valid {
		http.Error(w, `{"error":"no quota data yet — make a request first"}`, http.StatusServiceUnavailable)
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

func parseHeaders(h http.Header, statusCode int) {
	sessionUtil := getFloatHeader(h, headerPrefix+"-5h-utilization")
	sessionStatus := getHeader(h, headerPrefix+"-5h-status")
	sessionReset := getIntHeader(h, headerPrefix+"-5h-reset")
	weeklyUtil := getFloatHeader(h, headerPrefix+"-7d-utilization")
	weeklyStatus := getHeader(h, headerPrefix+"-7d-status")
	weeklyReset := getIntHeader(h, headerPrefix+"-7d-reset")
	representative := getHeader(h, headerPrefix+"-representative-claim")

	overallStatus := sessionStatus
	if weeklyStatus == "warning" || weeklyStatus == "rate_limited" {
		overallStatus = weeklyStatus
	}

	var bindingWindow string
	switch representative {
	case "five_hour":
		bindingWindow = "five_hour"
	case "seven_day":
		bindingWindow = "seven_day"
	default:
		if sessionReset > 0 && weeklyReset > 0 {
			if sessionReset < weeklyReset {
				bindingWindow = "five_hour"
			} else {
				bindingWindow = "seven_day"
			}
		}
	}

	hasData := sessionStatus != "unknown" || weeklyStatus != "unknown"

	state.mu.Lock()
	state.data = quotaData{
		CapturedAt:    time.Now().Unix(),
		OverallStatus: overallStatus,
		BindingWindow: bindingWindow,
		Session5h: &window{
			Utilization: sessionUtil,
			PercentLeft: math.Round((1 - sessionUtil) * 100),
			Status:      sessionStatus,
			ResetEpoch:  sessionReset,
		},
		Weekly7d: &window{
			Utilization: weeklyUtil,
			PercentLeft: math.Round((1 - weeklyUtil) * 100),
			Status:      weeklyStatus,
			ResetEpoch:  weeklyReset,
		},
	}
	state.valid = hasData
	state.mu.Unlock()
}

func getHeader(h http.Header, name string) string {
	if v := h.Get(name); v != "" {
		return v
	}
	return "unknown"
}

func getFloatHeader(h http.Header, name string) float64 {
	if v := h.Get(name); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

func getIntHeader(h http.Header, name string) int64 {
	if v := h.Get(name); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return 0
}
