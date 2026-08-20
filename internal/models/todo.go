// =============================================================================
// ここから下はサンプル。
// プロダクトが決まったら Todo を消して、自分たちの表
// (Post、Room、Message など)に書き換える。
//
// ★新しいモデルのファイルを作ったら、internal/database/database.go の
//
//	Migrate() のリストに1行足すこと。忘れると表が作られない。
//
// =============================================================================
package models

import "time"

// Todo = やることリスト1件分(サンプル)。
type Todo struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Title  string `gorm:"size:200;not null" json:"title"`
	IsDone bool   `gorm:"not null;default:false" json:"is_done"`

	// ★「この番号は users 表の id のこと」という紐付け(外部キー)。
	//
	//	*uint にしているので「持ち主なし」も許される。
	//	ログイン機能をOFFにしても動くようにするため。
	UserID *uint `gorm:"index" json:"user_id,omitempty"`

	// 番号から実物を引くショートカット。
	// database.DB.Preload("User").Find(&todos) と書くと、todo.User に作者が入る。
	// 「constraint:OnDelete:CASCADE」= ユーザーが消えたら、その人のTodoも一緒に消す。
	User *User `gorm:"constraint:OnDelete:CASCADE" json:"-"`

	CreatedAt time.Time `json:"created_at"`
}

func (Todo) TableName() string {
	return "todos"
}
