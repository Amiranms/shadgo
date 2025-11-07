//go:build !solution

package requestlog

import (
	"math/rand"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RndString(length int) string {
	return StringWithCharset(length, charset)
}

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

func Log(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hash := RndString(5)
			start := time.Now()
			commonFields := []zap.Field{
				zap.String("request_id", hash),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			}

			l.Info("request started", commonFields...)

			wrappedWriter := &responseWriter{ResponseWriter: w}

			defer func() {
				if rec := recover(); rec != nil {
					l.Panic("request panicked",
						append(commonFields,
							zap.Any("panic", rec),
						)...,
					)

					if !wrappedWriter.wroteHeader {
						w.WriteHeader(http.StatusInternalServerError)
					}
				}
			}()

			next.ServeHTTP(wrappedWriter, r)

			if !wrappedWriter.wroteHeader {
				wrappedWriter.WriteHeader(http.StatusOK)
			}

			l.Info("request finished",
				append(commonFields,
					zap.Int("status_code", wrappedWriter.status),
					zap.Duration("duration", time.Since(start)),
				)...,
			)
		})
	}
}
