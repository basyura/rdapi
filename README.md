# rdapi - Raindrop.io's API

Raindrop の API を使って、登録済みブックマークを取得する
コマンドラインツールです。

## 機能

- Raindrop OAuth2 で認証します。
- 初回実行時に認証 URL をブラウザで開きます。
- ターミナルで認可コード、またはリダイレクト後の URL を入力できます。
- 取得した `access_token` と `refresh_token` は `secret.toml` に保存します。
- 次回以降は保存済み token を使ってブックマークを取得します。
- ブックマークはブックマーク日時の降順で表示します。
- 出力はターミナル幅に合わせて 1 行に収まるよう省略します。

## 設定

設定ファイルは `~/.config/rdapi/config.toml` に作成します。

```toml
[auth]
client_id = "Raindrop の Client ID"
client_secret = "Raindrop の Client Secret"
redirect_uri = "Raindrop のアプリ設定に登録した Redirect URL"
```

`redirect_uri` は Raindrop のアプリ設定に登録した値と完全一致させてください。

token は `config.toml` と同じディレクトリの `secret.toml` に保存されます。
このファイルはプログラムが作成します。

```toml
[auth]
access_token = "..."
refresh_token = "..."
```

## 初回実行

```sh
go run main.go
```

`secret.toml` に token がない場合、認証 URL をブラウザで開きます。
ブラウザで認証後、ターミナルに表示された入力待ちへ code を入力します。

```text
Enter authorization code or redirected URL:
```

リダイレクト後の URL 全体を貼り付けても、`code` を抽出して処理します。
認証に成功すると `secret.toml` に token を保存し、そのままブックマークを
取得します。

## 通常実行

初回認証後は、以下だけでブックマークを取得できます。

```sh
go run main.go
```

`access_token` が保存済みの場合はそれを使います。
`access_token` がなく `refresh_token` がある場合は、access token を更新して
`secret.toml` に保存します。

## 出力形式

出力形式は以下です。

```text
yyyy/MM/dd : ブックマークタイトル
```

例:

```text
2026/05/15 : Raindrop API Documentation
```

タイトルが長い場合は、ターミナル幅に合わせて末尾を `…` で省略します。

## 注意

- `client_secret`、`access_token`、`refresh_token` は公開しないでください。
- `secret.toml` は Git 管理に含めないでください。
- 認可コードは一度使うと再利用できません。
