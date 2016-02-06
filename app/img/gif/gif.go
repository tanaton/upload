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
package gif

// GIF画像を扱う

import (
	"../../conf"
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
)

const CommentBlockSize = 255

var Version87a = []byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61} // GIF87a
var Version89a = []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61} // GIF89a
var CommentBlockHead = []byte{0x21, 0xfe}
var ColorSizeTable = [8]int{2, 4, 8, 16, 32, 64, 128, 256}

func Send(w io.Writer, fp *os.File, b []byte) (err error) {
	br := bufio.NewReader(fp)
	buff := make([]byte, 13)
	_, rerr := br.Read(buff)
	if rerr != nil {
		// io.EOFだった場合も戻る
		return rerr
	}
	if herr := CheckHeader(buff); herr != nil {
		return herr
	}
	var cbuff []byte
	if (buff[10] >> 7) == 1 {
		cbuff = make([]byte, ColorSizeTable[buff[10]&0x07]*3)
		_, cerr := br.Read(cbuff)
		if cerr != nil {
			return cerr
		}
	}

	_, err = w.Write(buff)
	if err != nil && err != io.EOF {
		return err
	}
	if cbuff != nil {
		_, err = w.Write(cbuff)
		if err != nil && err != io.EOF {
			return err
		}
	}
	_, err = io.Copy(w, addUserData(b))
	if err != nil && err != io.EOF {
		return err
	}
	_, err = io.Copy(w, br)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func CheckHeader(buff []byte) error {
	if len(buff) < 6 {
		return errors.New("invalid gif header length")
	}
	if bytes.Equal(Version89a, buff[:6]) || bytes.Equal(Version87a, buff[:6]) {
		// OK
	} else {
		// NG そんなバージョンのGIFは無い
		return errors.New("invalid gif version")
	}
	return nil
}

func addUserData(b []byte) io.Reader {
	var i uint = 0

	ret := &bytes.Buffer{}
	limit := uint(len(b))
	if limit > conf.Conf.UserDataSizeMax {
		limit = conf.Conf.UserDataSizeMax
	}

	ret.Write(CommentBlockHead)
	for i < limit {
		var num uint = CommentBlockSize
		if num > (limit - i) {
			num = (limit - i)
		}
		ret.WriteByte(byte(num))
		ret.Write(b[i : i+num])
		i += num
	}
	ret.WriteByte(0)
	return ret
}
