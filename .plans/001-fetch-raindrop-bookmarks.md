# Raindrop ブックマーク取得の実装計画

## 目的

Raindrop API を使って、登録済みブックマークを取得し、
コマンド実行時に一覧として表示できるようにする。

## 前提

- 認証は Raindrop OAuth2 の認可コードフローを使用する。
- OAuth クライアント情報は
  `~/.config/rdapi/config.toml` の `auth` セクションから読み込む。
  - `client_id`
  - `client_secret`
  - `redirect_uri`
- token 情報は `~/.config/rdapi/secret.toml` の `auth` セクションから読み込む。
  - `access_token`
  - `refresh_token`
- API 呼び出しでは `Authorization: Bearer <token>` ヘッダーを付与する。
- ブックマークは Raindrop API 上の `raindrops` として取得する。

## 修正対象

- `main.go`
- `main_test.go`
- `README.md`
- `.gitignore`
- `AGENTS.md`

## 実装方針

1. `~/.config/rdapi/config.toml` を読み込み、`auth.client_id` と
   `auth.client_secret`、任意の `auth.redirect_uri` を取得する。
2. `access_token` が設定されている場合は、それを使って API を呼び出す。
3. `access_token` がなく `refresh_token` がある場合は、
   refresh token で access token を更新する。
4. 認可コードが未指定かつ token が未保存の場合は、認可 URL を生成して
   ブラウザで開く。
   - ターミナルは code またはリダイレクト URL の入力待ちにする。
   - 入力された code で token を取得して保存する。
5. 認可コードが指定された場合は、
   `POST https://raindrop.io/oauth/access_token` で
   access token を取得する。
   - `code=...` やリダイレクト URL 全体が渡された場合も code を抽出する。
   - access token が含まれないレスポンスは本文を含めてエラー表示する。
6. 取得または更新した token を `secret.toml` に保存する。
7. `GET https://api.raindrop.io/rest/v1/raindrops/0` を呼び出す。
8. レスポンス JSON からブックマークの `title` と `created` を取り出す。
9. 取得したブックマークをブックマーク日時の降順で標準出力へ一覧表示する。
   - 表示フォーマットは `yyyy/MM/dd : ブックマークタイトル` とする。
   - ターミナル幅に合わせて 1 行に収まるよう省略する。
10. HTTP ステータスが成功以外の場合は、ステータス付きでエラーにする。
11. 通信エラーや JSON デコードエラーも呼び出し元へ返す。

## 確認方法

1. `go fmt` を実行する。
2. `go test ./...` を実行する。
3. 必要に応じて以下のように動作確認する。

```sh
go run .
go run . -code ...
go run . -redirect-uri ... -code ...
```

## 注意点

- `client_id`、`client_secret`、access token はソースコードへ埋め込まない。
- token は `config.toml` ではなく `secret.toml` に保存する。
- `config.toml` に token が残っているケースは考慮しない。
- `redirect_uri` は Raindrop のアプリ設定に登録した値と完全一致させる。
- 認可コードは初回 token 取得のみで使い、以後は保存した token を使う。
- 利用者向けの使い方は `README.md` に記載する。
- 作業者向けの注意点は `AGENTS.md` に記載する。
- まずは最小構成で取得と表示を実装し、検索条件やページングは
  必要になった時点で追加する。
