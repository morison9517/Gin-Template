// =============================================================================
// internal/demo/ = 動作確認用のデモページ一式(プロダクトには含まれません)
//
// ▼ このフォルダの決まり
//
//	・開発モードのときだけ取り付ける。本番には出ず、デモ用の表も作られない。
//	・トップページ("/")は handlers/page.go に "/" が無いときだけ引き受ける。
//	  自分たちのトップページを書いた瞬間に自動で出なくなる(消す作業は不要)。
//	・いつでも /__demo で開ける。
//	・★画面(page.html)は1枚完結。base.html も style.css も使わない。
//	  「セットアップが動いているか」を見せる計器なので、共通のファイルを
//	  作り替えても壊れないようにしてある。書き方の見本にはしないこと
//	  (見本は web/templates/pages/login.html)。
//
// ▼ ★Ginだけの注意点
//
//	Ginは同じURLを2回登録すると起動した瞬間に落ちる(FlaskとDjangoは
//	先に登録したほうが勝つだけ)。そこで下の hasRoute() で「"/" が空いているか」
//	を先に確認している。★この確認を消すと、自分たちのトップページを作った日に
//	アプリが起動しなくなる。
//
// ▼ 不要になったら、このフォルダと router.go の「デモ」の数行を消すだけ。
//
// =============================================================================
package demo

import (
	_ "embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"

	"case_gin/internal/config"
	"case_gin/internal/database"
	"case_gin/internal/middleware"
)

// ★//go:embed = このHTMLをアプリ本体の中に取り込む指示。
//
//	web/templates/ に置くと共通の仕組みが自動で base.html と合体させてしまう。
//	1枚完結にしたいので、このフォルダに置いて自分で読み込んでいる。
//
//go:embed page.html
var pageHTML string

// デモ画面のテンプレート。起動時に1回だけ組み立てる。
var pageTmpl *template.Template

// 設定は Register() で受け取って覚えておく(画面に「ログイン機能:有効」を出すため)。
var appConfig *config.Config

// Register = デモを取り付ける。router.go から、開発モードのときだけ呼ばれる。
func Register(r *gin.Engine, cfg *config.Config) error {
	appConfig = cfg

	// デモ用の表を用意する。
	// ★ここでやっているので、本番のDBには todos 表が作られない。
	if err := database.DB.AutoMigrate(&Todo{}); err != nil {
		return err
	}

	// page.html を読み込む。書き間違いがあればここで気づける。
	tmpl, err := template.New("demo").Parse(pageHTML)
	if err != nil {
		return err
	}
	pageTmpl = tmpl

	// /__demo は常に開けるようにしておく(自分たちのトップページを作った後も
	// 「DBに繋がっているか」をここで確認できる)。
	r.GET("/__demo", index)
	registerAPIRoutes(r)

	// ★トップページが空いているときだけ、デモが "/" を引き受ける。
	//   handlers/page.go に "/" を書いた後は、この条件が外れてデモは出なくなる。
	//   (確認せずに登録すると、URLの二重登録でGinが起動時に落ちる)
	if !hasRoute(r, http.MethodGet, "/") {
		r.GET("/", index)
	}

	return nil
}

// hasRoute = そのURLが既に登録されているか調べる。
//
// r.Routes() は「今この受付に登録されている全URLの一覧」。
// ★Ginは同じURLを2回登録すると落ちるので、登録の前に必ずここを通す。
func hasRoute(r *gin.Engine, method, path string) bool {
	for _, info := range r.Routes() {
		if info.Method == method && info.Path == path {
			return true
		}
	}
	return false
}

// index = デモページ。"/" と "/__demo" の両方から使われる。
//
// ★普通のページと違い、c.HTML(...) を使っていない。
//
//	c.HTML は base.html と合体させる共通の仕組みを通るので、
//	1枚完結のこのページには使えない。ここでは自分で組み立てて返している。
//	自分たちのページを作るときは c.HTML を使うこと(見本は handlers/page.go)。
func index(c *gin.Context) {
	dbStatus, dbMessage := checkDB()

	data := map[string]any{
		"Title":       "セットアップ確認",
		"DBStatus":    dbStatus,
		"DBMessage":   dbMessage,
		"AuthEnabled": appConfig.AuthEnabled,
		"CSRFToken":   middleware.CSRFToken(c),

		// ★真偽値だけを渡している(利用者オブジェクトそのものは渡さない)。
		//   画面に名前を出すと .Username を読むことになり、チームが
		//   Userの項目名を変えた日にデモが壊れる。
		//   「ログイン中か」だけならフレームワーク側の概念なので絶対に壊れない。
		"LoggedIn": middleware.CurrentUser(c) != nil,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)

	if err := pageTmpl.Execute(c.Writer, data); err != nil {
		// ここまで来たらHTMLを書き始めた後なので、画面には途中まで出ている。
		// 原因はターミナルのログで確認する。
		_ = c.Error(err)
	}
}

// checkDB = DBに繋がるか実際に試す。
//
// 「返事してください」という最小の確認を送り、返事が来るかで判定している。
//
// エラーで落とさず画面は出す理由:
// DBが起動しきっていないだけでこの画面が真っ白になると原因が分かりにくい。
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
