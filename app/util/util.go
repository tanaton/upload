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
package util

// 細かい便利機能

import (
	"../conf"
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var Stdlog chan<- string
var Errlog chan<- string

func init() {
	Stdlog = loggerProc(os.Stdout, 128)
	Errlog = loggerProc(os.Stderr, 1)
}

func loggerProc(w io.Writer, bsize int) chan<- string {
	if bsize <= 0 {
		bsize = 1
	}
	c := make(chan string, bsize)
	go writeLogProc(c, w)
	return c
}

func writeLogProc(c <-chan string, w io.Writer) {
	for s := range c {
		if len(s) > 0 && s[len(s)-1] != '\n' {
			s += "\n"
		}
		io.WriteString(w, s)
	}
}

func Putlog(r *http.Request, code int, size int64) {
	rh, _, _ := net.SplitHostPort(r.RemoteAddr)
	date := CreateDateNowLog()
	p := r.URL.Path
	if r.URL.RawQuery != "" {
		p += "?" + r.URL.RawQuery
	}
	s := fmt.Sprintf(`%s - - [%s] "%s %s %s" %d %d`, rh, date, r.Method, p, r.Proto, code, size)
	Stdlog <- s
}

func PutlogError(r *http.Request, code int, size int64, err error) {
	rh, _, _ := net.SplitHostPort(r.RemoteAddr)
	date := CreateDateNowLog()
	p := r.URL.Path
	if r.URL.RawQuery != "" {
		p += "?" + r.URL.RawQuery
	}
	s := fmt.Sprintf(`%s - - [%s] "%s %s %s" %d %d =>%s`, rh, date, r.Method, p, r.Proto, code, size, err.Error())
	Errlog <- s
}

func Stack() []byte {
	buf := make([]byte, 32*1024)
	s := runtime.Stack(buf, false)
	return buf[:s:s]
}

// Less Than for a pair of int arguments
func LTInt(v2, v1 int) bool {
	return v2 < v1
}
func LTUint64(v2, v1 uint64) bool {
	return v2 < v1
}

// Minimum of a pair of int arguments
func MinInt(v1, v2 int) (m int) {
	if LTInt(v2, v1) {
		m = v2
	} else {
		m = v1
	}
	return
}
func MinUint64(v1, v2 uint64) (m uint64) {
	if LTUint64(v2, v1) {
		m = v2
	} else {
		m = v1
	}
	return
}

// Minimum of a slice of int arguments
func MinIntS(v []int) (m int) {
	l := len(v)
	if l > 0 {
		m = v[0]
	}
	for i := 1; i < l; i++ {
		m = MinInt(m, v[i])
	}
	return
}
func MinUintS64(v []uint64) (m uint64) {
	l := len(v)
	if l > 0 {
		m = v[0]
	}
	for i := 1; i < l; i++ {
		m = MinUint64(m, v[i])
	}
	return
}

// Minimum of a variable number of int arguments
func MinIntV(v1 int, vn ...int) (m int) {
	m = v1
	if len(vn) > 0 {
		m = MinInt(m, MinIntS(vn))
	}
	return
}
func MinUintV64(v1 uint64, vn ...uint64) (m uint64) {
	m = v1
	if len(vn) > 0 {
		m = MinUint64(m, MinUintS64(vn))
	}
	return
}

func Range(start, limit, step int) (ret []int) {
	if start <= limit {
		if step <= 0 {
			return
		}
		ret = make([]int, 0, (limit-start)/step)
		for i := start; i <= limit; i += step {
			ret = append(ret, i)
		}
	} else {
		if step >= 0 {
			return
		}
		ret = make([]int, 0, (start-limit)/(step*-1))
		for i := start; i >= limit; i += step {
			ret = append(ret, i)
		}
	}
	return
}

func Utf8Substr(s string, max uint) string {
	r := []rune(s)
	l := uint(len(r))
	if l > max {
		l = max
	}
	return string(r[:l:l])
}

func CreateDateString(mod time.Time) string {
	return mod.Format("2006/01/02(Mon) 15:04:05")
}

func CreateDateNowLog() string {
	return time.Now().Format("02/Jan/2006:15:04:05 -0700")
}

type multiCloser []io.Closer

func MultiCloser(closers ...io.Closer) io.Closer {
	return multiCloser(closers)
}

func (mc multiCloser) Close() (err error) {
	for _, c := range mc {
		err = c.Close()
	}
	return
}

func MakedirAll(p string) error {
	if _, err := os.Stat(p); err != nil {
		// フォルダが無いので作る
		if err := os.MkdirAll(p, 0666); err != nil {
			return err
		}
	}
	return nil
}

func CreateDataPath(num int64, ext string) string {
	return fmt.Sprintf("%s/image/%03d/%04d.%s", conf.Conf.DataDir, num/10000, num%10000, ext)
}

func CreateThumbPath(num int64) string {
	return fmt.Sprintf("%s/thumb/%03d/%04d.jpg", conf.Conf.DataDir, num/10000, num%10000)
}

func CreateDeletePath(num int64, ext string) string {
	return fmt.Sprintf("%s/delete/%03d/%04d.%s", conf.Conf.DataDir, num/10000, num%10000, ext)
}

func CreateStorePass(pass string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(pass+conf.Conf.PassSalt)))
}

func DeleteImageFile(id int64, ext string) {
	// エラーは無視して完走させる
	dp := CreateDataPath(id, ext)
	delp := CreateDeletePath(id, ext)
	tp := CreateThumbPath(id)
	// 削除フォルダ作成
	MakedirAll(filepath.Dir(delp))
	// ファイル移動
	os.Rename(dp, delp)
	// サムネイルは消す
	os.Remove(tp)
}

func EncodeImageId(num int64) string {
	return strconv.FormatInt(num, 36)
}
func DecodeImageId(str string) (int64, error) {
	return strconv.ParseInt(str, 36, 64)
}
