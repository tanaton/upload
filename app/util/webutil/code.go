/*
The MIT License (MIT)

Copyright (c) 2015 tanaton

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package webutil

import (
	"../../conf"
	"net/http"
	"time"
)

func NoContent(w http.ResponseWriter) (int, int64, error) {
	// 204
	w.WriteHeader(http.StatusNoContent)
	return http.StatusNoContent, 0, nil
}

func MovedPermanently(w http.ResponseWriter, u string) (int, int64, error) {
	// 301
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Location", u)
	w.WriteHeader(http.StatusMovedPermanently)
	msg := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>301 Moved Permanently</title>
</head>
<body>
<a href="` + u + `">` + u + `</a>
</body>
</html>`
	n, err := w.Write([]byte(msg))
	return http.StatusMovedPermanently, int64(n), err
}

func SeeOther(w http.ResponseWriter, u string) (int, int64, error) {
	// 303
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Location", u)
	w.WriteHeader(http.StatusSeeOther)
	msg := `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>303 See Other</title>
</head>
<body>
<a href="` + u + `">` + u + `</a>
</body>
</html>`
	n, err := w.Write([]byte(msg))
	return http.StatusSeeOther, int64(n), err
}

func NotModified(w http.ResponseWriter) (int, int64, error) {
	// 304
	w.WriteHeader(http.StatusNotModified)
	return http.StatusNotModified, 0, nil
}

func BadRequest(w http.ResponseWriter, r *http.Request) (int, int64, error) {
	// 400
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	n, err := w.Write([]byte(BadRequestMessage))
	return http.StatusBadRequest, int64(n), err
}

func Forbidden(w http.ResponseWriter, r *http.Request) (int, int64, error) {
	// 403
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	n, err := w.Write([]byte(ForbiddenMessage))
	return http.StatusForbidden, int64(n), err
}

func NotFound(w http.ResponseWriter, r *http.Request) (int, int64, error) {
	// 404
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	n, err := w.Write([]byte(NotFoundMessage))
	return http.StatusNotFound, int64(n), err
}

func NotFoundImage(w http.ResponseWriter, r *http.Request) (int, int64, error) {
	// 404
	now := time.Now()
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Last-Modified", CreateModString(now))
	w.Header().Set("Expires", CreateModString(now.Add(conf.OneYearSec*time.Second)))
	w.Header().Set("Cache-Control", "public, max-age="+conf.OneYearSecStr)
	w.WriteHeader(http.StatusNotFound)
	n, err := w.Write(imageNotFound)
	return http.StatusNotFound, int64(n), err
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) (int, int64, error) {
	// 405
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusMethodNotAllowed)
	n, err := w.Write([]byte(MethodNotAllowedMessage))
	return http.StatusMethodNotAllowed, int64(n), err
}

func RequestEntityTooLarge(w http.ResponseWriter, r *http.Request) (int, int64, error) {
	// 413
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusRequestEntityTooLarge)
	n, err := w.Write([]byte(RequestEntityTooLargeMessage))
	return http.StatusRequestEntityTooLarge, int64(n), err
}

func RequestURITooLong(w http.ResponseWriter, r *http.Request) (int, int64, error) {
	// 414
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusRequestURITooLong)
	n, err := w.Write([]byte(RequestURITooLongMessage))
	return http.StatusRequestURITooLong, int64(n), err
}

func UnsupportedMediaType(w http.ResponseWriter, r *http.Request) (int, int64, error) {
	// 415
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnsupportedMediaType)
	n, err := w.Write([]byte(UnsupportedMediaTypeMessage))
	return http.StatusUnsupportedMediaType, int64(n), err
}

func InternalServerError(w http.ResponseWriter, r *http.Request) (int, int64, error) {
	// 500
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	n, err := w.Write([]byte(InternalServerErrorMessage))
	return http.StatusInternalServerError, int64(n), err
}

func NotImplemented(w http.ResponseWriter, r *http.Request) (int, int64, error) {
	// 501
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Public", "GET, HEAD, POST")
	w.WriteHeader(http.StatusNotImplemented)
	n, err := w.Write([]byte(NotImplementedMessage))
	return http.StatusNotImplemented, int64(n), err
}

func ServiceUnavailable(w http.ResponseWriter, r *http.Request) (int, int64, error) {
	// 503
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusServiceUnavailable)
	n, err := w.Write([]byte(ServiceUnavailableMessage))
	return http.StatusServiceUnavailable, int64(n), err
}
