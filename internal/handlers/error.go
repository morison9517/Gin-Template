// =============================================================================
// error.go = エラー画面(400 / 401 / 403 / 404 / 500)を返す担当
//
// ▼ 使い方
//
//	handlers.ShowError(c, http.StatusNotFound)
//	return
//
//	★呼んだら必ず return すること。呼んだだけでは処理は止まらない。
//
// ▼ 自動で出る場合が2つある
//
//	① 存在しないURL … 下の RegisterErrorRoutes(NoRoute)
//	② 処理の途中で落ちたとき … router.go の CustomRecovery から呼ばれる
//
// ▼ 文言を直す・種類を増やす
//
//	下の errorPages の表を書き換えるだけ。HTML(error.html)は共通の1枚。
//
// ▼ 消したいとき
//
//	このファイルと web/templates/pages/error.html を消し、router.go の
//	RegisterErrorRoutes と CustomRecovery を元の gin.Recovery() に戻す。
//
// =============================================================================
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"case_gin/internal/view"
)

// errorPage = 画面に出す2行(英語の種類名と日本語の説明)。
type errorPage struct {
	Name    string
	Message string
}

// errorPages = 状態番号ごとの文言。★文言を直すのはここ。
//
// ★400番台は送ってきた側の問題、500番台はこちら側の問題。
//
//	この区切りは検索エンジンも見ている。「ページが無い」ときに200を返すと、
//	無いページが検索結果に載り続ける。
//
// ★401 と 403 の違いは「誰か分からない」と「見せられない」。
//
//	このテンプレートのログイン必須ページは、403ではなくログイン画面へ
//	案内している(middleware/auth.go)。そのほうが親切なため。
var errorPages = map[int]errorPage{
	http.StatusBadRequest: {
		Name:    "Bad Request",
		Message: "リクエストの内容に誤りがあります。",
	},
	http.StatusUnauthorized: {
		Name:    "Unauthorized",
		Message: "このページを表示する権限がありません。",
	},
	http.StatusForbidden: {
		Name:    "Forbidden",
		Message: "このページを表示する権限がありません。",
	},
	http.StatusNotFound: {
		Name:    "Not Found",
		Message: "指定されたページが見つかりませんでした。",
	},
	http.StatusInternalServerError: {
		Name:    "Internal Server Error",
		Message: "サーバー側で問題が発生しました。時間をおいて試してください。",
	},
}

// fallbackErrorPage = 表に無い番号が来たときの文言。
//
// ★表に無いからといって何も返さないと真っ白な画面になる。
var fallbackErrorPage = errorPage{
	Name:    "Error",
	Message: "問題が発生しました。",
}

// RegisterErrorRoutes = このファイルが担当するURLを登録する。
//
// NoRoute = どのURLにも当てはまらなかったときの受け皿。
// 特定のURLではないので他の登録と取り合いにならない。
func RegisterErrorRoutes(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		ShowError(c, http.StatusNotFound)
	})
}

// ShowError = エラー画面を返して、以降の処理を止める。
//
// ★JavaScript向けのURL(/api/...)にはHTMLではなくJSONを返す。
//
//	main.js は {"error": "..."} の形を待っているので、HTMLを返すと
//	JS側が「JSONとして読めない」と別のエラーになり、本当の原因が消える。
func ShowError(c *gin.Context, status int) {
	page, ok := errorPages[status]
	if !ok {
		page = fallbackErrorPage
	}

	// ★判定の並びは middleware/csrf.go と合わせてある。片方だけ直すとズレる。
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/__demo/api/") {
		c.AbortWithStatusJSON(status, gin.H{"error": page.Message})
		return
	}

	c.HTML(status, "error.html", view.Page(c, gin.H{
		"Title": strconv.Itoa(status),

		"ErrorCode":    status,
		"ErrorName":    page.Name,
		"ErrorMessage": page.Message,
	}))

	// ★Abort() が必要。これが無いと呼び出し元の後ろの行が動いてしまう。
	c.Abort()
}
