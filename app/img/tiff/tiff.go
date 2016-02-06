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
package tiff

// TIFF画像を扱う

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
	formatMotorola
	formatIntel
)

const TiffVer = 42

func Send(w io.Writer, fp *os.File, b []byte) (err error) {
	st, sterr := fp.Stat()
	if sterr != nil {
		return sterr
	}
	size := st.Size()
	br := bufio.NewReader(fp)
	var tmp []byte
	tmp, err = br.Peek(8)
	if err != nil {
		// io.EOFだった場合も戻る
		return err
	}
	var bin binary.ByteOrder
	switch string(tmp[:2]) {
	case "MM":
		bin = binary.BigEndian
	case "II":
		bin = binary.LittleEndian
	default:
		return errors.New("invalid tiff format")
	}
	// 最初のIFDまでのデータを転送
	ptr := bin.Uint32(tmp[4:8])
	_, err = io.CopyN(w, br, int64(ptr))
	if err != nil && err != io.EOF {
		return err
	}

	// 最初のIFDのデータ数を改竄する
	ec := make([]byte, 4)
	_, err = br.Read(ec[:2])
	if err != nil {
		// io.EOFだった場合も戻る
		return err
	}
	cntentry := bin.Uint16(ec[:2])
	bin.PutUint16(ec[2:4], cntentry+1)
	_, err = w.Write(ec[2:4])
	if err != nil && err != io.EOF {
		return err
	}
	// 既存エントリのIFDポインタを必要ならばずらす
	entrybuf := &bytes.Buffer{}
	_, err = io.CopyN(entrybuf, br, int64(cntentry)*12)
	if err != nil {
		// io.EOFだった場合も戻る
		return err
	}
	entry := entrybuf.Bytes()
	ifdmaxptr := ptr + (uint32(cntentry) * 12)
	for i := 0; i < int(cntentry); i++ {
		j := i * 12
		ifdPointerCheck(entry[j : j+12], ifdmaxptr, bin)
	}
	// 読みだしたデータを転送
	_, err = io.Copy(w, entrybuf)
	if err != nil && err != io.EOF {
		return err
	}

	limit := uint(len(b))
	if limit > conf.Conf.UserDataSizeMax {
		limit = conf.Conf.UserDataSizeMax
	}

	// 自前のIFDエントリ開始
	// テキストエントリ
	bin.PutUint16(tmp[:2], 270)
	w.Write(tmp[:2])
	// ASCIIコード
	bin.PutUint16(tmp[:2], 2)
	w.Write(tmp[:2])
	// データ長(null文字含む)
	bin.PutUint32(tmp[:4], uint32(limit+1))
	w.Write(tmp[:4])
	// データ位置へのポインタ
	bin.PutUint32(tmp[:4], uint32(size+12))
	w.Write(tmp[:4])

	// 次のIFDへのポインタ
	// 複数枚のTIFFには面倒なので対応しないことにする
	_, err = br.Read(tmp[:4]) // 空読み
	if err != nil {
		// io.EOFだった場合も戻る
		return err
	}
	_, err = w.Write([]byte{0, 0, 0, 0})
	if err != nil && err != io.EOF {
		return err
	}

	// ファイル内容の残りを全部転送
	_, err = io.Copy(w, br)
	if err != nil && err != io.EOF {
		return err
	}

	// 肝心のデータを転送
	_, err = w.Write(b[:limit])
	if err != nil && err != io.EOF {
		return err
	}
	_, err = w.Write([]byte{0})
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func ifdPointerCheck(entry []byte, ifdmaxptr uint32, bin binary.ByteOrder) {
	t := bin.Uint16(entry[2:4])
	s := bin.Uint32(entry[4:8])
	p := bin.Uint32(entry[8:12])
	switch t {
	case 1, 2, 6, 7: // 1byte
		break
	case 3, 8: // 2byte
		s *= 2
	case 4, 9, 11: // 4byte
		s *= 4
	default:
		s *= 8
	}
	if s > 4 && p > ifdmaxptr {
		// 現IFD領域よりも後ろのデータを指している場合はIFD分ずらす
		bin.PutUint32(entry[8:12], p+12)
	}
}

func CheckHeader(buff []byte) error {
	if len(buff) < 4 {
		return errors.New("invalid tiff header length")
	}
	var format int
	switch string(buff[:2]) {
	case "MM":
		format = formatMotorola
		if binary.BigEndian.Uint16(buff[2:4]) != TiffVer {
			return errors.New("invalid tiff format")
		}
	case "II":
		format = formatIntel
		if binary.LittleEndian.Uint16(buff[2:4]) != TiffVer {
			return errors.New("invalid tiff format")
		}
	default:
		format = formatNone
	}
	if format == formatNone {
		return errors.New("invalid tiff format")
	}
	return nil
}
