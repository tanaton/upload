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
	"./library/deepzoom"
	"./util"
	"./util/webutil"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"
)

type DefaultData struct {
	TLS  bool
	Host string
}

type IndexData struct {
	DefaultData
	MainUrl   string
	ImageUrl  string
	ThumbUrl  string
	ViewerUrl string
	SubTitle  string
	Pd        *PageData
	Find      func(map[string]struct{}, string) bool
}

type DziData struct {
	DefaultData
	DziPath  string
	UserData string
	Num      int64
}

var dashboardTempl = template.Must(template.ParseFiles("template/dashboard.templ"))
var dziviewerTempl = template.Must(template.ParseFiles("template/dziviewer.templ"))
var normalTempls = template.New("")

func IndexHtml(w http.ResponseWriter, r *http.Request, tls bool) (code int, size int64, err error) {
	var out webutil.Output
	var iu string
	var tu string
	var vu string
	out.Code = http.StatusOK
	out.Header = http.Header{}
	out.ZFlag = true
	// サムネイルはsslじゃないパターンも許容する
	if tls {
		tu = "https://" + conf.Conf.Host + "/t/"
	} else {
		tu = "http://" + conf.Conf.Host + "/t/"
	}
	iu = "https://" + conf.Conf.Host + "/i/"
	vu = "https://" + conf.Conf.Host + "/v/"

	// ヘッダー出力
	out.Header.Set("Content-Type", "text/html; charset=utf-8")
	if tls {
		out.Header.Set("Strict-Transport-Security", "max-age="+conf.OneYearSecStr)
	}

	data := &IndexData{
		DefaultData: DefaultData{
			TLS:  tls,
			Host: conf.Conf.Host,
		},
		ImageUrl:  iu,
		ThumbUrl:  tu,
		ViewerUrl: vu,
		SubTitle:  "",
		Find:      Find,
	}
	// ページ取得
	data.Pd, err = GetPage(r)

	wc, ws := webutil.PreOutput(w, r, out)
	defer func() {
		wc.Close()
		size = ws.Size()
	}()
	//dashboardTempl := template.Must(template.ParseFiles("template/dashboard.templ"))
	dashboardTempl.Execute(wc, data)
	return out.Code, 0, nil
}

func Find(m map[string]struct{}, key string) bool {
	_, ok := m[key]
	return ok
}

func NormalHtml(w http.ResponseWriter, r *http.Request, tls bool) (code int, size int64, err error) {
	out := webutil.Output{
		Code:   http.StatusOK,
		Header: http.Header{},
		ZFlag:  true,
	}
	// ヘッダー出力
	out.Header.Set("Content-Type", "text/html; charset=utf-8")
	if tls {
		out.Header.Set("Strict-Transport-Security", "max-age="+conf.OneYearSecStr)
	}

	data := &DefaultData{
		TLS:  tls,
		Host: conf.Conf.Host,
	}
	wc, ws := webutil.PreOutput(w, r, out)
	defer func() {
		wc.Close()
		size = ws.Size()
	}()

	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	name := path.Clean(upath)
	t := normalTempls.Lookup(name)
	if t == nil {
		fp, err := http.Dir(conf.Conf.WebRootDir).Open(name)
		if err != nil {
			return webutil.NotFound(w, r)
		}
		defer fp.Close()
		b, err := ioutil.ReadAll(fp)
		if err != nil {
			return webutil.NotFound(w, r)
		}
		t = normalTempls.New(name)
		_, perr := t.Parse(string(b))
		if perr != nil {
			return webutil.NotFound(w, r)
		}
	}
	//dashboardTempl := template.Must(template.ParseFiles("template/dashboard.templ"))
	t.Execute(wc, data)
	return out.Code, 0, nil
}

func ViewerHtml(w http.ResponseWriter, r *http.Request, tls bool) (code int, size int64, err error) {
	var out webutil.Output
	out.Code = http.StatusOK
	out.Header = http.Header{}
	out.ZFlag = false
	_, file := path.Split(path.Clean(r.URL.Path))

	num, perr := util.DecodeImageID(file)
	if perr != nil {
		return webutil.NotFound(w, r)
	}
	// パス取得
	dzip := deepzoom.CreateDziPathSystem(num)
	if _, err := os.Stat(dzip); err != nil {
		// 生成されてない模様
		return webutil.NotFound(w, r)
	}
	ud, _ := webutil.CreateUserDataIndent(r)

	// ヘッダー出力
	out.Header.Set("Content-Type", "text/html; charset=utf-8")
	if tls {
		out.Header.Set("Strict-Transport-Security", "max-age="+conf.OneYearSecStr)
	}

	data := &DziData{
		DefaultData: DefaultData{
			TLS:  tls,
			Host: conf.Conf.Host,
		},
		DziPath:  "/" + deepzoom.CreateDziPath(num),
		UserData: template.JSEscapeString(string(ud)),
		Num:      num,
	}
	wc, ws := webutil.PreOutput(w, r, out)
	defer func() {
		wc.Close()
		size = ws.Size()
	}()
	//dziviewerTempl = template.Must(template.ParseFiles("template/dziviewer.templ"))
	dziviewerTempl.Execute(wc, data)
	return out.Code, 0, nil
}
