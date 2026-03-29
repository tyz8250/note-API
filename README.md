# Note API

## 概要

Goの標準ライブラリ（`net/http`, `encoding/json`）で実装したメモCRUD APIです。  
現在は SQLite（`modernc.org/sqlite`）を導入し、`notes` テーブルを作成して起動しています。

現状のデータ保存先は以下の通りです。

- `GET /notes` は SQLite の `notes` テーブルから取得
- `POST /notes` は SQLite の `notes` テーブルへ保存
- `GET /notes/{id}`, `PUT /notes/{id}`, `DELETE /notes/{id}` はメモリ配列を使用

## 実行方法

### サーバ起動

```bash
go run main.go
```

起動時に `notes.db` が作成され、`notes` テーブルが存在しない場合は自動作成されます。

## API 利用例

### メモ一覧取得

```bash
curl http://localhost:8080/notes
```

### メモ取得

```bash
curl http://localhost:8080/notes/1
```

### メモ作成

```bash
curl -X POST http://localhost:8080/notes \
  -H "Content-Type: application/json" \
  -d '{"title": "新しいメモ", "content": "メモの内容"}'
```

### メモ更新

```bash
curl -X PUT http://localhost:8080/notes/1 \
  -H "Content-Type: application/json" \
  -d '{"title": "更新されたメモ", "content": "更新された内容"}'
```

### メモ削除

```bash
curl -X DELETE http://localhost:8080/notes/1
```

## メモ

現時点ではDB移行の途中段階です。今後は `GET /notes/{id}` 以降のCRUDも SQLite に統一予定です。
