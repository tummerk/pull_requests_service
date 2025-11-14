package middlewarex

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		logger(r.Context()).Info("request", RequestLogRestapi(r))

		ctx := context.WithValue(r.Context(), "startTime", startTime)
		ctx = context.WithValue(ctx, "ip", r.RemoteAddr)
		ctx = context.WithValue(ctx, "userAgent", r.UserAgent())
		r = r.WithContext(ctx)

		lw := LoggingResponseWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
			Size:           0,
		}
		next.ServeHTTP(&lw, r)
		logger(r.Context()).Info("response", ResponseLogRestapi(r.Context(), lw))
	})
}

func RequestLogRestapi(r *http.Request) slog.Attr {
	requestInfo := []slog.Attr{
		slog.String("method", r.Method),
		slog.String("path", r.URL.String()),
		slog.String("host", r.Host),
		slog.String("user_agent", r.UserAgent()),
		slog.String("ip", r.RemoteAddr),
	}
	return slog.Any("request_info", requestInfo)
}

func ResponseLogRestapi(ctx context.Context, w LoggingResponseWriter) slog.Attr {
	start := ctx.Value("startTime")
	startTime := start.(time.Time)
	responseInfo := []slog.Attr{
		slog.Int("status", w.StatusCode),
		slog.Int("Size", w.Size),
		slog.Int64("duration", time.Since(startTime).Milliseconds()),
	}
	return slog.Any("response_info", responseInfo)
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Size       int
}

func (lw *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(b)
	if err != nil {
		return size, err
	}
	lw.Size += size
	return size, nil
}

func (lw *LoggingResponseWriter) WriteHeader(statusCode int) {
	lw.ResponseWriter.WriteHeader(statusCode)
	lw.StatusCode = statusCode
}
