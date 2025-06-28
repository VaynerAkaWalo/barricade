package htp

import (
	"barricade/internal/infrastructure/logging"
	"barricade/pkg/uuid"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type errorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type statusCatcher struct {
	writer     http.ResponseWriter
	statusCode int
}

type HttpHandler func(http.ResponseWriter, *http.Request) error

func (handler HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	txId := uuid.New()
	ctx := context.WithValue(r.Context(), logging.TxKey, txId)

	var logAttr []any

	var code int
	sc := &statusCatcher{writer: w}

	err := handler(sc, r.WithContext(ctx))
	if err != nil {
		var httpError *HttpError
		if errors.As(err, &httpError) {
			code = httpError.Code
		}

		_ = WriteResponse(w, code, errorResponse{Message: err.Error(), Code: code})
		logAttr = append(logAttr, "error", err.Error())
	}

	if sc.statusCode != 0 {
		code = sc.statusCode
	}

	logAttr = append(logAttr, "status_code", code, "response_time", time.Since(start).Milliseconds())

	if code < 400 {
		slog.InfoContext(ctx, "request completed", logAttr...)
	} else {
		slog.ErrorContext(ctx, "request failed", logAttr...)
	}
}

func (s *statusCatcher) Header() http.Header {
	return s.writer.Header()
}

func (s *statusCatcher) Write(bytes []byte) (int, error) {
	return s.writer.Write(bytes)
}

func (s *statusCatcher) WriteHeader(statusCode int) {
	s.statusCode = statusCode
	s.writer.WriteHeader(statusCode)
}
