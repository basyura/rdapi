# main.go の type 分離計画

## 目的

`main.go` に定義されているデータ構造を用途別パッケージへ分離し、
`main` パッケージの責務を CLI の処理フロー中心に整理する。

## 修正対象

- `.gitignore` に `.DS_Store` を追加する。
- `authConfig` を `config` パッケージへ移動する。
  - 新規ファイル: `config/auth_config.go`
  - 型名: `AuthSettings`
- `loadAuthConfig` と `loadAuthSecrets` を `config` パッケージへ移動する。
  - 新規ファイル: `config/auth_loader.go`
  - 補助関数: `loadAuthConfig`
  - 補助関数: `loadAuthSecrets`
  - 補助関数: `parseAuthSection`
- `tokenResponse` を `api` パッケージへ移動する。
  - 新規ファイル: `api/token.go`
  - 型名: `TokenSet`
- `raindropsResponse` を `api` パッケージへ移動する。
  - 新規ファイル: `api/raindrops.go`
  - 型名: `raindropsResponse`
- `raindrop` を `api` パッケージへ移動する。
  - 新規ファイル: `api/raindrops.go`
  - 型名: `Raindrop`
- `authorizationURL` 関数を `api` パッケージへ移動する。
  - 新規ファイル: `api/authorization.go`
  - 関数名: `BuildAuthorizationURL`
- `promptAuthorizationCode` 関数は CLI 入力処理として `cli` パッケージへ移動する。
  - 新規ファイル: `cli/prompt.go`
  - 関数名: `PromptAuthorizationCode`
- `defaultConfigPath` 関数を `config` パッケージへ移動する。
  - 新規ファイル: `config/path.go`
  - 補助関数: `defaultConfigPath`
- `defaultSecretPath` 関数を `config` パッケージへ移動する。
  - ファイル: `config/path.go`
  - 補助関数: `defaultSecretPath`
- 表示幅調整用の文字列関数を `term` パッケージへ移動する。
  - 新規ファイル: `term/display_width.go`
  - 関数名: `TruncateByDisplayWidth`
  - 関数名: `DisplayWidth`
  - 補助関数: `runeDisplayWidth`
- 端末幅取得関数を `term` パッケージへ移動する。
  - 新規ファイル: `term/terminal_width.go`
  - 関数名: `TerminalWidth`
- `main.go` の API/OAuth 関連関数を役割別に移動する。
  - `ExtractAuthorizationCode`
  - `ExchangeCode`
  - `RefreshAccessToken`
  - `FetchAllRaindrops`
  - `FetchRaindropsPage`
- `openBrowser` は CLI 操作として `cli` パッケージへ移動する。
  - 新規ファイル: `cli/open_url.go`
  - 関数名: `OpenURL`
- `SaveAuthTokens` は `config` パッケージの公開関数とし、ファイル保存処理は補助関数 `saveAuthTokensToFile` とする。
- `UpsertAuthValue` は `config` パッケージ内の補助関数 `upsertAuthValue` とする。

## 実装方針

- `main.go` の CLI 引数は廃止し、設定は `config.toml` と `secret.toml` だけで管理する。
- `main.go` は設定ファイルと secret ファイルの path を意識しない。
- path 解決は `config` パッケージ内で行う。
- 読み込みは `config.LoadAuth()` から行う。
- token 保存は `config.SaveAuthTokens()` から行う。
- `redirectURI` は `config.toml` の `auth.redirect_uri` のみを使う。
- 認可コードは引数では受け取らず、必要時に対話入力で受け取る。
- `main.go` から対象の type 定義を削除する。
- `main.go` に `rdapi/api` と `rdapi/config` の import を追加する。
- 既存の関数シグネチャと変数宣言を新しい公開型へ置き換える。
- 設定ファイルと secret ファイルの読み込みは `config` パッケージへ寄せる。
- `raindropsResponse.Items` は `[]Raindrop` として定義する。
- JSON タグと既存のフィールド構造は変更しない。
- テスト内の型参照も新しいパッケージ名に合わせて更新する。
- macOS のメタデータファイル `.DS_Store` を Git 管理対象から除外する。
- OAuth 認可 URL 生成は `api.BuildAuthorizationURL` から呼び出す。
- API 通信に必要な補助関数は `api` パッケージ内へ寄せる。
- token 保存は `config.SaveAuthTokens` から呼び出す。
- 認可コード入力処理は `cli.PromptAuthorizationCode` から呼び出す。
- ブラウザ起動は `cli.OpenURL` から呼び出す。
- デフォルト設定ファイルと secret ファイルの path 解決は `config` パッケージ内で行う。
- 表示幅の省略処理は `term.TruncateByDisplayWidth` から呼び出す。
- 端末幅は `term.TerminalWidth` から取得する。
- テストは対象パッケージごとの `_test.go` へ分割する。
- API 通信処理は token 系、raindrops 系、エラー補助でファイル分割する。
- 内部 decode 用の response struct と token upsert 補助は非公開にする。
- ブックマークの表示整形は `view` パッケージへ移動する。
  - 新規ファイル: `view/raindrops.go`
- `view.FormatRaindrops` は呼び出し元のスライス順を変更しないよう、
  コピーしてからソートする。
- `config.AuthConfig` は設定値と secret 値を統合した結果を表すため、
  `AuthSettings` に変更する。
- `config.AuthSettings` のファイル名は `config/auth_settings.go` に変更する。
- 読み込み関数は戻り値の型に合わせて `config.LoadAuthSettings` に変更する。
- `cli.OpenURL` は汎用 URL 起動関数としてエラーメッセージも汎用化する。
- `view.FormatRaindrops` は端末幅を引数で受け取り、端末幅取得の責務を
  呼び出し側へ移す。
- Raindrop API の page size は `raindropsPerPage` 定数で表す。
- token API の JSON decode 用 struct は private な `tokenResponse` とし、
  公開する `TokenSet` は利用する token 値だけを持つ。
- Raindrop API の JSON decode 用 struct は private な `raindropResponse`
  とし、公開する `Raindrop` は `CreatedAt time.Time` を持つ内部モデルにする。
- `view` は `FormatRaindrops` で表示行を生成し、`PrintLines` は
  出力だけを担当する。
- パッケージごとのテストファイルを追加する。
  - `api/authorization_code_test.go`
  - `config/path_test.go`
  - `config/token_store_test.go`
  - `term/display_width_test.go`
  - `view/raindrops_test.go`

## 確認

- ファイル編集後に `gofmt` を実行する。
- `go test ./...` を実行して既存動作が維持されていることを確認する。
