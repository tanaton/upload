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
package img

import (
	"./bmp"
	"./gif"
	"./jpeg"
	"./png"
	"./tiff"
	"./webp"
	"bufio"
	"errors"
)

type Type int

const (
	TypeUndefined Type = iota
	TypeJpeg
	TypePng
	TypeGif
	TypeTiff
	TypeWebp
	TypeBmp
	TypeMax
)

var unknownImage = errors.New("unknown image type")

func (t Type) String() string {
	return t.Ext()
}

func (t Type) Ext() (ret string) {
	switch t {
	case TypeJpeg:
		ret = "jpg"
	case TypePng:
		ret = "png"
	case TypeGif:
		ret = "gif"
	case TypeTiff:
		ret = "tiff"
	case TypeWebp:
		ret = "webp"
	case TypeBmp:
		ret = "bmp"
	default:
	}
	return
}

func (t Type) CheckFile(br *bufio.Reader) (err error) {
	switch t {
	case TypeJpeg: // JPEG
		err = CheckJpeg(br)
	case TypePng:
		err = CheckPng(br)
	case TypeGif:
		err = CheckGif(br)
	case TypeBmp:
		err = CheckBmp(br)
	case TypeTiff:
		err = CheckTiff(br)
	case TypeWebp:
		err = CheckWebp(br)
	default:
		// 知らないファイル
		err = unknownImage
	}
	return
}

func CheckJpeg(br *bufio.Reader) error {
	// JPEG系のアップロードチェック
	buff, err := br.Peek(2)
	if err != nil {
		return err
	}
	if herr := jpeg.CheckHeader(buff); herr != nil {
		return herr
	}
	return nil
}

func CheckPng(br *bufio.Reader) error {
	// PNG系のアップロードチェック
	buff, err := br.Peek(33)
	if err != nil {
		return err
	}
	if herr := png.CheckHeader(buff); herr != nil {
		return herr
	}
	return nil
}

func CheckGif(br *bufio.Reader) error {
	// GIF系のアップロードチェック
	buff, err := br.Peek(13)
	if err != nil {
		return err
	}
	if herr := gif.CheckHeader(buff); herr != nil {
		return herr
	}
	return nil
}

func CheckBmp(br *bufio.Reader) error {
	// BMP系のアップロードチェック
	buff, err := br.Peek(18)
	if err != nil {
		return err
	}
	if herr := bmp.CheckHeader(buff); herr != nil {
		return herr
	}
	return nil
}

func CheckTiff(br *bufio.Reader) error {
	// TIFF系のアップロードチェック
	buff, err := br.Peek(4)
	if err != nil {
		return err
	}
	if herr := tiff.CheckHeader(buff); herr != nil {
		return herr
	}
	return nil
}

func CheckWebp(br *bufio.Reader) error {
	// WEBP系のアップロードチェック
	buff, err := br.Peek(16)
	if err != nil {
		return err
	}
	if herr := webp.CheckHeader(buff); herr != nil {
		return herr
	}
	return nil
}

func MimeType(mt string, br *bufio.Reader) (ret Type, err error) {
	switch mt {
	case "image/jpeg", "image/jpg", "image/pjpeg":
		// JPEG系
		ret = TypeJpeg
	case "image/png", "image/x-png":
		// PNG系
		ret = TypePng
	case "image/gif":
		// gif
		ret = TypeGif
	case "image/bmp":
		// bmp
		ret = TypeBmp
	case "image/tiff":
		// tiff
		ret = TypeTiff
	case "image/webp":
		// webp
		ret = TypeWebp
	default:
		for ret = TypeUndefined + 1; ret < TypeMax; ret++ {
			err = ret.CheckFile(br)
			if err == nil {
				break
			}
		}
		if ret >= TypeMax {
			ret = TypeUndefined
			err = unknownImage
		}
		return
	}
	err = ret.CheckFile(br)
	return
}

func ExtType(ext string) (ret Type, err error) {
	switch ext {
	case "jpg":
		ret = TypeJpeg
	case "png":
		ret = TypePng
	case "gif":
		ret = TypeGif
	case "tiff":
		ret = TypeTiff
	case "webp":
		ret = TypeWebp
	case "bmp":
		ret = TypeBmp
	default:
		return TypeUndefined, unknownImage
	}
	return
}
