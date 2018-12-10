package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"time"
)

func initLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000000"
}

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

func getLogLevel(status int) zerolog.Level {
	switch {
	case status < 200:
		fallthrough
	case status < 300:
		fallthrough
	case status < 400:
		return zerolog.InfoLevel
	case status < 500:
		return zerolog.WarnLevel
	default:
		return zerolog.ErrorLevel
	}
}

func ZeroLogLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		sw := statusWriter{ResponseWriter: w}

		next.ServeHTTP(&sw, r)

		log.WithLevel(getLogLevel(sw.status)).
			Str("method", r.Method).
			Str("from", r.RemoteAddr).
			Int("status", sw.status).
			Int("length", sw.length).
			Dur("duration", time.Since(t1)).
			Str("url", r.URL.String()).
			Msg("access")
	})
}
