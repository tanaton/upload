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
package conf

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"regexp"
)

// 設定一覧
const (
	Ver           = "0.06"
	OneYearSec    = 31104000   // だいたい一年
	OneYearSecStr = "31104000" // だいたい一年
)

type Config struct {
	Host           string
	ImageHost      string
	ThumbHost      string
	BaseUrl        string
	DataDir        string
	WebRootDir     string
	DziDir         string
	PassSalt       string
	SslCertPath    string
	SslPrivkeyPath string

	UrlLengthMax      uint
	TimeoutHandlerSec uint
	TimeoutReadSec    uint
	TimeoutWriteSec   uint
	HeaderSizeMax     uint
	MemorySizeMax     uint
	UserDataSizeMax   uint
	FormValueSizeMax  uint
	DeleteMinuteMax   uint
	ThumbPixelSize    uint
	PageSize          uint
	PaginateDefault   uint

	DBConnSize  uint
	DBIdleSize  uint
	DBUser      string
	DBName      string
	DBPass      string
	DBHost      string
	DBTable     string
	DBWaitTable string
	DBDelTable  string

	DBDescLengthMax  uint
	DBTagsLengthMax  uint
	DBPassLengthMax  uint
	DBStampLengthMax uint
}

var Conf Config
var RegSpace = regexp.MustCompile(`[\s　]+`)
var confErrMsg = errors.New("設定ファイルに誤りがあります。")

func init() {
	var p string
	if len(os.Args) >= 2 {
		p = os.Args[1]
	} else {
		p = "./config.json"
	}
	fp, err := os.Open(p)
	if err != nil {
		log.Fatal(err)
	}
	err = json.NewDecoder(fp).Decode(&Conf)
	if err != nil {
		log.Fatal(err)
	}

	// 必須
	if Conf.Host == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.ImageHost == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.ThumbHost == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.BaseUrl == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.DBUser == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.DBName == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.DBPass == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.DBHost == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.DBTable == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.DBWaitTable == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.DBDelTable == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.SslCertPath == "" {
		log.Fatal(confErrMsg)
	}
	if Conf.SslPrivkeyPath == "" {
		log.Fatal(confErrMsg)
	}

	// 任意
	if Conf.DataDir == "" {
		Conf.DataDir = "./data"
	}
	if Conf.WebRootDir == "" {
		Conf.WebRootDir = "./public_html"
	}
	if Conf.DziDir == "" {
		Conf.DziDir = "_dzi"
	}
	if Conf.PassSalt == "" {
		Conf.PassSalt = "きっと、澄みわたる朝色よりも、今、確かに此処にいるあなたと、出逢いの数だけのふれあいに、この手は繋がっている。"
	}
	if Conf.UrlLengthMax == 0 {
		Conf.UrlLengthMax = 256
	}
	if Conf.TimeoutHandlerSec == 0 {
		Conf.TimeoutHandlerSec = 300
	}
	if Conf.TimeoutReadSec == 0 {
		Conf.TimeoutReadSec = 250
	}
	if Conf.TimeoutWriteSec == 0 {
		Conf.TimeoutWriteSec = 250
	}
	if Conf.HeaderSizeMax == 0 {
		Conf.HeaderSizeMax = (32 << 10) // 32KB
	}
	if Conf.MemorySizeMax == 0 {
		Conf.MemorySizeMax = (8 << 20) // 8MB
	}
	if Conf.UserDataSizeMax == 0 {
		Conf.UserDataSizeMax = (63 << 10) // 63KB
	}
	if Conf.FormValueSizeMax == 0 {
		Conf.FormValueSizeMax = (10 << 10) // 10KB
	}
	if Conf.DeleteMinuteMax == 0 {
		Conf.DeleteMinuteMax = 50000
	}
	if Conf.ThumbPixelSize == 0 {
		Conf.ThumbPixelSize = 150
	}
	if Conf.PageSize == 0 {
		Conf.PageSize = 50
	}
	if Conf.PaginateDefault == 0 {
		Conf.PaginateDefault = 10
	}
	if Conf.DBConnSize == 0 {
		Conf.DBConnSize = 2
	}
	if Conf.DBIdleSize == 0 {
		Conf.DBIdleSize = 2
	}
	if Conf.DBDescLengthMax == 0 {
		Conf.DBDescLengthMax = 140
	}
	if Conf.DBTagsLengthMax == 0 {
		Conf.DBTagsLengthMax = 255
	}
	if Conf.DBPassLengthMax == 0 {
		Conf.DBPassLengthMax = 255
	}
	if Conf.DBStampLengthMax == 0 {
		Conf.DBStampLengthMax = 32
	}
}
