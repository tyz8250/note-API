# Error Handling

## 目的

note API におけるエラー分類と HTTP ステータスの対応を固定する。

## 基本方針

このAPIでは、エラーは JSON 形式で返す。

## エラーレスポンスの例

```json
{
  "error": "not found"
}
```

```json
{
  "error": "internal server error"
}
```

```json
{
  "error": "invalid id"
}
```

```json
{
  "error": "title is required"
}
```

```json
{
  "error": "content is required"
}
```

```json
{
  "error": "invalid request body"
}
```

### HTTP ステータスの考え方

- 400 Bad Request:クライアントからの入力が不正な場合
- 404 Not Found:指定したリソースが存在しない場合
- 500 Internal Server Error:サーバー内部エラーが発生した場合

以下の表は、代表的なエラーケースと HTTP ステータス、および返却するエラーメッセージの対応を示す。

| ケース | ステータス | エラーメッセージ |
|--------|------------|----------------|
| 対象メモが存在しない | 404 | not found |
| DB操作失敗などのサーバー内部エラー | 500 | internal server error |
| 不正なID | 400 | invalid id |
| 必須項目不足（タイトル） | 400 | title is required |
| 必須項目不足（内容） | 400 | content is required |
| 不正なリクエストボディ | 400 | invalid request body |
