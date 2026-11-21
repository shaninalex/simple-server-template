package httpx

import (
	"net/http"
	"simple-server-template/app/core/errs"
)

// StatusFromError maps an error classification to an HTTP status code.
func StatusFromError(err error) int {
	switch errs.KindOf(err) {
	case errs.KindValidation:
		return http.StatusBadRequest
	case errs.KindNotFound:
		return http.StatusNotFound
	case errs.KindConflict:
		return http.StatusConflict
	case errs.KindForbidden:
		return http.StatusForbidden
	case errs.KindUnauthenticated:
		return http.StatusUnauthorized
	case errs.KindRateLimited:
		return http.StatusTooManyRequests
	case errs.KindPlatform:
		// An outer dependency (remote API, mail server, ...) failed.
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}
