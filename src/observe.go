package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"translog-nix/src/logs"
	"translog-nix/src/narinfo"
)

var storePathPattern = regexp.MustCompile(`^/nix/store/([0-9abcdfghijklmnpqrsvwxyz]{32})-[a-zA-Z0-9+._=?-]+$`)

var (
	errInvalidObservationTarget = errors.New("invalid cache_url or store_path")
	errUpstreamRequest          = errors.New("upstream request failed")
	errNarInfoNotFound          = errors.New("upstream narinfo not found")
)

type observationTarget struct {
	CacheURL  string `json:"cache_url"`
	StorePath string `json:"store_path"`
}

type observation struct {
	info  narinfo.NarInfo
	index uint64
}

type observer struct {
	logDir     string
	privateKey string
	client     *http.Client

	mu     sync.Mutex
	latest map[observationTarget]observation
}

// serveObserve は POST /api/v1/observe を処理する。
func (o *observer) serveObserve(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	target, upstreamRequest, err := parseObservationRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	info, err := o.fetchNarInfo(upstreamRequest, target.StorePath)
	switch {
	case errors.Is(err, errNarInfoNotFound):
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	index, created, err := o.record(r.Context(), target, info)
	if err != nil {
		log.Printf("append observation: %v", err)
		http.Error(w, "append observation failed", http.StatusInternalServerError)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]uint64{"index": index})
}

func newObserver(logDir, privateKey string) *observer {
	return &observer{
		logDir:     logDir,
		privateKey: privateKey,
		client:     &http.Client{Timeout: 10 * time.Second},
		latest:     make(map[observationTarget]observation),
	}
}

func parseObservationRequest(r *http.Request) (observationTarget, *http.Request, error) {
	var target observationTarget
	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		return target, nil, errors.New("expected cache_url and store_path")
	}

	base, err := url.Parse(target.CacheURL)

	//TODO: あまりにもifが長くて可読性に考慮した結果このswitchが誕生したが、正しいか分からない
	// そもそも、こんなに長いifが発生するような状況を作ったことが悪い可能性があり、
	// 各型自体にバリデーション的要素を持たせるべきかもしれない
	switch {
	case err != nil,
		base.Hostname() == "",
		base.Scheme != "http" && base.Scheme != "https",
		base.User != nil,
		base.ForceQuery || base.RawQuery != "",
		base.Fragment != "":
		return target, nil, errInvalidObservationTarget
	}
	target.CacheURL = strings.TrimRight(base.String(), "/")

	path := storePathPattern.FindStringSubmatch(target.StorePath)
	if path == nil {
		return target, nil, errInvalidObservationTarget
	}

	upstreamURL := base.JoinPath(path[1] + ".narinfo")
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, upstreamURL.String(), nil)
	if err != nil {
		return target, nil, errors.New("invalid upstream URL")
	}
	return target, req, nil
}

func (o *observer) fetchNarInfo(req *http.Request, storePath string) (narinfo.NarInfo, error) {
	resp, err := o.client.Do(req)
	if err != nil {
		return narinfo.NarInfo{}, errUpstreamRequest
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return narinfo.NarInfo{}, errNarInfoNotFound
	default:
		return narinfo.NarInfo{}, errUpstreamRequest
	}

	raw, err := io.ReadAll(io.LimitReader(resp.Body, math.MaxUint16+1))
	if err != nil || len(raw) > math.MaxUint16 || !utf8.Valid(raw) {
		return narinfo.NarInfo{}, errors.New("invalid upstream body")
	}
	info, err := narinfo.Parse(string(raw))
	if err != nil || info.StorePath != storePath {
		return narinfo.NarInfo{}, errors.New("invalid upstream narinfo")
	}
	return info, nil
}

func (o *observer) record(ctx context.Context, target observationTarget, info narinfo.NarInfo) (uint64, bool, error) {
	// 比較と追記を直列化し、同じ本文の同時リクエストで重複追記を防ぐ。
	// TODO: 現状メモリにしか重複記述対策のデータを保持していないので、
	// 再起動後には重複した記述を行ってしまう問題がある
	o.mu.Lock()
	defer o.mu.Unlock()
	if previous, found := o.latest[target]; found && previous.info.Equal(info) {
		return previous.index, false, nil
	}
	entry := logs.Entry{CacheURL: target.CacheURL, NarInfo: info.String()}
	index, err := logs.Append(context.WithoutCancel(ctx), o.logDir, o.privateKey, entry)
	if err != nil {
		return 0, false, err
	}
	o.latest[target] = observation{info: info, index: index}
	return index, true, nil
}
