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
	"./db"
	"./form"
	"./img"
	"./library/deepzoom"
	"./util"
	"./util/webutil"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/nfnt/resize"
	"image"
	"image/color"
	"image/draw"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Nombre struct {
	Prev int64
	Next int64
	Now  int64
	List []int64
}

type PageData struct {
	Rows           int64
	PageMax        int64
	PageSize       int64
	ThumbPixelSize int64
	Tagmap         map[string]struct{}
	Oldtags        string
	List           []db.Item
	Pagination     Nombre
}

type thumbChangeItem struct {
	path string
	size int
}

type deepZoomItem struct {
	num  int64
	imgt img.Type
}

var ImgStamp = map[string]image.Image{
	"adclick":             img.MustReadStamp(conf.Conf.WebRootDir + "/omake/stamp/adclick.png"),
	"afiblog":             img.MustReadStamp(conf.Conf.WebRootDir + "/omake/stamp/afiblog.png"),
	"confidential":        img.MustReadStamp(conf.Conf.WebRootDir + "/omake/stamp/confidential.png"),
	"pca_circle":          img.MustReadStamp(conf.Conf.WebRootDir + "/omake/stamp/pca_circle.png"),
	"pca_circle_yukisann": img.MustReadStamp(conf.Conf.WebRootDir + "/omake/stamp/pca_circle_yukisann.png"),
	"tsks_character":      img.MustReadStamp(conf.Conf.WebRootDir + "/omake/stamp/tsks_character.png"),
	"tsks_background":     img.MustReadStamp(conf.Conf.WebRootDir + "/omake/stamp/tsks_background.png"),
}
var ThumbChange = map[string]thumbChangeItem{
	"gurotyu":     thumbChangeItem{path: conf.Conf.WebRootDir + "/omake/thumb/gurotyu.jpg"},
	"iill":        thumbChangeItem{path: conf.Conf.WebRootDir + "/omake/thumb/iill.jpg"},
	"mosaic_proc": thumbChangeItem{size: 20},
}
var dzChan chan deepZoomItem

func init() {
	dzChan = make(chan deepZoomItem, 128)
}

func BatchProc() {
	// 1分ごとのタイマー
	tic := time.NewTicker(time.Minute)
LOOP:
	for {
		select {
		case _, ok := <-tic.C:
			// 時間による削除処理
			if !ok {
				break LOOP
			}
			waitDelete()
		case it, ok := <-dzChan:
			// DeepZoom画像生成処理
			if !ok {
				break LOOP
			}
			ext := it.imgt.Ext()
			dp := util.CreateDataPath(it.num, ext)
			im, err := img.Decode(dp, it.imgt)
			if err != nil {
				continue
			}
			dz := deepzoom.NewDeepZoom(256, true)
			dz.MakeTiles(im, it.num)
		}
	}
	tic.Stop()
}

func ImageList(w http.ResponseWriter, r *http.Request, tls bool) (code int, size int64, err error) {
	if r.Method != "GET" {
		// 対応していないメソッド
		return webutil.MethodNotAllowed(w, r)
	}
	var out webutil.Output
	out.Code = http.StatusOK
	out.Header = http.Header{}

	pd, perr := GetPage(r)
	if perr != nil {
		// DBエラーはコンテンツ無しとして流す
		return webutil.NoContent(w)
	} else if len(pd.List) > 0 {
		mod := pd.List[0].Date
		if webutil.CheckNotModified(r, mod) {
			// 304
			return webutil.NotModified(w)
		}
		out.Header.Set("Last-Modified", webutil.CreateModString(mod))
	}
	// ヘッダー出力
	out.Header.Set("Content-Type", "application/json; charset=utf-8")
	out.ZFlag = len(pd.List) > 3

	var data []byte
	data, err = json.Marshal(pd)
	out.Reader = bytes.NewReader(data)
	return webutil.Print(w, r, out)
}

// アップロード処理
func Upload(w http.ResponseWriter, r *http.Request, tls bool) (code int, size int64, err error) {
	var sc string
	if tls {
		sc = "https"
	} else {
		sc = "http"
	}

	if r.Method != "POST" {
		// 対応していないメソッド
		return webutil.MethodNotAllowed(w, r)
	}
	// 送信されたデータをテンポラリファイルに保存
	f, rferr := form.ReadForm(r, int64(conf.Conf.MemorySizeMax))
	if rferr != nil {
		switch rferr.Code {
		case http.StatusBadRequest:
			// 400
			return webutil.BadRequest(w, r)
		case http.StatusRequestEntityTooLarge:
			// 413
			return webutil.RequestEntityTooLarge(w, r)
		case http.StatusUnsupportedMediaType:
			// 415
			return webutil.UnsupportedMediaType(w, r)
		default:
			// 500
			return webutil.InternalServerError(w, r)
		}
	}
	// 添付ファイルは確実に消す
	defer func() {
		f.RemoveAll()
	}()
	if f.File == nil {
		// 400
		return webutil.BadRequest(w, r)
	}
	// 画像変換
	if cierr := convertImage(f); cierr != nil {
		switch cierr.Code {
		case http.StatusBadRequest:
			// 400
			return webutil.BadRequest(w, r)
		case http.StatusRequestEntityTooLarge:
			// 413
			return webutil.RequestEntityTooLarge(w, r)
		case http.StatusUnsupportedMediaType:
			// 415
			return webutil.UnsupportedMediaType(w, r)
		default:
			// 500
			return webutil.InternalServerError(w, r)
		}
	}
	// この時点でアップロードの第一段階は終了
	// ファイル名を決めてデータベースに登録
	num, err := db.Insert(db.InsertItem{
		RemoteAddr: r.RemoteAddr,
		Ext:        f.File.Imgt.Ext(),
		Size:       f.File.Size,
		Width:      f.File.Imgw,
		Height:     f.File.Imgh,
		Desc:       f.Value["description"],
		PassCode:   f.Value["passcode"],
		Hash:       f.File.Hash,
		Tags:       f.Value["tags"],
		DelMin:     f.IntValue["delete_wait_minute"],
	})
	if err != nil {
		// DBの挿入に失敗
		return webutil.InternalServerError(w, r)
	}
	// フォルダ確認
	ext := f.File.Imgt.Ext()
	dp := util.CreateDataPath(num, ext)
	tp := util.CreateThumbPath(num)
	if util.MakedirAll(filepath.Dir(dp)) != nil {
		return webutil.InternalServerError(w, r)
	}
	if util.MakedirAll(filepath.Dir(tp)) != nil {
		return webutil.InternalServerError(w, r)
	}
	// ファイル移動
	if os.Rename(f.File.Tf, dp) != nil {
		return webutil.InternalServerError(w, r)
	}
	f.File.Tf = ""
	if os.Rename(f.File.Ttf, tp) != nil {
		return webutil.InternalServerError(w, r)
	}
	f.File.Ttf = ""
	// アップロード処理完了
	imgid := util.EncodeImageId(num)
	// Dziファイル生成予約
	dzChan <- deepZoomItem{
		num:  num,
		imgt: f.File.Imgt,
	}

	// おまけ処理
	v := r.URL.Query()
	switch v.Get("jump") {
	case "true", "top":
		return webutil.SeeOther(w, sc+"://"+conf.Conf.Host+"/")
	case "image":
		var u string
		if tls {
			u = sc + "://" + conf.Conf.Host + "/i/" + imgid + "." + ext
		} else {
			u = sc + "://" + conf.Conf.ImageHost + "/" + imgid + "." + ext
		}
		return webutil.SeeOther(w, u)
	default:
	}
	// ヘッダー出力
	out := webutil.Retbuf(r, webutil.UploadSuccessNet, http.StatusOK)
	out.Header.Set("X-Iill-FileID", imgid)
	out.Header.Set("X-Iill-FileExt", ext)
	return webutil.Print(w, r, out)
}

func Update(w http.ResponseWriter, r *http.Request, tls bool) (code int, size int64, err error) {
	if r.Method != "POST" {
		// 対応していないメソッド
		return webutil.MethodNotAllowed(w, r)
	}
	// 送信されたデータをテンポラリファイルに保存
	f, rferr := form.ReadForm(r, int64(conf.Conf.MemorySizeMax))
	if rferr != nil {
		switch rferr.Code {
		case http.StatusBadRequest:
			// 400
			return webutil.BadRequest(w, r)
		case http.StatusRequestEntityTooLarge:
			// 413
			return webutil.RequestEntityTooLarge(w, r)
		case http.StatusUnsupportedMediaType:
			// 415
			return webutil.UnsupportedMediaType(w, r)
		default:
			// 500
			return webutil.InternalServerError(w, r)
		}
	}
	// 添付ファイルは確実に消す
	defer func() {
		f.RemoveAll()
	}()
	// この時点で更新の第一段階は終了
	if f.File != nil {
		// ファイルはアップロードできない
		// 400
		return webutil.BadRequest(w, r)
	}
	rawkey := f.Value["key"]
	if rawkey == "" {
		// 400
		return webutil.BadRequest(w, r)
	}
	v := r.URL.Query()
	id, err := util.DecodeImageId(v.Get("id"))
	if err != nil {
		// 400
		return webutil.BadRequest(w, r)
	}
	key := util.CreateStorePass(rawkey)
	err = db.Update(id, key, db.UpdateItem{
		Desc:     f.Value["description"],
		PassCode: f.Value["passcode"],
		Tags:     f.Value["tags"],
		DelMin:   f.IntValue["delete_wait_minute"],
	})
	if err != nil {
		// DBの更新に失敗
		return webutil.InternalServerError(w, r)
	}
	// 正常終了
	switch v.Get("jump") {
	case "true", "top":
		var sc string
		if tls {
			sc = "https"
		} else {
			sc = "http"
		}
		return webutil.SeeOther(w, sc+"://"+conf.Conf.Host+"/")
	default:
	}
	// ヘッダー出力
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return http.StatusOK, 0, nil
}

func Delete(w http.ResponseWriter, r *http.Request, tls bool) (code int, size int64, err error) {
	if r.Method != "GET" {
		// 対応していないメソッド
		return webutil.MethodNotAllowed(w, r)
	}
	v := r.URL.Query()
	id, pierr := util.DecodeImageId(v.Get("id"))
	if pierr != nil {
		return webutil.BadRequest(w, r)
	}
	ext, dierr := db.Delete(id, v.Get("passcode"), true)
	if dierr != nil {
		// DB削除失敗
		return webutil.BadRequest(w, r)
	}
	// 削除成功
	// ファイルを削除する
	util.DeleteImageFile(id, ext)
	// Dziファイルを削除する
	deepzoom.DeleteTiles(id)
	// 正常終了
	switch v.Get("jump") {
	case "true", "top":
		var sc string
		if tls {
			sc = "https"
		} else {
			sc = "http"
		}
		return webutil.SeeOther(w, sc+"://"+conf.Conf.Host+"/")
	default:
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return http.StatusOK, 0, nil
}

func waitDelete() {
	// 削除時間に達した画像のリストを取得
	list, err := db.GetWaitDeleteList()
	if err != nil {
		return
	}
	for _, id := range list {
		// DBの中身を削除する
		ext, derr := db.Delete(id, "", false)
		if derr != nil {
			return
		}
		// 削除成功
		// ファイルを削除する
		util.DeleteImageFile(id, ext)
		// Dziファイルを削除する
		deepzoom.DeleteTiles(id)
	}
}

func createThumbnail(im image.Image, f *form.Form) (reterr error) {
	r := im.Bounds()
	fh := f.File
	fh.Imgw = r.Max.X - r.Min.X
	fh.Imgh = r.Max.Y - r.Min.Y

	// 一時ファイルの生成
	file, err := ioutil.TempFile("", "thumbimg-")
	if err != nil {
		return err
	}
	defer file.Close()

	// tmpファイル名をセットしておく
	fh.Ttf = file.Name()

	tc, ok := ThumbChange[f.Value["thumb_change"]]
	if ok && tc.path != "" {
		rfp, err := os.Open(tc.path)
		if err != nil {
			return err
		}
		defer rfp.Close()
		_, reterr = io.Copy(file, rfp)
		if reterr == io.EOF {
			reterr = nil
		}
	} else {
		// サムネイル用に縮小
		var ts int
		var ps int
		var pt image.Point
		var m image.Image

		if ok && tc.size != 0 {
			ps = tc.size
		} else {
			ps = int(conf.Conf.ThumbPixelSize)
		}

		if fh.Imgw > fh.Imgh {
			// 横幅のほうが大きい場合
			// 縦幅を固定して縮小
			if fh.Imgh > ps {
				im = resize.Resize(0, uint(ps), im, resize.Bicubic)
				ts = ps
				r = im.Bounds()
			} else {
				ts = fh.Imgh
			}
			pt = image.Point{
				X: ((r.Max.X - r.Min.X) / 2) - (ts / 2),
				Y: 0,
			}
		} else {
			// 縦幅のほうが大きい場合
			// 横幅を固定して縮小
			if fh.Imgw > ps {
				im = resize.Resize(uint(ps), 0, im, resize.Bicubic)
				ts = ps
				r = im.Bounds()
			} else {
				ts = fh.Imgw
			}
			pt = image.Point{
				X: 0,
				Y: ((r.Max.Y - r.Min.Y) / 2) - (ts / 2),
			}
		}

		if si, ok := im.(img.SubImager); ok {
			// SubImageがあるなら使う
			m = si.SubImage(image.Rect(pt.X, pt.Y, ts+pt.X, ts+pt.Y))
		} else {
			// 新しい画像の用意
			rgba := image.NewRGBA(image.Rect(0, 0, ts, ts))
			// 真ん中付近を切り抜く
			draw.Draw(rgba, rgba.Bounds(), im, pt, draw.Src)
			m = rgba
		}
		// 書き込み
		reterr = img.EncodeThumb(file, m)
	}
	return
}

func GetPage(r *http.Request) (*PageData, error) {
	v := r.URL.Query()
	page, aierr := strconv.ParseInt(v.Get("p"), 10, 32)
	if aierr != nil {
		page = 0
	} else if page > 0 {
		page--
	} else if page < 0 {
		page = 0
	}

	var tagmap map[string]struct{}
	var oldtags string
	var list []db.Item
	var max int64
	var err error
	if tags, ok := v["tag"]; ok && len(tags) > 0 {
		tagmap = make(map[string]struct{}, len(tags))
		for _, it := range tags {
			tagmap[it] = struct{}{}
		}
		list, max, err = db.GetPageTags(page*int64(conf.Conf.PageSize), int64(conf.Conf.PageSize), tagmap)
	} else {
		list, max, err = db.GetPage(page*int64(conf.Conf.PageSize), int64(conf.Conf.PageSize))
	}
	if err != nil {
		return nil, err
	}
	if tagmap != nil {
		tmptags := make([]string, 0, len(tagmap))
		for key, _ := range tagmap {
			tmptags = append(tmptags, "tag="+url.QueryEscape(key))
		}
		oldtags = strings.Join(tmptags, "&")
	}
	pmax := (max / int64(conf.Conf.PageSize)) + 1
	return &PageData{
		Rows:           max,
		PageMax:        pmax,
		PageSize:       int64(conf.Conf.PageSize),
		ThumbPixelSize: int64(conf.Conf.ThumbPixelSize),
		Tagmap:         tagmap,
		Oldtags:        oldtags,
		List:           list,
		Pagination:     createPagination(page+1, pmax),
	}, nil
}

func createPagination(pnow, pmax int64) Nombre {
	var prev int64
	var next int64
	var list []int64
	if pnow > 1 {
		prev = pnow - 1
	}
	if pnow < pmax {
		next = pnow + 1
	}

	if pmax <= int64(conf.Conf.PaginateDefault) {
		list = make([]int64, pmax)
		for i, _ := range list {
			list[i] = int64(i) + 1
		}
	} else if (pnow - (int64(conf.Conf.PaginateDefault) / 2)) > 0 {
		list = make([]int64, conf.Conf.PaginateDefault)
		offset := pnow - (int64(conf.Conf.PaginateDefault) / 2)
		for i, _ := range list {
			list[i] = int64(i) + offset
		}
	} else {
		list = make([]int64, conf.Conf.PaginateDefault)
		for i, _ := range list {
			list[i] = int64(i) + 1
		}
	}
	return Nombre{
		Prev: prev,
		Next: next,
		Now:  pnow,
		List: list,
	}
}

func convertImage(f *form.Form) *form.Error {
	var err error
	var im image.Image
	n := f.File.Size
	// デコード
	im, err = img.Decode(f.File.Tf, f.File.Imgt)
	if err != nil {
		return form.NewError(http.StatusUnsupportedMediaType, err.Error())
	}
	// サムネイル生成
	if err = createThumbnail(im, f); err != nil {
		// 失敗するということは画像が変
		// 415
		return form.NewError(http.StatusUnsupportedMediaType, err.Error())
	}

	stamp := ImgStamp[f.Value["stamp"]]
	if (stamp != nil) && ((f.File.Imgt == img.TypeJpeg) || (f.File.Imgt == img.TypePng) || (f.File.Imgt == img.TypeBmp)) {
		im, err = stampImage(im, stamp, f.Value["stamp_position"])
		if err != nil {
			return form.NewError(http.StatusInternalServerError, err.Error())
		}
		switch f.File.Imgt {
		case img.TypeJpeg:
			n, err = createJpeg(im, f.File)
		case img.TypePng, img.TypeBmp:
			n, err = createPng(im, f.File)
		default:
			err = errors.New("invalid filetype")
		}
		if err != nil {
			return form.NewError(http.StatusInternalServerError, err.Error())
		}
	} else if f.File.Imgt == img.TypeBmp {
		// ビットマップの場合
		n, err = createPng(im, f.File)
		if err != nil {
			return form.NewError(http.StatusInternalServerError, err.Error())
		}
	}

	// ファイルサイズを保存
	f.File.Size = n
	return nil
}

func stampImage(im, stamp image.Image, pos string) (image.Image, error) {
	var mr image.Rectangle
	b := im.Bounds()
	sb := stamp.Bounds()
	dst, ok := im.(draw.Image)
	if ok == false {
		dst = image.NewRGBA(b)
		draw.Draw(dst, b, im, b.Min, draw.Src)
	}

	// スタンプ用に縮小
	if (b.Max.X > (sb.Max.X * 2)) && (b.Max.Y > (sb.Max.Y * 2)) {
		// 描画先が十分に大きいので縮小しない
	} else {
		// 描画先が小さいので縮小する
		// あんまり縮小が必要ならばスタンプはあきらめる
		if (((b.Max.X / 2) - 10) > (sb.Max.X / 3)) && ((b.Max.Y / 2) > (sb.Max.Y / 3)) {
			x := uint((b.Max.X / 2) - 10)
			y := uint(b.Max.Y / 2)
			stamp = resize.Thumbnail(x, y, stamp, resize.Bicubic)
			sb = stamp.Bounds()
		} else {
			return nil, errors.New("image size too min")
		}
	}

	switch pos {
	case "ur":
		mr = sb.Add(image.Point{X: b.Max.X - sb.Max.X - (sb.Max.X / 6), Y: sb.Max.Y / 6})
	case "cc":
		mr = sb.Add(image.Point{X: (b.Max.X / 2) - (sb.Max.X / 2), Y: (b.Max.Y / 2) - (sb.Max.Y / 2)})
	case "ll":
		mr = sb.Add(image.Point{X: sb.Max.X / 6, Y: b.Max.Y - sb.Max.Y - (sb.Max.Y / 6)})
	case "lr":
		mr = sb.Add(image.Point{X: b.Max.X - sb.Max.X - (sb.Max.X / 6), Y: b.Max.Y - sb.Max.Y - (sb.Max.Y / 6)})
	default:
		mr = sb.Add(image.Point{X: sb.Max.X / 6, Y: sb.Max.Y / 6})
	}

	// 日本郵政のはんこ作成ツール「http://yubin-nenga.jp/hanko/」の朱色
	src := &image.Uniform{color.RGBA{203, 27, 24, 255}}
	draw.DrawMask(dst, mr, src, image.ZP, stamp, sb.Min, draw.Over)
	return dst, nil
}

func createPng(im image.Image, fh *form.FileHeader) (int64, error) {
	fp, err := os.OpenFile(fh.Tf, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return 0, err
	}
	defer fp.Close()
	eerr := img.EncodePng(fp, im)
	if eerr != nil {
		return 0, eerr
	}
	st, err := fp.Stat()
	if err != nil {
		return 0, err
	}
	fh.Imgt = img.TypePng
	return st.Size(), nil
}

func createJpeg(im image.Image, fh *form.FileHeader) (int64, error) {
	fp, err := os.OpenFile(fh.Tf, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return 0, err
	}
	defer fp.Close()
	eerr := img.EncodeJpeg(fp, im)
	if eerr != nil {
		return 0, eerr
	}
	st, err := fp.Stat()
	if err != nil {
		return 0, err
	}
	fh.Imgt = img.TypeJpeg
	return st.Size(), nil
}
