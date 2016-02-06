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
package form

import (
	"../conf"
	"../img"
	"../util"
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

// フォームから受け取るデータ

type FileHeader struct {
	Tf     string
	Ttf    string
	Imgt   img.Type
	Size   int64
	Imgw   int
	Imgh   int
	Header textproto.MIMEHeader
	Hash   string
}

type Form struct {
	Value    map[string]string
	IntValue map[string]int64
	File     *FileHeader
}

func (f *Form) RemoveAll() (err error) {
	if f.File != nil {
		if f.File.Tf != "" {
			e := os.Remove(f.File.Tf)
			if e != nil && err == nil {
				err = e
			}
		}
		if f.File.Ttf != "" {
			e := os.Remove(f.File.Ttf)
			if e != nil && err == nil {
				err = e
			}
		}
	}
	return err
}

type Error struct {
	Code    int
	Message string
}

func (fe Error) Error() string {
	return fe.Message
}
func NewError(code int, msg string) *Error {
	return &Error{
		Code:    code,
		Message: msg,
	}
}

func ReadForm(req *http.Request, maxMemory int64) (retform *Form, formerr *Error) {
	v := req.Header.Get("Content-Type")
	if v == "" {
		// 400
		return nil, NewError(http.StatusBadRequest, "invalid Content-Type")
	}
	mediaType, params, mterr := mime.ParseMediaType(v)
	if mterr != nil || mediaType != "multipart/form-data" {
		// 400
		return nil, NewError(http.StatusBadRequest, "invalid Content-Type")
	}
	cl, clerr := strconv.ParseInt(req.Header.Get("Content-Length"), 10, 32)
	if clerr != nil || cl > maxMemory {
		// Content-Lengthが設定されている場合、長さを確認する
		// 413
		return nil, NewError(http.StatusRequestEntityTooLarge, "upload data too large")
	}

	mr := multipart.NewReader(req.Body, params["boundary"])
	f := &Form{
		Value:    make(map[string]string),
		IntValue: make(map[string]int64),
	}
	defer func() {
		if formerr != nil {
			// エラーで戻る場合はクリア
			f.RemoveAll()
		}
	}()
	for {
		p, nerr := mr.NextPart()
		if nerr == io.EOF {
			break
		}
		if nerr != nil {
			// 400
			return nil, NewError(http.StatusBadRequest, nerr.Error())
		}

		name := p.FormName()
		if name == "" {
			continue
		}

		if p.FileName() == "" {
			// 普通のフォームデータ
			var b bytes.Buffer
			n, cerr := io.CopyN(&b, p, maxMemory)
			if cerr != nil && cerr != io.EOF {
				// 500
				return nil, NewError(http.StatusInternalServerError, cerr.Error())
			}
			maxMemory -= n
			if maxMemory <= 0 {
				// 413
				return nil, NewError(http.StatusRequestEntityTooLarge, "upload data too large")
			}
			data := b.String()
			if utf8.ValidString(data) == false {
				// 不正なUTF-8文字が含まれていた場合
				return nil, NewError(http.StatusBadRequest, "invalid post data")
			}
			switch name {
			case "description":
				data = util.Utf8Substr(data, conf.Conf.DBDescLengthMax)
				f.Value[name] = data
			case "tags":
				data = conf.RegSpace.ReplaceAllString(data, " ")
				data = util.Utf8Substr(strings.Trim(data, " "), conf.Conf.DBTagsLengthMax)
				f.Value[name] = data
			case "passcode", "key":
				data = util.Utf8Substr(data, conf.Conf.DBPassLengthMax)
				f.Value[name] = data
			case "stamp", "stamp_position", "thumb_change":
				data = util.Utf8Substr(data, conf.Conf.DBStampLengthMax)
				f.Value[name] = data
			case "delete_wait_minute":
				if data != "" {
					num, err := strconv.ParseInt(data, 10, 32)
					if err != nil {
						return nil, NewError(http.StatusBadRequest, "invalid data")
					}
					if (num > 0) && (num <= int64(conf.Conf.DeleteMinuteMax)) {
						f.IntValue[name] = num
					}
				}
			default:
				// 400
				return nil, NewError(http.StatusBadRequest, "invalid post key")
			}
		} else if name == "uploadfile" {
			// ファイルのデータ
			if f.File != nil {
				// すでにある
				// 400
				return nil, NewError(http.StatusBadRequest, "File is found")
			}
			br := bufio.NewReader(p) // bufio.Readerが多段になるのが気になる
			t, mterr := img.MimeType(p.Header.Get("Content-Type"), br)
			if mterr != nil {
				// 415
				return nil, NewError(http.StatusUnsupportedMediaType, mterr.Error())
			}
			// 早い段階でセットしておく
			f.File = &FileHeader{
				Imgt:   t,
				Header: p.Header,
			}

			// 一時ファイルに保存
			n, serr := storeFile(f.File, br, maxMemory)
			if serr != nil {
				// 500
				return nil, NewError(http.StatusInternalServerError, serr.Error())
			}
			maxMemory -= n
			if maxMemory <= 0 {
				// 413
				return nil, NewError(http.StatusRequestEntityTooLarge, "upload data too large")
			}
			// ファイルサイズを保存
			f.File.Size = n
		} else {
			// 400
			return nil, NewError(http.StatusBadRequest, "unknown data")
		}
	}
	return f, nil
}

func storeFile(fh *FileHeader, br *bufio.Reader, maxMemory int64) (int64, error) {
	// 一時ファイルの生成
	file, err := ioutil.TempFile("", "uploadimg-")
	if err != nil {
		return 0, err
	}
	defer file.Close()
	// 早い段階でセットしておく
	fh.Tf = file.Name()

	// ハッシュ生成もついでにやる
	h := md5.New()
	// ファイルコピー
	n, err := io.CopyN(file, io.TeeReader(br, h), maxMemory)
	if err != nil && err != io.EOF {
		return 0, err
	}
	fh.Hash = fmt.Sprintf("%X", h.Sum(nil))
	return n, nil
}
