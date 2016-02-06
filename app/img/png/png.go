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
package png

// PNG画像を扱う

import (
	"../../conf"
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
)

var Signature = []byte{137, 80, 78, 71, 13, 10, 26, 10}
var IHDR = []byte{73, 72, 68, 82}   // IHDR
var TEXT = []byte{116, 69, 88, 116} // tEXt
var Comment = []byte{'U', 's', 'e', 'r', 'D', 'a', 't', 'a', 0}

func Send(w io.Writer, fp *os.File, b []byte) (err error) {
	br := bufio.NewReader(fp)
	buff := make([]byte, 33)
	_, rerr := br.Read(buff)
	if rerr != nil {
		// io.EOFだった場合も戻る
		return rerr
	}
	if herr := CheckHeader(buff); herr != nil {
		return herr
	}

	_, err = w.Write(buff)
	if err != nil && err != io.EOF {
		return err
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
	if len(buff) < 16 {
		return errors.New("invalid png header length")
	}
	if bytes.Equal(Signature, buff[:8]) == false {
		return errors.New("invalid png signature")
	}
	if binary.BigEndian.Uint32(buff[8:12]) != 13 {
		return errors.New("invalid png IHDR length")
	}
	if bytes.Equal(IHDR, buff[12:16]) == false {
		return errors.New("invalid png format")
	}
	return nil
}

func addUserData(b []byte) io.Reader {
	ret := &bytes.Buffer{}
	limit := uint(len(b))
	if limit > conf.Conf.UserDataSizeMax {
		limit = conf.Conf.UserDataSizeMax
	}

	// checksum
	h := crc32.NewIEEE()
	// length
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, uint32(limit)+uint32(len(Comment)))
	ret.Write(tmp)

	w := io.MultiWriter(ret, h)
	w.Write(TEXT)
	w.Write(Comment)
	w.Write(b[:limit])

	// crc32
	binary.BigEndian.PutUint32(tmp, h.Sum32())
	ret.Write(tmp)

	return ret
}
