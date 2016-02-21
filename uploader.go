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
package main

import (
	"./app"
	"./app/conf"
	"./app/util"
	"./app/util/webutil"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

type imgcHandle struct {
	tls bool
	fs  http.Handler
}

var denyNetworkStr = []string{ /*"27.120.104.14/32",*/ }

func main() {
	webutil.InitDeny(denyNetworkStr)
	go app.BatchProc()
	h1s := &http.Server{
		Addr: ":80",
		Handler: http.TimeoutHandler(&imgcHandle{
			fs: http.FileServer(http.Dir(conf.Conf.WebRootDir)),
		}, time.Duration(conf.Conf.TimeoutHandlerSec)*time.Second, webutil.ServiceUnavailableMessage),
		ReadTimeout:    time.Duration(conf.Conf.TimeoutReadSec) * time.Second,
		WriteTimeout:   time.Duration(conf.Conf.TimeoutWriteSec) * time.Second,
		MaxHeaderBytes: int(conf.Conf.HeaderSizeMax),
	}
	h2s := &http.Server{
		Addr: ":443",
		Handler: &imgcHandle{
			tls: true,
			fs:  http.FileServer(http.Dir(conf.Conf.WebRootDir)),
		},
	}
	log.Printf("listen start %s\n", h1s.Addr)
	// サーバ起動
	go func() {
		log.Println(h2s.ListenAndServeTLS(conf.Conf.SslCertPath, conf.Conf.SslPrivkeyPath))
	}()
	log.Fatal(h1s.ListenAndServe())
}

func (ih *imgcHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// アクセス制限
	if webutil.Deny(w, r) {
		return
	}
	// 面倒なリクエストを弾く
	if webutil.Lazy(w, r) {
		return
	}

	var code int
	var size int64
	var err error
	host := r.Host
	p := path.Clean(r.URL.Path)
	dir, name := path.Split(p)
	ext := path.Ext(name)

	if ih.tls && host != "kntn.org" {
		return
	}

	if ((host == conf.Conf.ImageHost) || (dir == "/i/")) && name != "" {
		// ファイル送信
		switch ext {
		case ".jpg", ".png", ".gif", ".tiff", ".webp":
			code, size, err = app.Download(w, r, strings.TrimSuffix(name, ext), ext[1:])
		default:
			code, size, err = webutil.NotFound(w, r)
		}
	} else if ((host == conf.Conf.ThumbHost) || (dir == "/t/")) && name != "" {
		// ファイル送信
		switch ext {
		case ".jpg":
			code, size, err = app.DownloadThumb(w, r, strings.TrimSuffix(name, ext))
		default:
			code, size, err = webutil.NotFound(w, r)
		}
	} else if strings.Index(p, "/v/") == 0 {
		if ext == "" {
			code, size, err = app.ViewerHtml(w, r, ih.tls)
		} else {
			code, size, err = webutil.NotFound(w, r)
		}
	} else if p == "/api/list" {
		// ページリスト取得
		code, size, err = app.ImageList(w, r, ih.tls)
	} else if p == "/api/upload" {
		// アップロード受付
		code, size, err = app.Upload(w, r, ih.tls)
	} else if p == "/api/update" {
		// 更新受付
		code, size, err = app.Update(w, r, ih.tls)
	} else if p == "/api/delete" {
		// 削除受付
		code, size, err = app.Delete(w, r, ih.tls)
	} else if conf.Conf.Host != "" && host != conf.Conf.Host {
		// 共通のページを使わせるために転送
		var sc string
		if ih.tls {
			sc = "https"
		} else {
			sc = "http"
		}
		u := sc + "://" + conf.Conf.Host + p
		if r.URL.RawQuery != "" {
			u += "?" + r.URL.RawQuery
		}
		code, size, err = webutil.MovedPermanently(w, u)
	} else if p == "/" || p == "/index.html" {
		// トップページ
		code, size, err = app.IndexHtml(w, r, ih.tls)
	} else if ext == ".html" {
		code, size, err = app.NormalHtml(w, r, ih.tls)
	} else {
		// 後はファイルサーバーさんに任せる
		tmpw := webutil.NewResponseWriter()
		ih.fs.ServeHTTP(tmpw, r)
		if tmpw.Code == http.StatusNotFound {
			// 404だった場合、自前のエラーページを表示する
			code, size, err = webutil.NotFound(w, r)
		} else {
			// 通常出力
			var zf bool
			switch ext {
			case ".jpg", ".png", ".gif", ".tiff", ".webp":
				zf = false
			case "":
				if dir == "/.well-known/acme-challenge/" {
					tmpw.Header().Set("Content-Type", "text/plain")
					zf = false
				} else {
					zf = true
				}
			default:
				zf = true
			}
			code, size, err = webutil.Print(w, r, webutil.Output{
				Code:   tmpw.Code,
				Header: tmpw.Header(),
				Reader: tmpw.Buf,
				ZFlag:  zf && (tmpw.Buf.Len() > 1024),
			})
		}
	}

	// ログ出力
	if err != nil {
		util.PutlogError(r, code, size, err)
	}
	util.Putlog(r, code, size)
}
