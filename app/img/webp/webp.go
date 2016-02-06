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
package webp

// WEBP画像を扱う

import (
	"../../conf"
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const (
	formatNone = iota
	formatVP8
	formatVP8L
	formatVP8X
)

var XMPChunk = []byte{'X', 'M', 'P', ' '}

func Send(w io.Writer, fp *os.File, b []byte) (err error) {
	br := bufio.NewReader(fp)
	buff := make([]byte, 16)
	_, rerr := br.Read(buff)
	if rerr != nil {
		// io.EOFだった場合も戻る
		return rerr
	}
	if herr := CheckHeader(buff); herr != nil {
		return herr
	}

	bin := binary.LittleEndian
	ud := addUserData(b)
	bin.PutUint32(buff[4:8], bin.Uint32(buff[4:8]) + uint32(ud.Len()))
	_, err = w.Write(buff)
	if err != nil && err != io.EOF {
		return err
	}
	_, err = io.Copy(w, br)
	if err != nil && err != io.EOF {
		return err
	}
	// webpの場合、後ろにくっつける
	_, err = io.Copy(w, ud)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func CheckHeader(buff []byte) error {
	if len(buff) < 16 {
		return errors.New("invalid webp header length")
	}
	var format int
	switch string(buff[8:16]) {
	case "WEBPVP8 ":
		format = formatVP8
	case "WEBPVP8L":
		format = formatVP8L
	case "WEBPVP8X":
		format = formatVP8X
	default:
		format = formatNone
	}
	if string(buff[:4]) != "RIFF" || format == formatNone {
		return errors.New("invalid webp format")
	}
	return nil
}

func addUserData(b []byte) *bytes.Buffer {
	ret := &bytes.Buffer{}
	limit := uint(len(b))
	if limit > conf.Conf.UserDataSizeMax {
		limit = conf.Conf.UserDataSizeMax
	}

	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(limit))

	ret.Write(XMPChunk)
	ret.Write(tmp)
	ret.Write(b[:limit])
	if limit%2 == 1 {
		// 奇数だったら0でパディング
		ret.WriteByte(0)
	}
	return ret
}
