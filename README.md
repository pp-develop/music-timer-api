# 　music-timer-api

## 概要
[music-timer](https://github.com/pp-develop/music-timer) 用の WEB API。
時間を指定して Spotify / SoundCloud のプレイリストを作成します。

## セットアップ
`.env` ファイルを作成してください。

## 使い方

### 1. コンテナの作成と起動

Web開発用:
```bash
$ cd build
$ CADDY_ENV=web docker compose up -d
```

Native(モバイル)開発用:
```bash
$ cd build
$ CADDY_ENV=native docker compose up -d
```

> **注意:** `CADDY_ENV` を切り替える際は `--force-recreate` を付けてください。付けないとCaddyコンテナが再作成されず、前回の設定が使われ続けます。
> ```bash
> $ CADDY_ENV=web docker compose up -d --force-recreate caddy
> ```
>
> 確認方法:
> ```bash
> $ docker exec caddy cat /etc/caddy/Caddyfile | head -3
> # default_sni localhost → Web用
> # default_sni 192.168.0.32 → Native用
> ```

### 2. CockroachDB の初期化
```bash
$ docker exec -it cockroachdb bash
$ cockroach sql --insecure --host=localhost:26257 < /cockroach/init.sql
```

### 3. トラックデータの初期化（初回セットアップ時に必要）
```bash
$ curl -X POST http://localhost:8080/tracks \
  -H "Content-Type: application/json" \
  -d '{"market": "JP"}'
```
