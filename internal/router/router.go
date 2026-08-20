// =============================================================================
// router.go = URLの受付をまとめて組み立てる場所
//
// ★このファイルは土台です。普段は触りません。
//
//	URLを増やしたいときに触るのは handlers/ の中のファイルです。
//	ここを触るのは「新しい担当ファイルを増やしたとき」だけ(下に2行足す)。
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
//
// =============================================================================
package router

import (
	"github.com/gin-gonic/gin"

	"case_gin/internal/config"
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

	// gin.Default() = 「アクセスの記録」と「異常時に落ちない仕掛け」が
	// 最初から付いた受付。
	// 異常時に落ちない仕掛けがあるので、1か所でエラーが出てもアプリ全体は止まらない。
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// アップロードの受け皿サイズ。
	// ★本当の上限は本番のNginx側でも設定する(でないと巨大ファイルで詰まる)。
	r.MaxMultipartMemory = 16 << 20 // 16MB

	// ★「アクセス元のIPをどこまで信用するか」の設定。
	//   nil = 誰も信用しない(自分に直接来たアドレスだけを見る)。
	//   本番でNginxを挟むときは、Nginxのアドレスをここに指定する。
	//   指定しないと起動時に警告が出続ける。
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	// --- 画面の型紙(base.html)を使えるようにする ---
	renderer, err := view.Setup(cfg, templateDir)
	if err != nil {
		return nil, err
	}
	r.HTMLRender = renderer

	// --- CSS / JS / 画像 ---
	// ★ここを Use より先に書いている理由
	//   Ginは「Useより後に登録したURL」にだけ仕掛けを適用する。
	//   先に書くことで、画像1枚読むたびにログイン確認が走るのを避けられる。
	r.Static("/static", staticDir)

	// --- 全ページ共通の仕掛け(順番に意味がある) ---
	r.Use(middleware.Session(cfg.SecretKey, cfg.IsProduction()))
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

	return r, nil
}
