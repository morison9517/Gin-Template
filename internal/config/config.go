// =============================================================================
// config.go = 設定を1か所に集める場所
//
//	.env(金庫) ──読む──> config.go ──渡す──> アプリの各所
//
//	コード内に直接パスワードやDB住所を書くと、変更時に全ファイルを探し回るうえ、
//	GitHubに秘密を上げてしまう。
//
// ▼ package config = このフォルダは config という1つのまとまり、という表札
//
//	他のファイルから import "case_gin/internal/config" と書いて呼び出す。
//	Goでは「1フォルダ = 1パッケージ」で、フォルダ名とパッケージ名を揃えるのが決まり。
//
// ▼ ★Goの大事なルール:大文字で始まる名前だけが外から使える
//
//	Config  → 他のパッケージから使える(公開)
//	loadEnv → このフォルダの中だけ(非公開)
//	「なぜか他のファイルから見えない」はほぼこれが原因。
//
// =============================================================================
package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config = アプリ全体の設定をまとめた入れ物。
//
// ▼ struct(構造体)= 項目に名前を付けた入れ物
//
//	「設定」という1つの塊にしておくと、関数に渡すときも引数1個で済む。
type Config struct {
	Env         string // development / production
	Port        string // アプリが待ち受ける番号
	SecretKey   string // ログイン状態をブラウザに預けるときの割り印
	AuthEnabled bool   // ログイン機能を使うか

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
}

// Load = .env を読んで Config を組み立てて返す。起動時に1回だけ呼ぶ。
//
// ▼ *Config の * について
//
//	Config だと「設定の中身をまるごとコピーして渡す」、
//	*Config だと「設定の置き場所(住所)を渡す」という意味になる。
//	住所を渡せば全員が同じ1つの設定を見るので、コピーのズレが起きない。
func Load() *Config {
	// .env が無くてもエラーにしない。
	// 本番(AWS)では .env ファイルではなく、サーバー側の環境変数を直接使うため。
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env が見つかりませんでした(環境変数を直接使います)")
	}

	cfg := &Config{
		Env:         env("APP_ENV", "development"),
		Port:        env("PORT", "8080"),
		SecretKey:   env("SECRET_KEY", "dev-secret-key-change-me"),
		AuthEnabled: envBool("AUTH_ENABLED", true),

		DBHost:     env("DB_HOST", "db"),
		DBPort:     env("DB_PORT", "3306"),
		DBName:     env("DB_NAME", "hack_app"),
		DBUser:     env("DB_USER", "hack_user"),
		DBPassword: env("DB_PASSWORD", "hack_password"),
	}

	// ★本番で割り印が初期値のままだと、ログイン状態を偽造される。
	//   起動時に気づけるよう、警告を出す。
	if cfg.IsProduction() && cfg.SecretKey == "dev-secret-key-change-me" {
		log.Println("[config] ⚠ 本番なのに SECRET_KEY が初期値のままです")
	}

	return cfg
}

// IsProduction = 本番モードかどうか。
//
// ▼ func (c *Config) ... = Config に付いている関数(メソッド)
//
//	cfg.IsProduction() と書けるようになる。
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// DSN = DBへの接続文字列を組み立てる。読み方:
//
//	ユーザー:パスワード@tcp(住所:ポート)/DB名?オプション
//
// parseTime=true … DBの日時を Go の時刻として受け取る(無いと文字列のまま届く)
// charset=utf8mb4 … 絵文字も扱える指定(utf8 だと絵文字でエラーになる)
// loc=Asia%2FTokyo … 日本時間で扱う(無いと9時間ずれる)
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Asia%%2FTokyo",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// env = 環境変数を読む。無ければ2つ目の値を使う。
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// envBool = .env の "true"/"false" という文字を、Goの true/false に変換する。
//
// .env に書けるのは文字だけなので、"false" をそのまま使うと
// 「中身のある文字 = true」と判定されてしまう。その事故を防ぐ。
func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
