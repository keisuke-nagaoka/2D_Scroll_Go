# go-game

`go_docker` 環境で動く個別のGoアプリ。`APP_BASE_PATH` (compose.ymlで設定) を
テンプレート内のリンク生成に使うことで、nginxのパスルーティング(`/go-game/`)配下でも
静的ファイルのパスがずれないようにしている。

## ローカル開発

```bash
cd go_docker
docker compose up -d
```

`docker/go/Dockerfile` のコンテナが `air` でファイル変更を検知し、自動でビルド・再起動する。

アクセス: http://localhost:9100/go-game/
