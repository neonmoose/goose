package srvutil

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Shopify/goose/logger"
	"github.com/Shopify/goose/metrics"
	"github.com/Shopify/goose/redact"
	"github.com/Shopify/goose/statsd"
)

type RequestObserver interface {
	BeforeRequest(*http.Request)
	AfterRequest(*http.Request, HTTPRecorder, time.Duration)
}

type DefaultRequestObserver struct{}

func (o *DefaultRequestObserver) BeforeRequest(r *http.Request) {
	ctx := r.Context()

	log(ctx).
		WithField("method", r.Method).
		WithField("headers", redact.Headers(r.Header)).
		Info("http request")
}

func (o *DefaultRequestObserver) AfterRequest(r *http.Request, recorder HTTPRecorder, requestDuration time.Duration) {
	ctx := r.Context()

	fields := map[string]interface{}{
		"statusCode":  recorder.StatusCode(),
		"statusClass": fmt.Sprintf("%dxx", recorder.StatusCode()/100), // 2xx, 5xx, etc.
	}

	ctx = statsd.WithTags(ctx, fields)
	ctx = logger.WithFields(ctx, fields)

	metrics.HTTPRequest.Duration(ctx, requestDuration)

	logger := log(ctx).
		WithField("headers", redact.Headers(recorder.Header()))

	if body := recorder.ResponseBody(); body != nil {
		logger = logger.WithField("responseBody", *body)
	}

	logger.Info("http response")
}
