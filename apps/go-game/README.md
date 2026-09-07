# go-game

`go_docker` 環境で動く個別のGoアプリ。`APP_BASE_PATH` (compose.ymlで設定) を
テンプレート内のリンク生成に使うことで、nginxのパスルーティング(`/go-game/`)配下でも
静的ファイルのパスがずれないようにしている。

## ゲーム内容

[Ebitengine](https://ebitengine.org/) で作った、スーパーマリオブラザーズ風の
横スクロールアクションのオリジナル実装(著作権のある画像・音声素材は使用せず、
すべて単色の図形で描画)。

- アイテムブロック: 下から叩くとコインが出現、触れると取得してスコア加算
- ブロック: 茶色は下から叩くと破壊、灰色は破壊不可
- 落とし穴: 落下すると即ゲームオーバー
- 敵: 触れると即ゲームオーバー(左右にパトロールし、壁で反転)
- 土管: 土管の上で下キー/下ボタンを押すと別のステージへ移動

### 操作方法

PC・スマートフォンどちらでもプレイできる(レスポンシブ)。

| 操作 | PC | スマホ・タッチ |
| --- | --- | --- |
| 左右移動 | ←→ または A/D | 画面左下のボタン |
| ジャンプ | Space / ↑ / W | 画面右下の JUMP ボタン |
| 土管に入る | ↓ または S | 画面右下の v ボタン |

ゲームオーバー時はジャンプ操作でステージ1から再スタートする。

### 実装構成

- `main.go`: 通常のGo HTTPサーバー。`templates/index.html` と `static/` (wasm本体を含む) を配信する
- `wasm/`: Ebitengineで書かれたゲーム本体。`GOOS=js GOARCH=wasm` でビルドし `static/game.wasm` に出力する
- `static/wasm_exec.js`: GoランタイムのJSグルーコード(Goの`$(go env GOROOT)`からコピー)

`static/game.wasm` と `static/wasm_exec.js` はビルド成果物のためGit管理外
(`.gitignore`参照)。`docker compose up` すると `.air.toml` の設定により
コンテナ起動時に自動でビルドされる。

## ローカル開発

```bash
cd go_docker
docker compose up -d
```

`docker/go/Dockerfile` のコンテナが `air` でファイル変更を検知し、自動でビルド・再起動する。

アクセス: http://localhost:9100/go-game/
