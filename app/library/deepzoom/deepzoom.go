/*
The MIT License (MIT)

Copyright (c) 2016 tanaton

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
package deepzoom

// 参考元（ほぼ移植）
// https://github.com/jeremytubbs/deepzoom

import (
	"../../conf"
	"../../img"
	"../../util"
	"bufio"
	"errors"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	Ext      = "dzi"
	PixelMax = 500000000
	LevelMax = 15
)

type Deepzoom struct {
	ts int
	to int
}

func NewDeepZoom(size int, overlap bool) *Deepzoom {
	dz := &Deepzoom{}
	dz.ts = size
	if overlap {
		dz.to = 1
	}
	return dz
}

func CreateDziPathSystem(num int64) string {
	dir := filepath.Join(filepath.Clean(conf.Conf.WebRootDir), filepath.Clean(conf.Conf.DziDir))
	return createDziPath(dir, num)
}

func CreateDziPath(num int64) string {
	return createDziPath(filepath.Clean(conf.Conf.DziDir), num)
}

func createDziPath(dir string, num int64) string {
	return fmt.Sprintf("%s/%03d/%04d.%s", dir, num/10000, num%10000, Ext)
}

func DeleteTiles(num int64) error {
	sp := CreateDziPathSystem(num)
	dir := strings.TrimSuffix(sp, "."+Ext) + "_files"

	var err error
	err = os.Remove(sp)
	if err != nil {
		return err
	}
	err = os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return nil
}

func CheckSize(w, h int) bool {
	return (uint64(w)*uint64(h)) < PixelMax && w < (2<<LevelMax) && h < (2<<LevelMax)
}

func (dz *Deepzoom) MakeTiles(im image.Image, num int64) error {
	sp := CreateDziPathSystem(num)
	dir := strings.TrimSuffix(sp, "."+Ext) + "_files"
	if err := util.MakedirAll(dir); err != nil {
		return err
	}

	// get image width and height
	rect := im.Bounds()
	height := rect.Max.Y
	width := rect.Max.X
	if CheckSize(width, height) == false {
		return errors.New("画像サイズが大きすぎるため、DZImageの生成を中止")
	}

	var maxDimension int
	if width > height {
		maxDimension = width
	} else {
		maxDimension = height
	}
	// calculate the number of levels
	numLevels := dz.getNumLevels(maxDimension)

	const parallel = 8
	var reterr error
	sync := make(chan struct{}, parallel)
	maxLevel := numLevels - 1
	for level := maxLevel; level >= 0; level-- {
		level_dir := filepath.Join(dir, strconv.Itoa(level))
		reterr = util.MakedirAll(level_dir)
		if reterr != nil {
			break
		}
		// calculate scale for level
		scale := dz.getScaleForLevel(numLevels, level)
		// calculate dimensions for levels
		w, h := dz.getDimensionForLevel(width, height, scale)

		var tmpimg image.Image
		switch maxLevel - level {
		case 0:
			// 縮小無し
			tmpimg = im
		case 1, 2, 3:
			// くっきり縮小
			tmpimg = resize.Resize(uint(w), uint(h), im, resize.Bicubic)
		case 4, 5, 6:
			// ぼんやり縮小
			tmpimg = resize.Resize(uint(w), uint(h), im, resize.Bilinear)
		default:
			// 適当に縮小
			tmpimg = resize.Resize(uint(w), uint(h), im, resize.NearestNeighbor)
		}
		// create tiles for level
		reterr = dz.createLevelTiles(sync, w, h, level, level_dir, tmpimg)
		if reterr != nil {
			break
		}
	}
	for i := parallel; i > 0; i-- {
		sync <- struct{}{}
	}
	close(sync)

	if reterr != nil {
		return reterr
	}

	wfp, err := os.OpenFile(sp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	dz.createDZI(wfp, height, width)
	wfp.Close()
	return nil
}

func (dz *Deepzoom) getNumLevels(maxDimension int) int {
	return ceil(math.Log2(float64(maxDimension))) + 1
}

func (dz *Deepzoom) getNumTiles(width, height int) (int, int) {
	return ceil(float64(width) / float64(dz.ts)), ceil(float64(height) / float64(dz.ts))
}

func (dz *Deepzoom) getScaleForLevel(numLevels, level int) float64 {
	maxLevel := numLevels - 1
	return math.Pow(0.5, float64(maxLevel-level))
}

func (dz *Deepzoom) getDimensionForLevel(width, height int, scale float64) (int, int) {
	return ceil(float64(width) * scale), ceil(float64(height) * scale)
}

func (dz *Deepzoom) createLevelTiles(sync chan struct{}, width, height, level int, dir string, im image.Image) error {
	// get column and row count for level
	columns, rows := dz.getNumTiles(width, height)
	// インターフェースを取得
	si, ok := im.(img.SubImager)
	// 終了用関数を用意
	endf := func() {
		<-sync
	}

	for column := 0; column < columns; column++ {
		for row := 0; row < rows; row++ {
			sync <- struct{}{}
			go func(column, row int) {
				defer endf()
				tile_file := fmt.Sprintf("%d_%d.jpg", column, row)
				x, y, w, h := dz.getTileBounds(level, column, row, width, height)
				tp := filepath.Join(dir, tile_file)
				wfp, err := os.OpenFile(tp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
				if err != nil {
					// TODO:エラーを拾う良い方法を考える
					return
				}
				defer wfp.Close()
				buf := bufio.NewWriterSize(wfp, 32*1024)

				var m image.Image
				if ok {
					m = si.SubImage(image.Rect(x, y, w+x, h+y))
				} else {
					gr := image.NewGray(image.Rect(0, 0, w, h))
					draw.Draw(gr, gr.Bounds(), im, image.Point{x, y}, draw.Src)
					m = gr
				}
				jpeg.Encode(buf, m, &jpeg.Options{Quality: 85})

				buf.Flush()
			}(column, row)
		}
	}
	return nil
}

func (dz *Deepzoom) getTileBoundsPosition(column, row int) (x int, y int) {
	if column == 0 {
		x = 0
	} else {
		x = dz.to
	}
	if row == 0 {
		y = 0
	} else {
		y = dz.to
	}
	x = (column * dz.ts) - x
	y = (row * dz.ts) - y
	return
}

func (dz *Deepzoom) getTileBounds(level, column, row, w, h int) (int, int, int, int) {
	x, y := dz.getTileBoundsPosition(column, row)
	c := 0
	if column == 0 {
		c = 1 * dz.to
	} else {
		c = 2 * dz.to
	}
	r := 0
	if row == 0 {
		r = 1 * dz.to
	} else {
		r = 2 * dz.to
	}
	width := dz.ts + c
	height := dz.ts + r
	newWidth := min(width, w-x)
	newHeight := min(height, h-y)
	return x, y, newWidth, newHeight
}

func (dz *Deepzoom) createDZI(wfp io.Writer, height, width int) {
	fmt.Fprintf(wfp, `<?xml version="1.0" encoding="UTF-8"?><Image xmlns="http://schemas.microsoft.com/deepzoom/2008" Format="jpg" Overlap="%d" TileSize="%d"><Size Height="%d" Width="%d" /></Image>`, dz.to, dz.ts, height, width)
}

func min(a, b int) (ret int) {
	if a < b {
		ret = a
	} else {
		ret = b
	}
	return
}

func ceil(a float64) int {
	ret := int(a)
	if a > float64(int(a)) {
		ret += 1
	}
	return ret
}
