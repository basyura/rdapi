# main.go の type 分離計画

## 目的

`main.go` に定義されているデータ構造を用途別パッケージへ分離し、
`main` パッケージの責務を CLI の処理フロー中心に整理する。

## 修正対象

- `.gitignore` に `.DS_Store` を追加する。
- `authConfig` を `config` パッケージへ移動する。
  - 新規ファイル: `config/auth_config.go`
  - 型名: `AuthConfig`
- `loadAuthConfig` と `loadAuthSecrets` を `config` パッケージへ移動する。
  - 新規ファイル: `config/auth_loader.go`
  - 関数名: `LoadAuthConfig`
  - 関数名: `LoadAuthSecrets`
  - 補助関数: `parseAuthSection`
- `tokenResponse` を `api` パッケージへ移動する。
  - 新規ファイル: `api/token_response.go`
  - 型名: `TokenResponse`
- `raindropsResponse` を `api` パッケージへ移動する。
  - 新規ファイル: `api/raindrops_response.go`
  - 型名: `RaindropsResponse`
- `raindrop` を `api` パッケージへ移動する。
  - 新規ファイル: `api/raindrop.go`
  - 型名: `Raindrop`
- `authorizationURL` 関数を `api` パッケージへ移動する。
  - 新規ファイル: `api/authorization.go`
  - 関数名: `CreateAuthorizationURL`
- `promptAuthorizationCode` 関数は CLI 入力処理として `main` パッケージに残す。
- `defaultConfigPath` 関数を `config` パッケージへ移動する。
  - 新規ファイル: `config/path.go`
  - 関数名: `GetDefaultConfigPath`
- `defaultSecretPath` 関数を `config` パッケージへ移動する。
  - ファイル: `config/path.go`
  - 関数名: `GetDefaultSecretPath`
- 表示幅調整用の文字列関数を `term` パッケージへ移動する。
  - 新規ファイル: `term/display_width.go`
  - 関数名: `TruncateByDisplayWidth`
  - 関数名: `GetDisplayWidth`
  - 補助関数: `getRuneDisplayWidth`
- 端末幅取得関数を `term` パッケージへ移動する。
  - 新規ファイル: `term/terminal_width.go`
  - 関数名: `GetTerminalWidth`
- `main.go` の API/OAuth 関連関数を役割別に移動する。
  - `ExtractAuthorizationCode`
  - `ExchangeCode`
  - `RefreshAccessToken`
  - `FetchAllRaindrops`
  - `FetchRaindropsPage`
- `openBrowser` は端末操作として `term` パッケージへ移動する。
  - 新規ファイル: `term/browser.go`
  - 関数名: `OpenBrowser`
- `SaveAuthTokens` と `UpsertAuthValue` は `config` パッケージへ移動する。

## 実装方針

- `main.go` の CLI 引数は廃止し、設定は `config.toml` と `secret.toml` だけで管理する。
- `main.go` は設定ファイルと secret ファイルの path を意識しない。
- path 解決は `config` パッケージ内で行う。
- 読み込みは `config.LoadAuth()` から行う。
- token 保存は `config.SaveDefaultAuthTokens()` から行う。
- `redirectURI` は `config.toml` の `auth.redirect_uri` のみを使う。
- 認可コードは引数では受け取らず、必要時に対話入力で受け取る。
- `main.go` から対象の type 定義を削除する。
- `main.go` に `rdapi/api` と `rdapi/config` の import を追加する。
- 既存の関数シグネチャと変数宣言を新しい公開型へ置き換える。
- 設定ファイルと secret ファイルの読み込みは `config` パッケージへ寄せる。
- `RaindropsResponse.Items` は `[]Raindrop` として定義する。
- JSON タグと既存のフィールド構造は変更しない。
- テスト内の型参照も新しいパッケージ名に合わせて更新する。
- macOS のメタデータファイル `.DS_Store` を Git 管理対象から除外する。
- OAuth 認可 URL 生成は `api.CreateAuthorizationURL` から呼び出す。
- API 通信に必要な補助関数は `api` パッケージ内へ寄せる。
- token 保存は `config.SaveDefaultAuthTokens` から呼び出す。
- 認可コード入力処理は CLI 操作として `main` パッケージ内に置く。
- ブラウザ起動は `term.OpenBrowser` から呼び出す。
- デフォルト設定ファイルと secret ファイルの path 解決は `config` パッケージ内で行う。
- 表示幅の省略処理は `term.TruncateByDisplayWidth` から呼び出す。
- 端末幅は `term.GetTerminalWidth` から取得する。

## 確認

- ファイル編集後に `gofmt` を実行する。
- `go test ./...` を実行して既存動作が維持されていることを確認する。
