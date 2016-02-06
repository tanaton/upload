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

// 細かい便利機能

import (
	"../../conf"
	"../../util"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type Output struct {
	Code   int
	Header http.Header
	Reader io.Reader
	ZFlag  bool
}

type DummyResponseWriter struct {
	Code int
	h    http.Header
	Buf  *bytes.Buffer
}

type denyMap struct {
	ipmap map[byte]*denyMap
	mask  byte
}

type Size interface {
	Size() int64
}

var imageNotFound []byte
var denyNetwork *denyMap

func (out *Output) Error() (ret string) {
	if out.Reader != nil {
		buf, err := ioutil.ReadAll(out.Reader)
		if err == nil {
			ret = string(buf)
		} else {
			ret = err.Error()
		}
	} else {
		ret = "Output Error"
	}
	return
}

func init() {
	var err error
	imageNotFound, err = ioutil.ReadFile(conf.Conf.WebRootDir + "/omake/thumb/notfound.png")
	if err != nil && err != io.EOF {
		panic(err)
	}
}

type NetFile struct {
	Buf   []byte
	Gzbuf []byte
}

func MustNewNetFile(data string) (b NetFile) {
	// 事前圧縮なので最大圧縮率で圧縮
	tmp := bytes.Buffer{}
	b.Buf = []byte(data)
	gz, _ := gzip.NewWriterLevel(&tmp, gzip.BestCompression)
	io.Copy(gz, bytes.NewReader(b.Buf))
	gz.Close()
	b.Gzbuf = tmp.Bytes()
	return b
}

func Retbuf(r *http.Request, n NetFile, code int) (out Output) {
	out.Header = http.Header{}
	out.ZFlag = false // 事前圧縮するためfalse
	out.Header.Set("Content-Type", "text/html; charset=utf-8")
	out.Header.Set("Vary", "Accept-Encoding")
	out.Code = code

	var buf []byte
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		// gzip圧縮を使う
		out.Header.Set("Content-Encoding", "gzip")
		buf = n.Gzbuf
	} else {
		buf = n.Buf
	}

	out.Reader = bytes.NewBuffer(buf)
	return
}

func Dispose(w http.ResponseWriter, r *http.Request) {
	if ret := recover(); ret != nil {
		if err, ok := ret.(error); ok {
			// 500を返しておく
			code := http.StatusInternalServerError
			w.WriteHeader(code)
			util.PutlogError(r, code, 0, err)
		}
	}
}

func Print(resw http.ResponseWriter, r *http.Request, out Output) (int, int64, error) {
	var err error
	// ヘッダー設定
	wc, ws := PreOutput(resw, r, out)
	defer wc.Close()
	// ボディ出力
	if r.Method != "HEAD" && out.Reader != nil {
		_, err = io.Copy(wc, out.Reader)
	}
	return out.Code, ws.Size(), err
}

func PreOutput(resw http.ResponseWriter, r *http.Request, out Output) (io.WriteCloser, Size) {
	// ヘッダー設定
	for key, _ := range out.Header {
		resw.Header().Set(key, out.Header.Get(key))
	}
	resw.Header().Set("X-Frame-Options", "SAMEORIGIN")

	// 出力フォーマット切り替え
	var wc io.WriteCloser
	sw := NewSizeCountWriter(resw)
	if out.ZFlag {
		resw.Header().Set("Vary", "Accept-Encoding")
		ae := r.Header.Get("Accept-Encoding")
		if strings.Contains(ae, "gzip") {
			// gzip圧縮
			resw.Header().Set("Content-Encoding", "gzip")
			wc, _ = gzip.NewWriterLevel(sw, gzip.BestSpeed)
		} else {
			// 圧縮しない
			wc = sw
		}
	} else {
		// 生データ
		wc = sw
	}
	// ステータスコード＆ヘッダー出力
	resw.WriteHeader(out.Code)
	return wc, sw
}

func Lazy(w http.ResponseWriter, r *http.Request) bool {
	switch r.Method {
	case "GET", "HEAD", "POST":
		// OK
	default:
		// NG
		NotImplemented(w, r)
		return true
	}
	if uint(len(r.URL.RequestURI())) >= conf.Conf.UrlLengthMax {
		RequestURITooLong(w, r)
		return true
	}
	return false
}

func Deny(w http.ResponseWriter, r *http.Request) bool {
	rh, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// 403
		Forbidden(w, r)
		return true
	}
	dm := denyNetwork
	var ok bool

	ipv4 := net.ParseIP(rh).To4()
	if ipv4 != nil {
		for _, it := range ipv4 {
			if dm, ok = dm.ipmap[it&dm.mask]; !ok {
				return false
			}
		}
	} else {
		// IPv6の事は特に考えていない
		return false
	}
	// 制限中のネットワークからのアクセス
	// 403
	Forbidden(w, r)
	return true
}

func InitDeny(denylist []string) {
	denyNetwork = newDenyMap()
	for _, it := range denylist {
		_, n, err := net.ParseCIDR(it)
		if err != nil {
			continue
		}
		var m [5]*denyMap
		var ok bool
		masklen := len(n.Mask)

		m[0] = denyNetwork
		for i := 0; i < 4; i++ {
			m[i+1], ok = m[i].ipmap[n.IP[i]]
			if !ok {
				m[i+1] = newDenyMap()
				m[i].ipmap[n.IP[i]] = m[i+1]
			}
			if masklen > i {
				m[i].mask &= n.Mask[i]
			}
		}
	}
}

func newDenyMap() *denyMap {
	return &denyMap{ipmap: make(map[byte]*denyMap), mask: 0xff}
}

func CreateModString(mod time.Time) string {
	return mod.UTC().Format(http.TimeFormat)
}

func GetIfModifiedSince(r *http.Request) time.Time {
	if m := r.Header.Get("If-Modified-Since"); m != "" {
		since, err := http.ParseTime(m)
		if err == nil {
			return since
		}
	}
	return time.Time{}
}

func CheckNotModified(r *http.Request, mod time.Time) bool {
	if mod.IsZero() == false {
		data := GetIfModifiedSince(r)
		if mod.Before(data.Add(1 * time.Second)) {
			// 更新なし
			return true
		}
	}
	return false
}

type SizeCountWriter struct {
	w io.Writer
	s int64
}

func NewSizeCountWriter(w io.Writer) *SizeCountWriter {
	return &SizeCountWriter{w: w}
}

func (scw *SizeCountWriter) Write(p []byte) (n int, err error) {
	n, err = scw.w.Write(p)
	scw.s += int64(n)
	return
}
func (_ *SizeCountWriter) Close() error {
	return nil
}
func (scw *SizeCountWriter) Size() int64 {
	return scw.s
}

type UserData struct {
	ServerDate time.Time
	RemoteAddr string
	RequestURI string
	Host       string
	Header     http.Header
}

func CreateUserData(r *http.Request) ([]byte, error) {
	ud := &UserData{
		ServerDate: time.Now(),
		RemoteAddr: r.RemoteAddr,
		RequestURI: r.RequestURI,
		Host:       r.Host,
		Header:     r.Header,
	}
	return json.Marshal(ud)
}

func CreateUserDataIndent(r *http.Request) ([]byte, error) {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "アクセス時間：%s\n", time.Now().String())
	fmt.Fprintf(buf, "IPアドレス：%s\n", r.RemoteAddr)
	if ua := r.Header.Get("User-Agent"); ua != "" {
		fmt.Fprintf(buf, "ユーザーエージェント：%s\n", ua)
	}
	return buf.Bytes(), nil
}

// src/pkg/net/lookup.goを参考に作成
func LookupAddrDeadline(addr string, timeout time.Duration) (host []string, err error) {
	if timeout <= 0 {
		err = errors.New("timeout")
		return
	}
	t := time.NewTimer(timeout)
	defer t.Stop()
	type res struct {
		host []string
		err  error
	}
	resc := make(chan res, 1)
	go func() {
		host, err := net.LookupAddr(addr)
		resc <- res{host, err}
	}()
	select {
	case <-t.C:
		err = errors.New("timeout")
	case r := <-resc:
		host, err = r.host, r.err
	}
	return
}

func NewResponseWriter() *DummyResponseWriter {
	return &DummyResponseWriter{
		Code: http.StatusOK,
		h:    http.Header{},
		Buf:  &bytes.Buffer{},
	}
}

func (drw *DummyResponseWriter) Header() http.Header {
	return drw.h
}

func (drw *DummyResponseWriter) Write(p []byte) (int, error) {
	return drw.Buf.Write(p)
}

func (drw *DummyResponseWriter) WriteHeader(code int) {
	drw.Code = code
}
