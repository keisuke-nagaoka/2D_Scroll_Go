# go_docker

Go言語アプリのローカル開発環境。本番 (https://go.keisuke-nagaoka.com/) と同じく、
1つのnginxの配下にパス単位 (`/go-game/`, `/go-app2/`, ...) で複数アプリを振り分ける構成。

## 構成

```
go_docker/
├── .gitignore
├── compose.yml
├── docker/
│   ├── go/
│   │   └── Dockerfile        # 全アプリ共通のGoビルド用イメージ(air導入済み)
│   └── nginx/
│       ├── default.conf      # パスごとにupstream/locationを追加して振り分け
│       ├── server.crt
│       └── server.key
└── apps/
    └── go-game/
        ├── .air.toml         # air設定(wasmとサーバーをビルド)
        ├── go.mod
        ├── go.sum
        ├── main.go           # HTTPサーバー
        ├── static/           # CSS / ビルド済みwasm
        ├── templates/
        ├── wasm/             # Ebitengineのゲーム本体(GOOS=js GOARCH=wasm)
        └── README.md
```

## 起動

`docker/nginx/server.crt` / `server.key` は自己署名証明書のため、リポジトリには含めていません（`.gitignore`で除外）。初回はまず生成してください。

```bash
openssl req -x509 -nodes -newkey rsa:2048 \
  -keyout docker/nginx/server.key \
  -out docker/nginx/server.crt \
  -days 365 \
  -subj "/CN=localhost"
```

```bash
cd go_docker
docker compose up -d
```

- http://localhost:9100/go-game/
- https://localhost:9543/go-game/ (自己署名証明書)

## 新しいアプリを追加する手順

1. `apps/go-game` をコピーして `apps/<新アプリ名>` を作成し、`go.mod` の `module` 名を変更する
2. `compose.yml` にサービスを追加する (コメントアウトされている `go-app2` を参考にする)
   - `APP_BASE_PATH` は `/<新アプリ名>` にする
   - `nginx` サービスの `depends_on` にも新サービスを追加する
3. `docker/nginx/default.conf` に `upstream` と `location /<新アプリ名>/` ブロックを追加する
   - `location` は HTTP(80) と HTTPS(443) の両方の `server` ブロックに追加する
   - アプリはコンテナ内の `8080` 番ポートで待ち受ける前提
4. `docker compose up -d --build` で反映する

## ホットリロード

各Goコンテナは [air](https://github.com/air-verse/air) を使ってソース変更を検知し、
自動で再ビルド・再起動する。設定は各アプリの `.air.toml` を参照。
(`go-game` では wasm (`static/game.wasm`) とサーバー本体の両方をビルドする)
