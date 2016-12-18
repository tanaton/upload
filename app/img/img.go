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

// 画像を扱う

import (
	"bufio"
	"errors"
	"github.com/tanaton/go-image/tiff"
	"golang.org/x/image/bmp"
	"golang.org/x/image/webp"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
)

const BufSize = 128 * 1024

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func Decode(path string, t Type) (image.Image, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	br := bufio.NewReader(fp)

	var im image.Image
	var imgerr error
	switch t {
	case TypeJpeg: // JPEG
		im, imgerr = jpeg.Decode(br)
	case TypePng: // PNG
		im, imgerr = png.Decode(br)
	case TypeGif: // GIF
		im, imgerr = gif.Decode(br)
	case TypeTiff: // TIFF
		im, imgerr = tiff.Decode(br)
	case TypeWebp: // WEBP
		im, imgerr = webp.Decode(br)
	case TypeBmp: // BMP
		im, imgerr = bmp.Decode(br)
	default:
		imgerr = errors.New("unknown image type")
	}
	if imgerr != nil {
		return nil, imgerr
	}
	return im, nil
}

func EncodeThumb(w io.Writer, im image.Image) error {
	bw := bufio.NewWriterSize(w, BufSize)
	err := jpeg.Encode(bw, im, &jpeg.Options{Quality: 80})
	if err != nil {
		return err
	}
	return bw.Flush()
}

func EncodeJpeg(w io.Writer, im image.Image) error {
	bw := bufio.NewWriterSize(w, BufSize)
	err := jpeg.Encode(bw, im, &jpeg.Options{Quality: 90})
	if err != nil {
		return err
	}
	return bw.Flush()
}

func EncodePng(w io.Writer, im image.Image) error {
	bw := bufio.NewWriterSize(w, BufSize)
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	err := enc.Encode(bw, im)
	if err != nil {
		return err
	}
	return bw.Flush()
}

func MustReadStamp(path string) image.Image {
	im, err := Decode(path, TypePng)
	if err != nil {
		panic(err)
	}
	return im
}
