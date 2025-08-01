package perf

import (
	"context"
	"crypto/tls"
	"maps"
	"net/http/httptrace"
	"time"

	"github.com/sirupsen/logrus"
)

func StartTimer(
	log *logrus.Entry,
	level logrus.Level,
	name string,
) (func(extra map[string]any), bool) {
	if log == nil || !log.Logger.IsLevelEnabled(level) {
		return func(map[string]any) {}, false
	}
	start := time.Now()
	return func(extra map[string]any) {
		fields := logrus.Fields{"dur": time.Since(start)}
		if extra != nil {
			maps.Copy(fields, extra)
		}
		log.WithFields(fields).Log(level, name)
	}, true
}

func NetTrace(
	ctx context.Context,
	log *logrus.Entry,
	level logrus.Level,
	name string,
	base logrus.Fields,
) (*httptrace.ClientTrace, func(int)) {
	if log == nil || !log.Logger.IsLevelEnabled(level) {
		return nil, func(int) {}
	}
	start := time.Now()
	record := func(event string, extra logrus.Fields) {
		f := logrus.Fields{
			"event":       event,
			"since_start": time.Since(start),
		}
		maps.Copy(f, base)
		maps.Copy(f, extra)
		log.WithFields(f).Log(level, name)
	}

	var dnsStart, connStart, tlsStart time.Time
	trace := &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone: func(httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				record("dns_done", logrus.Fields{"dns": time.Since(dnsStart)})
			}
		},
		ConnectStart: func(_, _ string) { connStart = time.Now() },
		ConnectDone: func(_, _ string, _ error) {
			if !connStart.IsZero() {
				record("conn_done", logrus.Fields{"conn": time.Since(connStart)})
			}
		},
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			if !tlsStart.IsZero() {
				record("tls_done", logrus.Fields{"tls": time.Since(tlsStart)})
			}
		},
		GotFirstResponseByte: func() { record("ttfb", nil) },
	}
	done := func(status int) {
		if status != 0 {
			record("done", logrus.Fields{"status": status})
		} else {
			record("done", nil)
		}
	}
	return trace, done
}
