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

var regFileType = /^image\/(?:jpeg|png|gif|tiff|webp|bmp)/;
var upformQueue = [];
var updragQueue = [];
var userAgent = {
	Touch: typeof document.ontouchstart != "undefined",
	Mobile: typeof window.orientation != "undefined",
	Pointer: window.navigator.pointerEnabled,
	MSPoniter: window.navigator.msPointerEnabled,
	FileReader: !!window.FileReader,
	DragAndDrop: 'ondrag' in document
};
var hidePopover;
var imageUrl;
var thumbUrl;
var viewerUrl;
if("https:" === document.location.protocol){
	imageUrl = 'https://kntn.org/i/';
	thumbUrl = 'https://kntn.org/t/';
} else {
	imageUrl = 'http://i.kntn.org/';
	thumbUrl = 'http://t.kntn.org/';
}
viewerUrl = 'https://kntn.org/v/';

hidePopover = (function(){
	var wait = true;
	var list = [];
	function hidePopoverLocal(){
		var i;
		var l = list.length;
		wait = false;
		for(i = 0; i < l; ++i){
			list[i].popover('hide');
		}
		list = [];
		wait = true;
	}

	$(document)
	.on('click', hidePopoverLocal)
	.on('click', '.popover', function(e){
		// イベントの伝搬は止めるけど、標準のイベントは止めない
		//e.preventDefault();
		e.stopPropagation();
	});

	$('#image-main-area')
	.on('show.bs.popover', '[data-toggle="popover"]', hidePopoverLocal)
	.on('shown.bs.popover', '[data-toggle="popover"]', function(){
		list.push($(this));
	})
	.on('hide.bs.popover', '[data-toggle="popover"]', function(){
		return !wait;
	});

	return hidePopoverLocal;
})();

function setPopover(page){
	var pop;
	if(page){
		pop = $('.page' + page).filter('[data-toggle="popover"]');
	} else {
		pop = $('[data-toggle="popover"]');
	}
	// 要素を絞って初期設定する
	pop.popover({
		content: function(){
			var imgid = $(this).attr('data-imgid');
			return $('#img-caption-' + imgid).html();
		},
		trigger: 'hover',
		placement: 'bottom',
		animation: false,
		html: true,
		template: '<div class="popover" role="tooltip" style="margin-top:0;"><div class="arrow"></div><h3 class="popover-title"></h3><div class="popover-content"></div></div>'
	});
};

function setTooltip(){
	$('[data-toggle="tooltip"]').tooltip();
};

function setEventGlobal(){
	if(userAgent.Touch && userAgent.Mobile && !(userAgent.Pointer || userAgent.MSPoniter)){
		$(".drag-and-drop").html("");
	}

	$('#updrag')
	.on('shown.bs.modal', hidePopover)
	.on('hidden.bs.modal', function(e){
		// リロードする
		location.reload();
	});

	$('#update')
	.on('show.bs.modal', function(e){
		var button = $(e.relatedTarget);
		var imgid = button.attr('data-imgid');
		$('#update-form').attr('action', '/api/update?id=' + imgid + '&jump=true');
	})
	.on('shown.bs.modal', hidePopover);

	$('#delete')
	.on('show.bs.modal', function(e){
		var button = $(e.relatedTarget);
		var imgid = button.attr('data-imgid');
		$('#delete-input-id').attr('value', imgid);
	})
	.on('shown.bs.modal', hidePopover);
};

function setEventScrollLast(){
	var tid;
	var tags = '';
	var searchQuery = getUrlVars();
	var nowPageNo = 1;
	var nowTags = searchQuery['tag'];
	if(searchQuery['p']){
		nowPageNo = +(searchQuery['p'][0] || 1);
	}

	if(nowTags){
		tags = 'tag=' + nowTags.join('&tag=') + '&';
	}
	function getPage(){
		var root = $('#image-main-area');
		// タイマー停止
		clearInterval(tid);
		if(nowPageNo >= (+root.attr('data-pagemax'))){
			// これ以上ページが無い事が分かっている場合は取得しない
			return;
		}
		// 取得
		$.ajax({
			url: '/api/list?' + tags + 'p=' + (nowPageNo + 1),
			type: 'get',
			contentType: false,
			processData: false,
			dataType: 'json',
			timeout: 10000
		}).then(function(data, status, jqxhr){
			// 成功
			// 画像の追加
			var ret = insertImage(data, nowPageNo);
			setPopover(nowPageNo);
			nowPageNo++;
			if(ret){
				// 次に読み込むデータがある場合
				// 次のイベントを設定
				setTimer();
			}
		}, function(){
			// 失敗
			var elem = $('<div></div>').addClass('alert alert-warning alert-dismissible').attr('role', 'alert');
			elem.html(
				'<button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button>'
				+ '<strong>警告</strong> ページの自動読み込みに失敗しました。'
			);
			root.after(elem);
		});
	}
	function setTimer(){
		tid = setInterval(function(){
			getScrollTop();
		}, 1000);
	}
	function getScrollTop(){
		// ページの大体一番下に到着したか確認
		var lastTop = $(document.body).height() - 50;
		var scrollBottom = $(window).scrollTop() + $(window).height();
		if(scrollBottom >= lastTop){
			// イベント発生
			getPage();
		}
	}
	setTimer();
};

function dropEvent(files){
	var elem = $('#updrag-list');
	elem.text('');
	updragQueue = [];
	// イベントの削除
	$('#updrag-submit').off('click', updragSubmit);

	setQueue(updragQueue, files);

	if(updragQueue.length > 0){
		updragQueue.sort(function(a, b){
			return (a.file.name > b.file.name) ? 1 : -1;
		});

		$('#updrag').modal('show');
		fileRead(elem, updragQueue);

		// 一回だけ
		$('#updrag-submit').one('click', updragSubmit);
	}
};

function updragSubmit(e){
	fileUpload(e, updragQueue, {
		tags: $('#updrag-input-tags').val(),
		passcode: $('#updrag-input-pass').val(),
		description: $('#updrag-input-comment').val(),
		stamp: $('#updrag-input-stamp').val(),
		stamp_position: $('#updrag-input-stamp-position').val(),
		thumb_change: $('#updrag-input-thumb-change').val(),
		delete_wait_minute: $('#updrag-input-delete-wait').val()
	}, function(){
		window.alert('アップロードが完了しました。');
		location.reload();
	}, function(){
		window.alert('アップロード途中で失敗しました。');
	});
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
					window.alert('アップロード失敗');
				}
			} else {
				window.alert('アップロード失敗');
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
				// ドロップする前にフォームを表示
				$('#updrag').modal('show');
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
	$('#upform-input-file')
	.attr('multiple', 'multiple')
	.on('change', upformLoadImage);
};

function upformLoadImage() {
	var elem = $('#upform-list');

	elem.text('');
	upformQueue = [];

	setQueue(upformQueue, this.files);

	if(upformQueue.length > 0){
		//upformQueue.sort(function(a, b){
		//	return (a.file.name > b.file.name) ? 1 : -1;
		//});
		fileRead(elem, upformQueue);

		$('.upform-submit').one('click', upformSubmit);
	}
};

function upformSubmit(e){
	$('#upform').one('hidden.bs.modal', function(e){
		// リロードする
		location.reload();
	});

	fileUpload(e, upformQueue, {
		tags: $('#upform-input-tags').val(),
		passcode: $('#upform-input-pass').val(),
		description: $('#upform-input-comment').val(),
		stamp: $('#upform-input-stamp').val(),
		stamp_position: $('#upform-input-stamp-position').val(),
		thumb_change: $('#upform-input-thumb-change').val(),
		delete_wait_minute: $('#upform-input-delete-wait').val()
	}, function(){
		window.alert('アップロードが完了しました。');
		location.reload();
	}, function(){
		$('.upform-submit').one('click', upformSubmit);
		window.alert('アップロード途中で失敗しました。');
	});
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
				q.elem.addClass('media').html(
					'<a href="#" class="media-left"><img src="' + e.target.result + '" width="64" height="64"></a>'
					+ '<div class="media-body">'
					+ '<h4 class="media-heading">' + q.file.name + '</h4>'
					+ '<p>ファイルの形式：' + q.file.type + '</p>'
					+ '<p>ファイルサイズ：' + q.file.size + ' Byte</p>'
					+ '</div>'
				);
			}
		})(queue[i]));
		// ファイルをデータURLとして読み込む
		reader.readAsDataURL(queue[i].file);
	}
};

function fileUpload(e, queue, config, okf, ngf){
	var key;
	e.preventDefault();
	e.stopPropagation();

	if(queue.length <= 0){
		ngf();
		return;
	}

	function request(){
		var it;
		var form = new window.FormData();

		if(queue.length <= 0){
			okf();
			return;
		}

		it = queue.shift();
		form.append('uploadfile', it.file, it.file.name);
		for(key in config){
			if(config.hasOwnProperty(key)){
				form.append(key, config[key]);
			}
		}

		return $.ajax({
			url: '/api/upload',
			type: 'POST',
			contentType: false,
			processData: false,
			data: form,
			dataType: 'html',
			timeout: 30000
		}).then(function(data, status, jqxhr){
			var imgid = jqxhr.getResponseHeader('X-Iill-FileID');
			var ext = jqxhr.getResponseHeader('X-Iill-FileExt');
			it.elem.text('');
			it.elem.html('<a href="' + imageUrl + imgid + '.' + ext + '" class="media-left" target="_blank">'
				+ '<img src="' + thumbUrl + imgid + '.jpg" width="64" height="64"></a>'
				+ '<div class="media-body">'
				+ '<h4 class="media-heading">' + imageUrl + imgid + '.' + ext + '</h4>'
				+ '</div>'
			);
			it.elem.addClass('alert alert-success').attr('role', 'alert');
			request();
		}, function(){
			it.elem.addClass('alert alert-danger').attr('role', 'alert');
			ngf();
		});
	}

	request();
};

function insertImage(data, page){
	// テンプレートとか使いたいなあ
	var i;
	var l = data.List.length;
	var j;
	var tl;
	var it;
	var val;
	var elem;
	var ts = data.ThumbPixelSize;
	var txt;
	var href;
	var root = $('#image-main-area');

	for(i = 0; i < l; ++i){
		it = data.List[i];
		elem = $('<div></div>')
			.attr('id', 'b' + it.Id)
			.addClass('image-box')
			.addClass('page' + page)
			.attr('data-imgid', it.Id)
			.attr('data-toggle', "popover")
			.attr('data-placement', "bottom")
			.attr('data-trigger', "hover")
			.attr('data-animation', "false")
			.attr('data-html', "true")
			.attr('title', imageUrl + it.Id + '.' + it.Ext);

		txt = '<a href="' + imageUrl + it.Id + '.' + it.Ext + '" target="_blank">'
			+ '<img src="' + thumbUrl + it.Id + '.jpg" id="img' + it.Id + '" width="' + ts + '" height="' + ts + '" alt="' + it.Desc + '">'
			+ '</a>'
			+ '<div class="caption hide" id="img-caption-' + it.Id + '">'
			+ (it.Dzi ? '<p><span style="color:red;">[new]</span> <a href="' + viewerUrl + it.Id + '" target="_blank">画像ビューアで開く</a></p>' : '')
			+ '<p>コメント：' + it.Desc + '</p>'
			+ '<p>画像サイズ：' + it.Height + 'x' + it.Width + '</p>'
			+ '<p>データサイズ：' + it.Size + '</p>'
			+ '<p>アップロード時間：' + it.Date + '</p>';

		if(it.Tags && it.Tags.length > 0){
			tl = it.Tags.length;
			txt += '<p>タグ：';
			for(j = 0; j < tl; ++j){
				val = it.Tags[j];
				if(data.Tagmap && data.Tagmap[val] !== undefined){
					href = data.Oldtags;
				} else {
					href = '/?tag=' + encodeURIComponent(val) + ((data.Oldtags !== '') ? '&' + data.Oldtags : '');
				}
				txt += ' <a href="' + href + '">' + val + '</a>';
			}
			txt += '</p>';
		}
		if(it.PassCode){
			txt += '<p>';
			txt += '[<a href="#update" data-toggle="modal" data-target="#update" data-imgid="' + it.Id + '">編集する</a>]';
			txt += '[<a href="#delete" data-toggle="modal" data-target="#delete" data-imgid="' + it.Id + '">削除する</a>]';
			txt += '</p>';
		}
		txt += '</div>';

		elem.html(txt);
		root.append(elem);
	}
	return l >= 50;
};

function getUrlVars(){
	var vars = {};
	var hash;
	var hashes = window.location.search.slice(window.location.search.indexOf('?') + 1).split('&');
	var l = hashes.length;
	var i;
	for(i = 0; i < l; i++){
		hash = hashes[i].split('=');
		if(vars[hash[0]] !== undefined){
			vars[hash[0]].push(hash[1]);
		} else {
			vars[hash[0]] = [hash[1]];
		}
	}
	return vars;
};

function setQueue(queue, files){
	var i;
	var l = files.length;
	var file;

	if(l > 50){
		window.alert("ファイル多すぎ。小分けにアップロードしてね。");
		return;
	}

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

$(function(){
	setEventGlobal();
	setPopover();
	setEventScrollLast();
	setTooltip();
	if(userAgent.FileReader){
		setDropUpload();
		setMultiUpload();
	}
});

})();
