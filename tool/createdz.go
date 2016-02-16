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
package main

import (
	"../app/conf"
	"../app/img"
	"../app/library/deepzoom"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	imgroot := filepath.Join(conf.Conf.DataDir, "image")
	fi, err := ioutil.ReadDir(imgroot)
	if err != nil {
		log.Fatal(err)
	}
	for _, it := range fi {
		if it.IsDir() {
			convert(filepath.Join(imgroot, it.Name()))
		}
	}
}

func convert(p string) {
	dz := deepzoom.NewDeepZoom(256, true)
	fi, err := ioutil.ReadDir(p)
	if err != nil {
		log.Fatal(err)
	}
	for _, it := range fi {
		if it.IsDir() {
			continue
		}
		filename := it.Name()
		ext := filepath.Ext(filename)
		if len(ext) < 2 {
			continue
		}
		name := strings.TrimSuffix(filename, ext)
		num, err := strconv.ParseInt(name, 10, 64)
		if err != nil {
			continue
		}
		t, err := img.ExtType(ext[1:])
		if err != nil {
			continue
		}
		im, err := img.Decode(filepath.Join(p, filename), t)
		if err != nil {
			continue
		}
		dzip := deepzoom.CreateDziPathSystem(num)
		if _, err := os.Stat(dzip); err != nil {
			rect := im.Bounds()
			h := rect.Max.Y
			w := rect.Max.X
			log.Printf("start %s => %dx%d\n", filename, w, h)
			dz.MakeTiles(im, num)
			log.Printf("end %s\n", filename)
		}
	}
}
