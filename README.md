# Note API

## 概要

Goの標準ライブラリ（net/http, encoding/json）を使って実装した、メモCRUD APIです。データは現在メモリ上で保持しています。
以下のような操作を実装しています。

- メモの作成 (POST /notes)
- メモの一覧表示 (GET /notes)
- メモの取得 (GET /notes/{id})
- メモの更新 (PUT /notes/{id})
- メモの削除 (DELETE /notes/{id})

コード上ではダミーデータを使用しています。

## 実行方法

#### サーバ起動
ターミナルを開き、以下のコマンドを実行してサーバを起動します。
```bash
go run main.go
```

#### メモ一覧取得
GETリクエストでメモの一覧を取得できます。ダミーデータの2つが初期状態では返ります。
```bash
curl http://localhost:8080/notes
```

#### メモ取得
GETリクエストで特定のメモを取得できます。
```bash
curl http://localhost:8080/notes/1
```

#### メモ作成
POSTリクエストでメモを作成できます。
```bash
curl -X POST http://localhost:8080/notes -H "Content-Type: application/json" -d '{"title": "新しいメモ", "content": "メモの内容"}'
```

#### メモ更新
PUTリクエストでメモを更新できます。
```bash
curl -X PUT http://localhost:8080/notes/1 -H "Content-Type: application/json" -d '{"title": "更新されたメモ", "content": "更新された内容"}'
```

#### メモ削除
DELETEリクエストでメモを削除できます。
```bash
curl -X DELETE http://localhost:8080/notes/1
```

## その他

### 設計で詰まったこと
- 頭のなかで整理できておらず、serverに関するコードを書いてないのにGETリクエストに関するコードのみ書いて混乱した。
  --> サーバー起動とハンドラ登録の流れが最初は曖昧だったが、HandleFunc と ListenAndServe の役割を分けて理解できた。

- メモの作成、更新、削除の際に、idを指定して操作する必要があることに気づかず、混乱した。そして、その実装方法がわからなかった。

- POST,PUTの理解してから実装すべきだった。インプットが足りていなかった。

- Decode,Encodeの理解が浅く、どっちがどっちかわからなくなった。
  -->POST/PUT では JSON を Decode して受け取り、レスポンスでは Encode して返す流れを学んだ。

### 今後実装したい機能
- バリデーション追加
- データベースの接続
- データベースの操作
- Handler,Service,Repositoryの分離
- エラーハンドリング
