package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv/v1.43.0/httpconv"
)

const meterName = "simple-server-template/app/adapters/httpx"

// durationBuckets duration boundaries.
// docs: https://opentelemetry.io/docs/specs/semconv/http/http-metrics/#http-server
var durationBuckets = []float64{
	0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10,
}

func Metrics(mp metric.MeterProvider) (func(http.Handler) http.Handler, error) {
	meter := mp.Meter(meterName)

	duration, err := httpconv.NewServerRequestDuration(meter, metric.WithExplicitBucketBoundaries(durationBuckets...))
	if err != nil {
		return nil, fmt.Errorf("http metrics: %w", err)
	}
	active, err := httpconv.NewServerActiveRequests(meter)
	if err != nil {
		return nil, fmt.Errorf("http metrics: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			method := requestMethod(r.Method)
			scheme := requestScheme(r)

			active.Add(r.Context(), 1, method, scheme)
			rec := newStatusRecorder(w)

			defer func() {
				p := recover()
				if p != nil {
					rec.status = http.StatusInternalServerError
				}
				active.Add(r.Context(), -1, method, scheme)
				duration.Record(r.Context(), time.Since(start).Seconds(), method, scheme, responseAttrs(duration, r, rec.status)...)
				if p != nil {
					panic(p)
				}
			}()

			next.ServeHTTP(rec, r)
		})
	}, nil
}

func responseAttrs(d httpconv.ServerRequestDuration, r *http.Request, status int) []attribute.KeyValue {
	attrs := []attribute.KeyValue{d.AttrResponseStatusCode(status)}
	if route := routeOf(r.Pattern); route != "" {
		attrs = append(attrs, d.AttrRoute(route))
	}
	if status >= 500 {
		attrs = append(attrs, d.AttrErrorType(httpconv.ErrorTypeAttr(strconv.Itoa(status))))
	}
	return attrs
}

func routeOf(pattern string) string {
	if i := strings.IndexByte(pattern, '/'); i >= 0 {
		return pattern[i:]
	}
	return ""
}

func requestMethod(m string) httpconv.RequestMethodAttr {
	switch m {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodConnect,
		http.MethodOptions, http.MethodTrace:
		return httpconv.RequestMethodAttr(m)
	default:
		return httpconv.RequestMethodOther
	}
}

func requestScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
