# 環境変数で設定を切り替えられるようにした

## 背景

これまではサーバーポートやDBパスをコードに直接書いていた。
ただ、この形だと環境変化に弱く、例えばポート番号を変更する場合直接コードをいじる必要があり手間である。
今回はコードから分離した形で環境変数を置くことを目的とする。

## 課題

- `:8080` が直書きされている
- `note.db` が直書きされている
- 設定値がコード内に散らばる

## 対応方針

- `PORT` を環境変数から読む
- `DB_PATH` を環境変数から読む
- 未設定時はデフォルト値を使う
- 設定取得処理をまとめる

## 実装

- `os.Getenv` で取得
- 未設定時は `8080`, `note.db` を使う
- `config.Load()` に集約

## 学んだこと

環境変数は、コードの外から設定を渡す仕組み。
『コードは動き方を書き、環境変数は動く場所ごとの差分を書く』

## 実装に詰まったこと

Config 構造体のフィールド名を `port` にしていたため、他パッケージから参照できずエラーになった。
Go では先頭大文字が公開、先頭小文字が package 内限定になる。
そのため `Port`, `DBPath` に修正した。

## 今後

- `.env` 導入
- Docker 実行時の設定注入
- 本番/開発設定の整理

## 参考記事

- https://pkg.go.dev/os
- https://12factor.net/config
- https://zenn.dev/kurusugawa/articles/golang-env-lib
- https://docs.docker.com/compose/how-tos/environment-variables/set-environment-variables/
