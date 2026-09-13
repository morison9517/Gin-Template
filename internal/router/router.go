// =============================================================================
// router.go = URLの受付をまとめて組み立てる場所
//
// ★このファイルは土台。普段は触らない。
//
//	URLを増やすときに触るのは handlers/ の中のファイル。
//	ここを触るのは「新しい担当ファイルを増やしたとき」だけ(下に1行足す)。
//
// ▼ 上から順に「通り道」を作っていくイメージ
//
//	ブラウザ
//	  ↓
//	[静的ファイル] CSS・JS・画像はここで直接返す(下の仕掛けを通らない)
//	  ↓
//	[セッション]   ブラウザのメモを読めるようにする
//	  ↓
//	[ログイン確認] 今アクセスしているのが誰か調べる
//	  ↓
//	[成りすまし対策] 送信に整理券が付いているか確認する
//	  ↓
//	[各ページ]     handlers/ の中の関数
//	  ↓
//	[受け皿]       どれにも当てはまらなければエラー画面(handlers/error.go)
//
// =============================================================================
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"case_gin/internal/config"
	"case_gin/internal/demo"
	"case_gin/internal/handlers"
	"case_gin/internal/middleware"
	"case_gin/internal/view"
)

// ファイルの置き場所。プロジェクトの入口(compose.ymlのある場所)から見た位置。
const (
	templateDir = "web/templates"
	staticDir   = "web/static"
)

// New = アプリの受付を組み立てて返す。
func New(cfg *config.Config) (*gin.Engine, error) {
	// 本番では起動ログを静かにし、デバッグ用の警告も出さないモードにする。
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// ▼ ★gin.Recovery() ではなく CustomRecovery() を使う理由
	//
	//   どちらも「1か所でエラーが出てもアプリ全体を止めない」係。
	//   違うのは止めたあとに何を返すかで、gin.Recovery() は中身が空の500
	//   (真っ白な画面)を返す。それでは落ちたことすら伝わらない。
	//   ★原因はこれまで通りログに出る(画面に出すとコードの中身が漏れる)。
	r.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, _ any) {
		handlers.ShowError(c, http.StatusInternalServerError)
	}))

	// アップロードの受け皿サイズ。
	// ★本番のNginx側(docker/nginx/app.inc)とも同じ値にしておく。
	r.MaxMultipartMemory = 16 << 20 // 16MB

	// ★「アクセス元のIPをどこまで信用するか」の設定。
	//
	//   Nginxを前に置くとGinから見た相手はNginxになる。本当のアクセス元は
	//   ヘッダーに書いてあるが、無条件で信じると利用者が自分で詐称できる。
	//   開発中は空(誰も信用しない)。本番は .env の TRUSTED_PROXIES に書く。
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}

	// --- 画面の型紙(base.html)を使えるようにする ---
	renderer, err := view.Setup(cfg, templateDir)
	if err != nil {
		return nil, err
	}
	r.HTMLRender = renderer

	// --- CSS / JS / 画像 ---
	// ★Use より先に書いている。Ginは「Useより後に登録したURL」にだけ
	//   仕掛けを適用するので、画像1枚ごとにログイン確認が走るのを避けられる。
	r.Static("/static", staticDir)

	// --- サイト直下に置かなければならない2つ ---
	// ★検索エンジンは /robots.txt しか見に来ない(/static/ に置いても読まれない)。
	// ★ブラウザは最後の手段として /favicon.ico を取りに来る。
	//   ファイルが無ければ404が返るだけで害はない。
	r.StaticFile("/robots.txt", staticDir+"/robots.txt")
	r.StaticFile("/favicon.ico", staticDir+"/favicon.ico")

	// --- 利用者が上げたファイル ---
	// ★開発モードのときだけGinが配る。本番ではNginxが配る(compose.prod.yml)。
	if !cfg.IsProduction() {
		r.Static("/media", cfg.UploadDir)
	}

	// --- 全ページ共通の仕掛け(順番に意味がある) ---
	// ★2つ目の引数が「HTTPSのときだけメモを送る」の指定。
	//   練習中はHTTPなので .env の SECURE_COOKIES で切り替えられる。
	r.Use(middleware.Session(cfg.SecretKey, cfg.IsProduction() && cfg.SecureCookies))
	r.Use(middleware.LoadUser()) // セッションが読めないと誰か分からないので、後
	r.Use(middleware.CSRF())

	// --- 担当ごとの受付を取り付ける ---
	// ★新しい担当ファイル(例:handlers/room.go)を作ったら、ここに1行足す。
	handlers.RegisterPageRoutes(r)
	handlers.RegisterAPIRoutes(r)

	// AUTH_ENABLED が false なら取り付けない(コードを消さずにOFFにできる)。
	if cfg.AuthEnabled {
		handlers.RegisterAuthRoutes(r)
	}

	// --- デモ(動作確認用のページ) ---
	// ★必ず最後に取り付ける。「自分たちのトップページが既にあるか」を見てから
	//   動くので、先に取り付けると判定できず、URLの二重登録で起動時に落ちる。
	if !cfg.IsProduction() {
		if err := demo.Register(r, cfg); err != nil {
			return nil, err
		}
	}

	// --- エラー画面(404など)の受け皿 ---
	handlers.RegisterErrorRoutes(r)

	return r, nil
}
