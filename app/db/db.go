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
package db

/* DB系 */

import (
	"../conf"
	"../library/deepzoom"
	"../util"
	"../util/webutil"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net"
	"strings"
	"time"
)

/*
CREATE DATABASE uploader DEFAULT CHARACTER SET utf8;

CREATE TABLE file_tags (
name VARCHAR(64) PRIMARY KEY
) ENGINE = Mroonga DEFAULT CHARSET = utf8 COLLATE = utf8_bin COMMENT = 'default_tokenizer "TokenDelimit"';

CREATE TABLE file_list(
id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
ext ENUM('jpg','png','gif','bmp','webp','webm', 'tiff'),
date DATETIME,
size INT UNSIGNED,
width INT UNSIGNED,
height INT UNSIGNED,
description VARCHAR(140),
passcode VARCHAR(32),
hash VARCHAR(32),
host VARCHAR(250),
tags TEXT COMMENT 'flags "COLUMN_VECTOR", type "file_tags"',
FULLTEXT INDEX tags_index(tags) COMMENT 'table "file_tags"',
FULLTEXT INDEX host_index(host),
INDEX hash_index(hash(8))
) ENGINE = Mroonga DEFAULT CHARSET utf8;

CREATE TABLE delete_wait_list(
id INT UNSIGNED PRIMARY KEY,
date DATETIME,
INDEX date_index(date)
) ENGINE = Innodb DEFAULT CHARSET utf8;

CREATE TABLE delete_file_list(
id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
up_id INT UNSIGNED NOT NULL,
ext ENUM('jpg','png','gif','bmp','webp','webm', 'tiff'),
date DATETIME,
up_date DATETIME,
description VARCHAR(140),
host VARCHAR(250),
FULLTEXT INDEX host_index(host)
) ENGINE = Mroonga DEFAULT CHARSET utf8;

SQLモードを変更して動作させる必要がある
SET @@GLOBAL.sql_mode='NO_ENGINE_SUBSTITUTION,STRICT_TRANS_TABLES,NO_BACKSLASH_ESCAPES';
*/

type Item struct {
	Id       string
	Ext      string
	Date     time.Time
	Size     int
	Width    int
	Height   int
	Desc     string
	PassCode bool
	Dzi      bool
	Tags     []string
}

type InsertItem struct {
	RemoteAddr string
	PassCode   string
	Ext        string
	Size       int64
	Width      int
	Height     int
	Desc       string
	Hash       string
	Tags       string
	DelMin     int64
}

type UpdateItem struct {
	PassCode string
	Desc     string
	Tags     string
	DelMin   int64
}

func Insert(ii InsertItem) (num int64, err error) {
	// 接続
	con, cerr := connect()
	if cerr != nil {
		return 0, cerr
	}
	tx, txerr := con.Begin()
	if txerr != nil {
		con.Close()
		return 0, txerr
	}
	defer func() {
		if err != nil {
			// ロールバック
			tx.Rollback()
		}
		con.Close()
	}()

	// 挿入
	var result sql.Result
	result, err = tx.Exec(createInsertQuery(ii))
	if err != nil {
		return 0, err
	}
	// 今のIDを取得
	num, err = result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if ii.DelMin > 0 {
		_, err = tx.Exec(createWaitInsertQuery(ii.DelMin, num))
		if err != nil {
			return 0, err
		}
	}
	// 反映させる
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return num, nil
}

func createInsertQuery(ii InsertItem) string {
	var host string
	addr, _, err := net.SplitHostPort(ii.RemoteAddr)
	if err == nil {
		hostlist, herr := webutil.LookupAddrDeadline(addr, time.Second*3)
		if herr == nil && len(hostlist) > 0 {
			host = hostlist[0]
		} else {
			host = addr
		}
	} else {
		host = ii.RemoteAddr
	}

	var passcode string
	if ii.PassCode != "" {
		passcode = util.CreateStorePass(ii.PassCode)
	}

	// クエリ生成
	return fmt.Sprintf("INSERT INTO %s(%s) VALUES('%s','%s',%d,%d,%d,'%s','%s','%s','%s','%s')",
		conf.Conf.DBTable,
		"ext,date,size,width,height,description,passcode,hash,host,tags",
		ii.Ext,
		time.Now().Format("2006-01-02 15:04:05"),
		ii.Size,
		ii.Width,
		ii.Height,
		sqlEscape(ii.Desc),
		sqlEscape(passcode),
		sqlEscape(ii.Hash),
		sqlEscape(host),
		sqlEscape(ii.Tags))
}

func createWaitInsertQuery(min, num int64) string {
	date := time.Now().Add(time.Minute * time.Duration(min))
	// クエリ生成
	return fmt.Sprintf("INSERT INTO %s(%s) VALUES(%d,'%s')",
		conf.Conf.DBWaitTable,
		"id,date",
		num,
		date.Format("2006-01-02 15:04:05"))
}

func GetPage(offset, limit int64) ([]Item, int64, error) {
	return GetPageTags(offset, limit, nil)
}

func GetPageTags(offset, limit int64, tags map[string]struct{}) ([]Item, int64, error) {
	if offset < 0 || limit < 0 || limit > 100 {
		return nil, 0, errors.New("invalid arguments")
	}
	con, err := connect()
	if err != nil {
		return nil, 0, err
	}
	defer con.Close()

	const selectOption = "SQL_CALC_FOUND_ROWS"
	const selectColum = "id,ext,date,size,width,height,description,passcode,tags"
	var query string
	if tags == nil || len(tags) == 0 {
		query = fmt.Sprintf("SELECT %s %s FROM %s ORDER BY id DESC LIMIT %d, %d",
			selectOption,
			selectColum,
			conf.Conf.DBTable,
			offset,
			limit)
	} else {
		taglist := make([]string, 0, len(tags))
		for key := range tags {
			taglist = append(taglist, sqlEscape(key))
		}
		query = fmt.Sprintf(`SELECT %s %s FROM %s WHERE MATCH(tags) AGAINST('+%s' IN BOOLEAN MODE) ORDER BY id DESC LIMIT %d, %d`,
			selectOption,
			selectColum,
			conf.Conf.DBTable,
			strings.Join(taglist, " +"),
			offset,
			limit)
	}
	// 取得
	rows, err := con.Query(query)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]Item, 0, limit)
	for rows.Next() {
		var it Item
		var id int64
		var tags string
		var passcode string
		err = rows.Scan(&id, &it.Ext, &it.Date, &it.Size, &it.Width, &it.Height, &it.Desc, &passcode, &tags)
		it.Id = util.EncodeImageID(id)
		it.PassCode = passcode != ""
		if tags != "" {
			it.Tags = strings.Split(tags, " ")
		}
		it.Dzi = deepzoom.CheckSize(it.Width, it.Height)
		list = append(list, it)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	searchmax := selectFoundRows(con)
	return list, searchmax, nil
}

func Update(id int64, key string, ui UpdateItem) error {
	// 接続
	con, err := connect()
	if err != nil {
		return err
	}
	defer con.Close()

	// 検索
	rows, err := con.Query(fmt.Sprintf("SELECT passcode FROM %s WHERE id=%d", conf.Conf.DBTable, id))
	if err != nil {
		return err
	}
	rows.Next()
	var itempass string
	err = rows.Scan(&itempass)
	rows.Close()
	if err = rows.Err(); err != nil {
		return err
	}

	if itempass == "" {
		return errors.New("can not be deleted")
	}
	pass := util.CreateStorePass(util.Utf8Substr(key, conf.Conf.DBPassLengthMax))
	if itempass != pass {
		return errors.New("illegal passcode")
	}

	var passcode string
	if ui.PassCode != "" {
		// 新しいパスコード
		passcode = util.CreateStorePass(ui.PassCode)
	} else {
		// 古いまま
		passcode = itempass
	}
	// 更新おｋ
	query := fmt.Sprintf("UPDATE %s SET description='%s',passcode='%s',tags='%s' WHERE id=%d",
		conf.Conf.DBTable,
		sqlEscape(ui.Desc),
		sqlEscape(passcode),
		sqlEscape(ui.Tags),
		id)
	_, err = con.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func Delete(id int64, key string, passflag bool) (string, error) {
	// 接続
	con, err := connect()
	if err != nil {
		return "", err
	}
	defer con.Close()

	// 検索
	rows, err := con.Query(fmt.Sprintf("SELECT id,ext,date,description,passcode,host FROM %s WHERE id=%d", conf.Conf.DBTable, id))
	if err != nil {
		return "", err
	}
	rows.Next()
	var d struct {
		id       int64
		ext      string
		date     time.Time
		desc     string
		passcode string
		host     string
	}
	err = rows.Scan(&d.id, &d.ext, &d.date, &d.desc, &d.passcode, &d.host)
	rows.Close()
	if err = rows.Err(); err != nil {
		return "", err
	}

	if passflag {
		// パスを確認する
		if d.passcode == "" {
			return "", errors.New("can not be deleted")
		}
		pass := util.CreateStorePass(util.Utf8Substr(key, conf.Conf.DBPassLengthMax))
		if d.passcode != pass {
			return "", errors.New("illegal passcode")
		}
	}

	// 削除する画像の情報を別テーブルに移す
	_, err = con.Exec(fmt.Sprintf("INSERT INTO %s(%s) VALUES(%d,'%s','%s','%s','%s','%s')",
		conf.Conf.DBDelTable,
		"up_id,ext,date,up_date,description,host",
		d.id,
		d.ext,
		time.Now().Format("2006-01-02 15:04:05"),
		d.date.Format("2006-01-02 15:04:05"),
		sqlEscape(d.desc),
		sqlEscape(d.host)))
	if err != nil {
		return "", err
	}

	// もう消してもおｋ
	// DBから削除
	_, err = con.Exec(fmt.Sprintf("DELETE FROM %s WHERE id=%d", conf.Conf.DBTable, d.id))
	if err != nil {
		return "", err
	}
	// 時間による削除テーブルの中身も削除
	con.Exec(fmt.Sprintf("DELETE FROM %s WHERE id=%d", conf.Conf.DBWaitTable, d.id))
	return d.ext, err
}

func GetWaitDeleteList() ([]int64, error) {
	con, err := connect()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	query := fmt.Sprintf(`SELECT id FROM %s WHERE date < '%s'`,
		conf.Conf.DBWaitTable,
		time.Now().Format("2006-01-02 15:04:05"))
	// 取得
	rows, err := con.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []int64{}
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		list = append(list, id)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func selectFoundRows(con *sql.DB) int64 {
	var searchmax int64
	result, err := con.Query("SELECT FOUND_ROWS()")
	if err == nil {
		result.Next()
		err = result.Scan(&searchmax)
		if err != nil {
			searchmax = 0
		}
		result.Close()
	}
	return searchmax
}

func sqlEscape(s string) string {
	return strings.Replace(s, "'", "''", -1)
}

func connect() (*sql.DB, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", conf.Conf.DBUser, conf.Conf.DBPass, conf.Conf.DBHost, conf.Conf.DBName))
	if err == nil {
		con.SetMaxOpenConns(int(conf.Conf.DBConnSize))
		con.SetMaxIdleConns(int(conf.Conf.DBIdleSize))
	}
	return con, err
}
