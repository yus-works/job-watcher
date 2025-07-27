package perf

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http/httptrace"
	"time"

	"github.com/yus-works/job-watcher/internal/logging"
)

func StartTimer(
	ctx context.Context,
	level slog.Level,
	name string,
	attrs ...slog.Attr,
) (func(extra ...slog.Attr), bool) {
	log := logging.From(ctx)
	if !log.Enabled(ctx, level) {
		return func(...slog.Attr) {}, false
	}
	start := time.Now()
	return func(extra ...slog.Attr) {
		all := make([]slog.Attr, 0, len(attrs)+len(extra)+1)
		all = append(all, attrs...)
		all = append(all, slog.Duration("dur", time.Since(start)))
		all = append(all, extra...)
		log.LogAttrs(ctx, level, name, all...)
	}, true
}

// NetTrace returns an httptrace and a done(status) logger if enabled; otherwise no-ops.
// Usage:
//
//	var trace *httptrace.ClientTrace
//	var done func(int)
//	trace, done = perf.NetTrace(ctx, log, slog.LevelDebug, "fetch", slog.String("feed", f.Name))
//	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
//	resp, err := c.Do(req)
//	if err == nil { done(resp.StatusCode) } else { done(0) }
func NetTrace(
	ctx context.Context, level slog.Level, name string, base ...slog.Attr,
) (*httptrace.ClientTrace, func(int)) {
	log := logging.From(ctx)

	if log == nil || !log.Enabled(ctx, level) {
		return nil, func(int) {}
	}

	start := time.Now()

	var dnsStart, connStart, tlsStart time.Time
	record := func(event string, extra ...slog.Attr) {
		log.LogAttrs(ctx, level, name,
			append(append([]slog.Attr{
				slog.String("event", event),
				slog.Duration("since_start", time.Since(start)),
			}, base...), extra...)...)
	}

	trace := &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone: func(httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				record("dns_done", slog.Duration("dns", time.Since(dnsStart)))
			}
		},
		ConnectStart: func(_, _ string) { connStart = time.Now() },
		ConnectDone: func(_, _ string, _ error) {
			if !connStart.IsZero() {
				record("conn_done", slog.Duration("conn", time.Since(connStart)))
			}
		},
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			if !tlsStart.IsZero() {
				record("tls_done", slog.Duration("tls", time.Since(tlsStart)))
			}
		},
		GotFirstResponseByte: func() { record("ttfb") },
	}

	done := func(status int) {
		attrs := base
		if status != 0 {
			attrs = append(attrs, slog.Int("status", status))
		}
		record("done", attrs...)
	}

	return trace, done
}
