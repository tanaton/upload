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
package webutil

// アップロード完了
const UploadSuccessMessage = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>アップロード完了</title>
</head>
<body>
<h1>正常にファイルをアップロードできました。</h1>
<p><a href="/">トップページへ戻る。</a></p>
</body>
</html>`

var UploadSuccessNet = MustNewNetFile(UploadSuccessMessage)

// 400
const BadRequestMessage = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>400 変なリクエスト</title>
</head>
<body>
<h1>リクエストが変です！</h1>
<p>リクエストに不正があります。正規の情報のみでリクエストを構築して下さい。</p>
</body>
</html>`

var BadRequestNet = MustNewNetFile(BadRequestMessage)

// 403
const ForbiddenMessage = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>403 アクセス禁止</title>
</head>
<body>
<h1>アクセスが禁止されています！</h1>
<p>公開するつもりのない場所へアクセスしたか、あなたのネットワークからのアクセスが禁止されています。</p>
</body>
</html>`

var ForbiddenNet = MustNewNetFile(ForbiddenMessage)

// 404
const NotFoundMessage = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>404 見つかりません</title>
</head>
<body>
<h1>ファイルが見つかりません！</h1>
<p>ファイルが見つかりませんでした。URLを確認してください。</p>
</body>
</html>`

var NotFoundNet = MustNewNetFile(NotFoundMessage)

// 405
const MethodNotAllowedMessage = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>405 許可していないメソッド</title>
</head>
<body>
<h1>許可していないメソッドです！</h1>
<p>このページでは許可されていないメソッドが使用されました。</p>
</body>
</html>`

var MethodNotAllowedNet = MustNewNetFile(MethodNotAllowedMessage)

// 413
const RequestEntityTooLargeMessage = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>413 無駄にでかいデータ</title>
</head>
<body>
<h1>データが大きすぎます！</h1>
<p>送付されたデータが大きすぎるため、処理を中断しました。</p>
</body>
</html>`

var RequestEntityTooLargeNet = MustNewNetFile(RequestEntityTooLargeMessage)

// 414
const RequestURITooLongMessage = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>414 無駄に長いURI</title>
</head>
<body>
<h1>無駄に長いURIです！</h1>
<p>URIが長すぎて解析するのが面倒くさいので削ってください。</p>
</body>
</html>`

var RequestURITooLongNet = MustNewNetFile(RequestURITooLongMessage)

// 415
const UnsupportedMediaTypeMessage = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>415 対応していないファイルタイプ</title>
</head>
<body>
<h1>対応していないファイルタイプです！</h1>
<p>JPEG/PNG/GIF/TIFF/WEBPに対応しています。TIFF/WEBPはブラウザによってはアップロードできないかも。</p>
</body>
</html>`

var UnsupportedMediaTypeNet = MustNewNetFile(UnsupportedMediaTypeMessage)

// 500
const InternalServerErrorMessage = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>500 やんごとなきエラー</title>
</head>
<body>
<h1>やんごとなきエラーです！</h1>
<p>リソースの操作に失敗した場合、とりあえずコレになります。</p>
</body>
</html>`

var InternalServerErrorNet = MustNewNetFile(InternalServerErrorMessage)

// 501
const NotImplementedMessage = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>501 対応していないメソッド</title>
</head>
<body>
<h1>対応していないメソッドです！</h1>
<p>色々と対応するのは面倒なのでGETとHEADと一部POSTリクエストのみに対応しています。</p>
</body>
</html>`

var NotImplementedNet = MustNewNetFile(NotImplementedMessage)

// 503
const ServiceUnavailableMessage = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="utf-8">
<title>503 タイムアウトしました</title>
</head>
<body>
<h1>タイムアウトしました！</h1>
<p>プログラムのバグ、もしくはアクセス過多、もしくはサーバがぶっ壊れています。</p>
</body>
</html>`

var ServiceUnavailableNet = MustNewNetFile(ServiceUnavailableMessage)
