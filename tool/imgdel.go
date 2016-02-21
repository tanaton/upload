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
package main

import (
	"../app/db"
	"../app/util"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var id int64
	if len(os.Args) >= 2 {
		var err error
		id, err = util.DecodeImageID(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("invalid arguments")
	}

	ext, dierr := db.Delete(id, "", false)
	if dierr != nil {
		// DB削除失敗
		log.Fatal(dierr)
	}
	// 削除成功
	// ファイルを削除する
	util.DeleteImageFile(id, ext)
	log.Println("success!!!")
}
