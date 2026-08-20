# Gin Template Demo

**Team AIB** のハッカソン用Gin(Go)開発テンプレート。
**`docker compose up` だけで、アプリとDBが揃った開発環境が立ち上がります。**

> リポジトリ名は `case_gin`、画面上の表示名は `Gin Template Demo` です。
> プロダクト名が決まったら、表示名は `internal/view/page.go` の
> `SiteName` の1か所を直してください(全ページのタイトルとヘッダーに反映されます)。

---

## いきなり動かす

```bash
# 1. 金庫を作る(初回だけ)
cp .env.example .env          # Windows: Copy-Item .env.example .env

# 2. 起動する
docker compose up --build
```

→ <http://localhost:8080> を開く

**DBに表を作るコマンドはありません。** 起動するたびに自動で用意されます。

うまくいかないときは **[docs/SETUP.md](docs/SETUP.md)** を見てください。困ったときの対処が全部書いてあります。

---

## 使っている技術

| 分類 | 技術 |
| --- | --- |
| フロント | 素のHTML / CSS / JavaScript(Goの標準テンプレート) |
| バック | Go 1.26 / Gin / GORM |
| DB | MySQL 8.4(確認は DBeaver、ポートは **3308**) |
| 環境 | Docker Compose / air(保存したら自動で作り直す道具) |
| 本番 | Nginx / AWS |

---

## フォルダの地図

「アプリ = 1軒のお店」だと思って読んでください。

```
case_gin/
├── cmd/server/main.go      お店の開店作業(起動の入口)
│
├── internal/               お店の裏側(★他のプロジェクトからは入れない場所)
│   ├── config/             設定を1か所に集める
│   ├── database/           DBに繋ぐ共用の道具
│   ├── models/             データの形(DBの表)を決める
│   ├── middleware/         入口に置く仕掛け(セッション・ログイン確認・成りすまし対策)
│   ├── handlers/           受付。URL → 処理
│   │   ├── page.go           画面(HTML)を返す
│   │   ├── auth.go           ログイン・新規登録
│   │   └── api.go            JavaScript向けにデータだけ返す
│   ├── router/             受付をまとめて組み立てる(土台)
│   └── view/               型紙(base.html)を使う仕組み(土台)
│
├── web/                    お客さんの目に入るもの
│   ├── templates/
│   │   ├── layouts/base.html   全ページ共通の型紙
│   │   ├── pages/              各ページの中身(ここに置くだけで使える)
│   │   └── partials/           型紙が長くなったら部品を切り出す置き場
│   └── static/                 CSS / JS / 画像
│
├── docs/                   チームで見る手順書
├── tools/                  開発中だけ使う小道具スクリプト
│
├── compose.yml             アプリとDBをまとめて動かす段取り表
├── Dockerfile              箱を組み立てるレシピ
├── .air.toml               保存したら自動で作り直す設定
├── go.mod / go.sum         買い物リストとレシート(全員が同じ部品を使うための記録)
│
├── .vscode/                チーム共通のエディタ設定
│
├── .env                    金庫(★GitHubに上げない)
└── .env.example            金庫の中身の見本(こちらは上げる)
```

> **`internal` という名前について**
> Goでは `internal` という名前のフォルダは「このプロジェクトの中からしか使えない」という
> 決まりになっています。お店のバックヤードのようなもので、外から勝手に触られません。
> 特別な設定は不要で、名前を付けるだけでそうなります。

---

## 担当ごとに触る場所

**基本的に他の人と同じファイルを触らないように分けてあります。** これでコンフリクト(変更の取り合い)がほぼ起きません。

| 担当 | 触る場所 |
| --- | --- |
| 見た目 | `web/templates/` `web/static/css/` |
| 画面の動き | `web/static/js/` |
| データの形 | `internal/models/` |
| URLと処理(画面) | `internal/handlers/page.go` |
| URLと処理(データ) | `internal/handlers/api.go` |
| ログイン | `internal/handlers/auth.go` |

`main.go` `router/` `view/` `config/` `database/` `middleware/` `compose.yml` `Dockerfile` は**土台**です。
触る必要が出たら、**先にチームに共有してから**変更してください(全員に影響します)。

---

## ページを1枚増やす手順

1. **`web/templates/pages/` にHTMLを1枚置く**(`index.html` をコピーするのが早い)

   ```html
   {{ define "content" }}
   <h1>マイページ</h1>
   {{ end }}
   ```

   > ヘッダーとフッターは書きません。`base.html` から自動的に付きます。

2. **`internal/handlers/page.go` に2つ足す**

   ```go
   func RegisterPageRoutes(r *gin.Engine) {
       r.GET("/", index)
       r.GET("/mypage", myPage)   // ← 1行足す
   }

   func myPage(c *gin.Context) {  // ← 関数を書く
       c.HTML(http.StatusOK, "mypage.html", view.Page(c, gin.H{
           "Title": "マイページ",
       }))
   }
   ```

3. 保存して数秒待つ → <http://localhost:8080/mypage>

**登録作業はこれだけです。** テンプレートの一覧に書き足す必要はありません(自動で見つけます)。

ログインしている人だけに見せたいときは、1行挟むだけです。

```go
r.GET("/mypage", middleware.RequireLogin(), myPage)
```

---

## base.html(型紙)の仕組み

**今回いちばん新しい考え方なので、ここだけ先に押さえてください。**

ヘッダーとフッターを全ページにコピペすると、直すときに全ファイルを回ることになります。
そこで**共通部分を1枚の「型紙」にまとめ、各ページは真ん中の中身だけを書く**形にしています。

```
base.html(型紙)                    pages/index.html(中身)
┌──────────────────┐
│ ヘッダー          │
├──────────────────┤
│                  │  ←──  {{ define "content" }}
│  ここが穴         │            <h1>ホーム</h1>
│                  │        {{ end }}
├──────────────────┤
│ フッター          │
└──────────────────┘
```

- 型紙側:`{{ block "content" . }}{{ end }}` と書いた場所が**穴**になる
- ページ側:`{{ define "content" }} ～ {{ end }}` に書いた中身が**その穴に入る**
- 名前(`content`)が両者で一致していることが条件

穴は3つ用意してあります。

| 穴の名前 | 用途 |
| --- | --- |
| `content` | ページの本体(必須) |
| `head_extra` | そのページだけで使うCSSを足したいとき |
| `scripts` | そのページだけで使うJSを足したいとき |

**ヘッダーを直したいときは `base.html` を1枚直すだけ**で、全ページに反映されます。

> ★つまずきやすい点が2つあります。詳しくは `web/templates/layouts/base.html` の
> 冒頭のコメントに書いてあるので、最初に一度読んでください。
>
> 1. `{{ .Title }}` のように書いても、**渡していない情報は空になる**(エラーも出ない)
> 2. `{{ range }}` の中では**ドットの意味が変わる**(全体の情報は `{{ $.SiteName }}`)

---

## よく使うコマンド

| やりたいこと | コマンド |
| --- | --- |
| 起動する | `docker compose up` |
| 裏で起動する | `docker compose up -d` |
| 止める | `docker compose down` |
| エラーを見る | `docker compose logs -f web` |
| DBを作り直す(データは消える) | `docker compose exec web go run ./cmd/server -reset-db` |
| 箱の中に入る | `docker compose exec web bash` |
| ライブラリを追加する | `docker compose exec web go get <ライブラリ>` |
| 書き方を自動で整える | `docker compose exec web go fmt ./...` |
| 怪しい書き方を調べる | `docker compose exec web go vet ./...` |

---

## Flask版(case_flask)との対応表

同じ構成で作ってあるので、片方が分かればもう片方も読めます。

| やること | Flask版 | Gin版 |
| --- | --- | --- |
| 起動の入口 | `src/web/app.py` | `cmd/server/main.go` |
| 設定 | `config.py` | `internal/config/` |
| 共用の道具 | `extensions.py` | `internal/database/` |
| データの形 | `models.py` | `internal/models/` |
| 画面を返す | `routes.py` | `internal/handlers/page.go` |
| ログイン | `auth/routes.py` | `internal/handlers/auth.go` |
| 型紙 | `templates/base.html`(`extends`) | `web/templates/layouts/base.html`(`define`) |
| 表を作る | `flask init-db` を実行 | **不要**(起動時に自動) |
| 表を作り直す | `flask drop-db` → `init-db` | `go run ./cmd/server -reset-db` |
| アプリのポート | 5000 | 8080 |
| DBのポート | 3307 | 3308 |

> ポートをずらしてあるので、**Flask版とGin版を同時に起動しても衝突しません。**

---

## ドキュメント

- **[docs/SETUP.md](docs/SETUP.md)** — 環境構築、日々の操作、DBeaverでの接続、困ったときの対処
