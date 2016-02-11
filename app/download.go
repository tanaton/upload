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
package app

import (
	"./conf"
	"./img/gif"
	"./img/jpeg"
	"./img/png"
	"./img/tiff"
	"./img/webp"
	"./util"
	"./util/webutil"
	"io"
	"net/http"
	"os"
	"time"
)

func Download(w http.ResponseWriter, r *http.Request, numstr, ext string) (code int, size int64, err error) {
	num, perr := util.DecodeImageId(numstr)
	if perr != nil {
		return webutil.NotFoundImage(w, r)
	}
	path := util.CreateDataPath(num, ext)

	// ファイル情報の読み取り
	fi, serr := os.Stat(path)
	if serr != nil {
		return webutil.NotFoundImage(w, r)
	}
	code = http.StatusOK
	size = fi.Size()
	mod := fi.ModTime()
	if webutil.CheckNotModified(r, mod) {
		// 304
		return webutil.NotModified(w)
	}
	// ヘッダーの設定
	w.Header().Set("Last-Modified", webutil.CreateModString(mod))
	w.Header().Set("Expires", webutil.CreateModString(time.Now().Add(conf.OneYearSec*time.Second)))
	// プロキシがキャッシュを共有しないようにする
	// ついでに、通信経路での変換や、部分的なキャッシュを防止する
	w.Header().Set("Cache-Control", "private, no-transform, max-age="+conf.OneYearSecStr)
	if r.Method != "HEAD" {
		fp, rerr := os.Open(path)
		if rerr != nil {
			return webutil.NotFoundImage(w, r)
		}
		defer fp.Close()

		b, _ := webutil.CreateUserData(r)
		switch ext {
		case "jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			err = jpeg.Send(w, fp, b)
		case "png":
			w.Header().Set("Content-Type", "image/png")
			err = png.Send(w, fp, b)
		case "gif":
			w.Header().Set("Content-Type", "image/gif")
			err = gif.Send(w, fp, b)
		case "tiff":
			w.Header().Set("Content-Type", "image/tiff")
			err = tiff.Send(w, fp, b)
		case "webp":
			w.Header().Set("Content-Type", "image/webp")
			err = webp.Send(w, fp, b)
		default:
			// 謎の拡張子
			return webutil.NotFoundImage(w, r)
		}
	}
	return code, size, err
}

func DownloadThumb(w http.ResponseWriter, r *http.Request, numstr string) (code int, size int64, err error) {
	num, perr := util.DecodeImageId(numstr)
	if perr != nil {
		return webutil.NotFoundImage(w, r)
	}
	path := util.CreateThumbPath(num)

	// ファイル情報の読み取り
	fi, serr := os.Stat(path)
	if serr != nil {
		return webutil.NotFoundImage(w, r)
	}
	code = http.StatusOK
	size = fi.Size()
	mod := fi.ModTime()
	if webutil.CheckNotModified(r, mod) {
		// 304
		return webutil.NotModified(w)
	}
	// ヘッダーの設定
	w.Header().Set("Last-Modified", webutil.CreateModString(mod))
	w.Header().Set("Expires", webutil.CreateModString(time.Now().Add(conf.OneYearSec*time.Second)))
	w.Header().Set("Cache-Control", "public, max-age="+conf.OneYearSecStr)
	if r.Method != "HEAD" {
		fp, rerr := os.Open(path)
		if rerr != nil {
			return webutil.NotFoundImage(w, r)
		}
		defer fp.Close()
		size, err = io.Copy(w, fp)
	}
	return code, size, err
}
