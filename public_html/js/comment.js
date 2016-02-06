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
(function(){
'use strict';

var regFileType = /^image\/(?:jpeg|png|gif|tiff|webp)/;
var upformQueue = [];
var updragQueue = [];
var userAgent = {
	FileReader: !!window.FileReader
};
var errorImageInfo = {
	type: "error",
	key: "error",
	text: "解析に失敗しました。"
};
var notsupportImageInfo = {
	type: "未対応",
	key: "未対応",
	text: "現在未対応のファイルです。"
};
var notfoundImageInfo = {
	type: "コメント無し",
	key: "コメント無し",
	text: "画像にコメントが含まれていないようです。"
};

function dropEvent(files){
	var elem = $('#list');
	elem.text('');
	updragQueue = [];

	setQueue(updragQueue, files);

	if(updragQueue.length > 0){
		updragQueue.sort(function(a, b){
			return +((a.file.name > b.file.name) || -1);
		});
		fileRead(elem, updragQueue);
	}
};

function setDropUpload(){
	var stop = function(e){
		// ブラウザのデフォルトの動作と、イベントの伝播をキャンセルする
		e.preventDefault();
		e.stopPropagation();
	};
	var dev = {
		start_: true,
		dndContainsFiles_: false,
		dispatch_: function(e){
			var url;
			var dt = e.originalEvent.dataTransfer;
			if(dt){
				if(dt.files && userAgent.FileReader){
					dropEvent(dt.files);
				} else {
					window.alert('ドロップ失敗');
				}
			} else {
				window.alert('ドロップ失敗');
			}
		}
	};

	// オーナーのイベント設定
	$(document).on({
		dragenter: function(e){
			// ファイルだったらデフォルトイベントをキャンセル
			var dt = e.originalEvent.dataTransfer;
			var l = dt.types.length;
			var i;
			if(dev.start_ && dt){
				for(i = 0; i < l && !dev.dndContainsFiles_; ++i){
					dev.dndContainsFiles_ = dt.types[i] === 'Files';
				}
			} else {
				dev.dndContainsFiles_ = false;
			}
			if(dev.dndContainsFiles_){
				e.preventDefault();
			}
		},
		dragover: function(e){
			var dt;
			if(dev.dndContainsFiles_){
				e.preventDefault();
				e.stopPropagation();
				dt = e.originalEvent.dataTransfer;
				dt.effectAllowed = 'all';
				dt.dropEffect = 'copy';
			}
		},
		dragstart: function(e){
			dev.start_ = false;
		},
		dragend: function(e){
			dev.start_ = true;
		},
		drop: function(e) {
			if(dev.dndContainsFiles_){
				e.preventDefault();
				e.stopPropagation();
				dev.dispatch_(e);
			}
			dev.start_ = true;
		}
	});
};

function setMultiUpload(){
	$('#input-file')
	.attr('multiple', 'multiple')
	.on('change', upformLoadImage);
};

function upformLoadImage() {
	var elem = $('#list');

	elem.text('');
	upformQueue = [];

	setQueue(upformQueue, this.files);

	if(upformQueue.length > 0){
		fileRead(elem, upformQueue);
	}
};

function fileRead(elem, queue){
	var i;
	var l;
	var reader;
	for(i = 0, l = queue.length; i < l; ++i){
		reader = new window.FileReader();
		// 要素を事前に用意する
		elem.append(queue[i].elem);
		// 読み込みが完了した時のイベントを設定
		$(reader).one('load', (function(q){
			return function(e){
				var data;
				var reader;
				var u8arr = new Uint8Array(e.target.result);

				switch(q.file.type){
				case "image/jpeg":
					data = jpegAnalyze(u8arr);
					break;
				case "image/png":
					data = pngAnalyze(u8arr);
					break;
				case "image/gif":
					data = gifAnalyze(u8arr);
					break;
				case "image/webp":
					data = webpAnalyze(u8arr);
					break;
				case "image/tiff":
					data = tiffAnalyze(u8arr);
					break;
				default:
					data = [notsupportImageInfo];
					break;
				}
				if(data.length === 0){
					data = [notfoundImageInfo];
				}

				reader = new window.FileReader();
				$(reader).one('load', function(e){
					// 画像情報要素を作成、表示する
					addImageInfoElem(e, q, data);
				});
				// ファイルをデータURLとして読み込む
				reader.readAsDataURL(q.file);
			}
		})(queue[i]));
		// ファイルをバイナリ配列として読み込む
		reader.readAsArrayBuffer(queue[i].file);
	}
};

function addImageInfoElem(e, q, data){
	var it;
	var elem;
	var i = 0;
	var l = data.length;
	var txt;
	var doc;
	var output;
	var err;
	for(; i < l; i++){
		it = data[i];
		txt = $("<div></div>").text(it.text).html();
		try {
			doc = JSON.parse(txt);
			output = "<pre>" + JSON.stringify(doc, null, 4) + "</pre>";
		} catch(err){
			if(data.xml && DOMParser){
				try {
					doc = (new DOMParser()).parseFromString(txt, "application/xml");
					if((doc != null) && ($(doc).find("parsererror").length === 0)){
						output = "<pre>" + formatXml(doc) + "</pre>";
					} else {
						output = txt;
					}
				} catch(err) {
					output = txt;
				}
			} else {
				output = txt;
			}
		}
		elem = $("<div></div>").addClass('media').html(
			'<a href="#" class="media-left"><img src="' + e.target.result + '" width="64" height="64"></a>'
			+ '<div class="media-body">'
			+ '<h4 class="media-heading">' + q.file.name + '</h4>'
			+ '<p>ファイルの形式：' + q.file.type + '</p>'
			+ '<p>ファイルサイズ：' + q.file.size + ' Byte</p>'
			+ '<p>コメントタイプ：' + $("<div></div>").text(it.type).html() + '</p>'
			+ '<p>コメントキー：' + $("<div></div>").text(it.key).html() + '</p>'
			+ '<p>コメント内容：' + output + '</p>'
			+ '</div>'
		);
		q.elem.append(elem);
	}
};

function setQueue(queue, files){
	var i;
	var l = files.length;
	var file;

	for(i = 0; i < l; ++i){
		file = files[i];
		if(!regFileType.test(file.type)){
			continue;
		}
		queue.push({
			elem: $("<div></div>"),
			file: file
		});
	}
};

function jpegAnalyze(u8buf){
	var data = [];
	var mark;
	var size;
	var txtsize;
	var l = u8buf.length;
	var index = 0;
	var dvbuf = new DataView(u8buf.buffer);

	while(index < l){
		mark = dvbuf.getUint16(index, false);	// ビッグエンディアン
		index += 2;

		switch(mark){
		case 0xFFD8:	// SOI
			break;
		case 0xFFD9:	// EOI
		case 0xFFDA:	// SOS
			index = l;
			break;
		case 0xFFFF:	// パディング
			index--;
			break;
		case 0xFFFE:	// コメント
			size = dvbuf.getUint16(index, false);	// ビッグエンディアン
			if(size >= 2){
				txtsize = size - 2;
				if((size >= 3) && (dvbuf.getUint8(index + size - 1) === 0)){
					txtsize--;
				}
			} else {
				txtsize = 0;
			}
			data.push({
				type: "COM",
				key: "",
				text: String.fromCharCode.apply(null, new Uint8Array(u8buf.buffer, index + 2, txtsize))
			});
			index += size;
			break;
		default:
			if(((mark >> 8) & 0xFF) === 0xFF){
				size = dvbuf.getUint16(index, false);	// ビッグエンディアン
				index += size;
			} else {
				// エラー
				index = l;
				data.push(errorImageInfo);
			}
			break;
		}
	}
	return data;
};

function pngAnalyze(u8buf){
	var data = [];
	var size;
	var ctype;
	var crc;
	var tmp;
	var zflag;
	var l = u8buf.length;
	var index = 8;
	var dvbuf = new DataView(u8buf.buffer);

	while(index < l){
		size = dvbuf.getUint32(index, false);	// ビッグエンディアン
		index += 4;
		ctype = dvbuf.getUint32(index, false);
		index += 4;
		switch(ctype){
		case 0x74455874:	// tEXt
			tmp = pngKeyword(dvbuf, l, index);
			data.push({
				type: "tEXt",
				key: tmp[0],
				text: String.fromCharCode.apply(null, new Uint8Array(u8buf.buffer, tmp[1], size - (tmp[1] - index)))
			});
			index = tmp[1];
			break;
		case 0x7A545874:	// zTXt
			tmp = pngKeyword(dvbuf, l, index);
			zflag = dvbuf.getUint8(tmp[1], false);
			tmp[1]++;
			data.push({
				type: "zTXt",
				key: tmp[0],
				text: (zflag === 0) ? pngZlibInflate(new Uint8Array(u8buf.buffer, tmp[1], size - (tmp[1] - index))) : "コメントを解凍できませんでした。"
			});
			index = tmp[1];
			break;
		default:
			break;
		}
		index += size;
		//crc = dvbuf.getUint32(index, false);
		index += 4;
	}
	return data;
};

function pngKeyword(dvbuf, l, index){
	var tmp;
	var keyword = [];
	var klen = index + 80;
	for(;index < klen && index < l; index++){
		tmp = dvbuf.getUint8(index, false);
		if(tmp === 0){
			// キーワード終端
			index++;
			break;
		}
		keyword.push(tmp);
	}
	return [String.fromCharCode.apply(null, keyword), index];
};

function pngZlibInflate(u8bufsub){
	var inflate = new Zlib.Inflate(u8bufsub);
	var plain = inflate.decompress();
	return plain;
};

function gifAnalyze(u8buf){
	var data = [];
	var mark;
	var tmp;
	var txt;
	var l = u8buf.length;
	var index = 13;
	var dvbuf = new DataView(u8buf.buffer);

	switch(String.fromCharCode.apply(null, new Uint8Array(u8buf.buffer, 0, 6))){
	case "GIF87a":
	case "GIF89a":
		break;
	default:
		return [errorImageInfo];
	}
	tmp = dvbuf.getUint8(10);
	if((tmp >> 7) === 1){
		// パレットの読み飛ばし
		index += gifAnalyze.ColorSizeTable[tmp & 0x07] * 3;
	}

	while(index < l){
		mark = dvbuf.getUint8(index);
		index++;

		switch(mark){
		case 0x2C:	// 画像データ
			index += 8;
			tmp = dvbuf.getUint8(index);
			index++;
			if((tmp >> 7) === 1){
				// パレットの読み飛ばし
				index += gifAnalyze.ColorSizeTable[tmp & 0x07] * 3;
			}
			index++;	// LZW Minimum Code Sizeを飛ばす
			index += gifSkipBlock(index, l, dvbuf);
			break;

		case 0x21:	// 拡張ブロック
			mark = dvbuf.getUint8(index);
			index++;

			switch(mark){
			case 0x01:	// Plain Text Extension
				index += 13;	// 0x0C + 1
				// fall through
			case 0xFE:	// Comment Extension
				txt = "";
				while(index < l){
					tmp = dvbuf.getUint8(index);
					index++;
					if(tmp === 0){
						break;
					}
					txt += String.fromCharCode.apply(null, new Uint8Array(u8buf.buffer, index, tmp));
					index += tmp;
				}
				data.push({
					type: (mark === 0xFE) ? "Comment Extension" : "Plain Text Extension",
					key: "",
					text: txt
				});
				break;

			case 0xF9:	// Graphic Control Extension
			case 0xFF:	// Application Extension
				tmp = dvbuf.getUint8(index);
				index++;
				if(tmp === 0){
					index = l;
					data.push(errorImageInfo);
					break;
				}
				index += tmp;
				index += gifSkipBlock(index, l, dvbuf);
				break;

			default:	// 謎のブロック
				index = l;
				data.push(errorImageInfo);
				break;
			}
			break;

		case 0x3B:	// 終端
			index = l;
			break;

		default:	// 謎のブロック
			index = l;
			data.push(errorImageInfo);
			break;
		}
	}
	return data;
};
gifAnalyze.ColorSizeTable = [
	1 << 1,
	1 << 2,
	1 << 3,
	1 << 4,
	1 << 5,
	1 << 6,
	1 << 7,
	1 << 8
];

function gifSkipBlock(index, l, dvbuf){
	var tmp;
	var old = index;
	// ブロックの読み飛ばし
	while(index < l){
		tmp = dvbuf.getUint8(index);
		index++;
		if(tmp === 0){
			break;
		}
		index += tmp;
	}
	return index - old;
};

function webpAnalyze(u8buf){
	var data = [];
	var mark;
	var size;
	var l = u8buf.length;
	var index = 12;
	var dvbuf = new DataView(u8buf.buffer);

	if(String.fromCharCode.apply(null, new Uint8Array(u8buf.buffer, 0, 4)) !== "RIFF"){
		return [errorImageInfo];
	}
	if(String.fromCharCode.apply(null, new Uint8Array(u8buf.buffer, 8, 4)) !== "WEBP"){
		return [errorImageInfo];
	}

	while(index < l){
		mark = String.fromCharCode.apply(null, new Uint8Array(u8buf.buffer, index, 4));
		index += 4;

		switch(mark){
		case "XMP ":	// XMPメタデータ
			size = dvbuf.getUint32(index, true);	// リトルエンディアン
			index += 4;
			data.push({
				xml: true,
				type: "XMP",
				key: "",
				text: String.fromCharCode.apply(null, new Uint8Array(u8buf.buffer, index, size))
			});
			index += size;
			break;

		case "VP8 ":
		case "VP8L":
		case "VP8X":
		case "ANIM":
		case "ANMF":
		case "ALPH":
		case "ICCP":
		case "EXIF":
			size = dvbuf.getUint32(index, true);	// リトルエンディアン
			index += size + 4;
			break;

		default:	// 未定義の何か
			index = l;
			data.push(errorImageInfo);
			break;
		}
		// パディングを読み飛ばす
		while((index < l) && (dvbuf.getUint8(index) === 0)){
			index++;
		}
	}
	return data;
};

function tiffAnalyze(u8buf){
	var data = [];
	var size;
	var ec;
	var bin;
	var l = u8buf.length;
	var index = 4;
	var ifdei = 0;
	var ifdindex = 0;
	var dvbuf = new DataView(u8buf.buffer);

	switch(String.fromCharCode.apply(null, new Uint8Array(u8buf.buffer, 0, 2))){
	case "MM":
		bin = false;
		break;
	case "II":
		bin = true;
		break;
	default:
		return [errorImageInfo];
	}
	// 生命、宇宙、そして万物についての究極の疑問の答え
	if(dvbuf.getUint16(2, bin) !== 42){
		return [errorImageInfo];
	}

	while(index < l){
		// IFD
		index = dvbuf.getUint32(index, bin);
		if(index === 0){
			break;
		}
		// エントリ数
		ec = dvbuf.getUint16(index, bin);
		ifdei = 0;
		index += 2;
		while(ifdei < ec){
			ifdindex = index + (ifdei * 12);
			ifdei++;
			switch(dvbuf.getUint16(ifdindex, bin)){
			case 270:	// ImageDescription
				if(dvbuf.getUint16(ifdindex + 2, bin) !== 2){
					break;	// ASCIIではないので抜ける
				}
				size = dvbuf.getUint32(ifdindex + 4, bin);
				if(size > 4){
					data.push({
						type: "ImageDescription",
						key: "",
						text: String.fromCharCode.apply(null, new Uint8Array(u8buf.buffer, dvbuf.getUint32(ifdindex + 8, bin), size - 1))
					});
				} else if(size > 1){
					data.push({
						type: "ImageDescription",
						key: "",
						text: String.fromCharCode.apply(null, new Uint8Array(u8buf.buffer, ifdindex + 8, size - 1))
					});
				} else {
					// 特に何もしない
				}
				break;
			default:	// スキップする
				break;
			}
		}
		index += ec * 12;
	}
	return data;
};

// http://stackoverflow.com/questions/376373/pretty-printing-xml-with-javascript
function formatXml(xml){
	var formatted = "";
	var reg = /(>)(<)(\/*)/g;
	var pad = 0;

	xml = xml.replace(reg, '$1\r\n$2$3');
	jQuery.each(xml.split('\r\n'), function(index, node){
		var indent = 0;
		var padding = "";
		var i;

		if(node.match(/.+<\/\w[^>]*>$/)){
			indent = 0;
		} else if(node.match(/^<\/\w/)){
			if(pad != 0){
				pad -= 1;
			}
		} else if(node.match(/^<\w[^>]*[^\/]>.*$/)){
			indent = 1;
		} else {
			indent = 0;
		}

		for(i = 0; i < pad; i++){
			padding += '  ';
		}

		formatted += padding + node + "\n";
		pad += indent;
	});

	return formatted;
}

$(function(){
	if(userAgent.FileReader){
		setDropUpload();
		setMultiUpload();
	}
});

})();
