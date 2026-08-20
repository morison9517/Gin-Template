// =============================================================================
// page.go = 画面(HTML)を返す受付
//
//	ブラウザ「/ をください」 → ここ → 「index.html をどうぞ」
//
// ▼ ★ファイルを分けている理由(チーム開発でいちばん効く工夫)
//
//	URLの登録を1つのファイルにまとめて書くと、6人が同じ行を編集することになり、
//	git で必ず衝突(コンフリクト)する。
//	そこで「担当ごとにファイルを分けて、自分のファイルの中で自分のURLを登録する」
//	形にしてある。
//
//	    page.go → 画面を返すURL      ← このファイル
//	    auth.go → ログイン関係のURL
//	    api.go  → JavaScript向けのURL
//
//	router.go は、それぞれの登録係を呼ぶだけ。普段は触らない。
//
// =============================================================================
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"case_gin/internal/database"
	"case_gin/internal/view"
)

// RegisterPageRoutes = このファイルが担当するURLを登録する。
//
// ★画面を1枚増やすときの手順
//  1. web/templates/pages/ にHTMLを1枚置く
//  2. ここに r.GET("/URL", 関数名) を1行足す
//  3. その関数を下に書く
func RegisterPageRoutes(r *gin.Engine) {
	r.GET("/", index)
	r.GET("/health", health)
}

// index = トップページ。
func index(c *gin.Context) {
	dbStatus, dbMessage := checkDB()

	// c.HTML(ステータス, 使うファイル名, 画面に渡す情報)
	//
	//	"index.html" は web/templates/pages/index.html のこと。
	//	view.Page(...) が、ここに書いた情報 + 共通の情報 をまとめてくれる。
	c.HTML(http.StatusOK, "index.html", view.Page(c, gin.H{
		"Title":     "ホーム",
		"DBStatus":  dbStatus,
		"DBMessage": dbMessage,
	}))
}

// checkDB = DBに繋がるか実際に試す。
//
// 「返事してください」という最小の確認を送り、返事が来るかで判定している。
//
// エラーで落とさず画面は出す理由:
// DBが起動しきっていないだけでトップページが真っ白になると原因が分かりにくい。
// 画面は出しつつ「DBだけ未接続」と伝えたほうが切り分けが速い。
func checkDB() (string, string) {
	sqlDB, err := database.DB.DB()
	if err != nil {
		return "ng", truncate(err.Error())
	}

	if err := sqlDB.Ping(); err != nil {
		return "ng", truncate(err.Error())
	}

	return "ok", ""
}

// truncate = エラー文が数百文字になることがあるので、先頭だけにする。
func truncate(s string) string {
	const limit = 120
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "..."
}

// health = 動作確認用。
//
// gin.H{...} を返すと、Ginが自動でJSONに変換して返す。
//
// AWSやNginxが「アプリが生きているか」を定期的に確認しに来る先。
// 開発中も「画面が出ない…アプリ自体は動いてる?」の切り分けに使える。
func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
