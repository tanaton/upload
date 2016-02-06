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
package jpeg

// JPEG画像を扱う

import (
	"../../conf"
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const MarkAPP0 = 0xFFE0
const MarkAPP15 = 0xFFEF

var MarkSliceSOI = []byte{0xFF, 0xD8} // Start of Image
var MarkSliceCOM = []byte{0xFF, 0xFE} // コメントマーカー

func Send(w io.Writer, fp *os.File, b []byte) (err error) {
	br := bufio.NewReader(fp)
	buff := make([]byte, 2)
	_, rerr := br.Read(buff)
	if rerr != nil {
		// io.EOFだった場合も戻る
		return rerr
	}
	if herr := CheckHeader(buff); herr != nil {
		return herr
	}
	w.Write(buff)

	// APPnをスキップするための処理
	for {
		markbuff, mberr := br.Peek(4)
		if mberr != nil {
			return mberr
		}
		mark := binary.BigEndian.Uint16(markbuff[:2])
		if mark >= MarkAPP0 && mark <= MarkAPP15 {
			l := binary.BigEndian.Uint16(markbuff[2:])
			if l < 2 {
				return errors.New("invalid length")
			}
			_, err = io.CopyN(w, br, int64(l)+2)
			if err != nil && err != io.EOF {
				return err
			}
		} else if mark == 0xFFFF {
			// セグメント間が0xFFで埋まっている場合、最後の0xFFのみ有効とする
			br.ReadByte()
			_, err = w.Write([]byte{0xFF})
			if err != nil && err != io.EOF {
				return err
			}
		} else {
			break
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
	if len(buff) < 2 {
		return errors.New("invalid jpeg header length")
	}
	if bytes.Equal(MarkSliceSOI, buff) == false {
		return errors.New("invalid jpeg format")
	}
	return nil
}

func addUserData(b []byte) io.Reader {
	data := &bytes.Buffer{}
	data.Write(MarkSliceCOM)
	limit := uint(len(b))
	if limit > conf.Conf.UserDataSizeMax {
		limit = conf.Conf.UserDataSizeMax
	}
	combuff := make([]byte, 2)
	binary.BigEndian.PutUint16(combuff, uint16(limit+3)) // 文字列長 + length + ヌル文字
	data.Write(combuff)
	data.Write(b[:limit])
	data.WriteByte(0)
	return data
}
