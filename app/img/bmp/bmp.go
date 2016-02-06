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
package bmp

// BMP画像を扱う

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

func Send(w io.Writer, fp *os.File, b []byte) (err error) {
	_, err = io.Copy(w, fp)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func CheckHeader(buff []byte) error {
	if len(buff) < 18 {
		return errors.New("invalid bmp header length")
	}
	if string(buff[:2]) == "BM" {
		// OK
	} else {
		// NG
		return errors.New("invalid bmp format")
	}

	hsize := binary.LittleEndian.Uint32(buff[14:18])
	switch hsize {
	case 40, 12:
		// OK
	default:
		return errors.New("invalid bmp version")
	}
	return nil
}
