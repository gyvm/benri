package repository

import (
	"errors"
	"fmt"

	crdb "github.com/cockroachdb/errors"
)

// ErrNotFound は、レコードが見つからない場合に返される公開エラーです。
// errors.Is での判定デモに使用します。
var ErrNotFound = errors.New("record not found")

// UserRepository は、ユーザーデータへのアクセスを抽象化します。
type UserRepository struct{}

// NewUserRepository は、UserRepositoryの新しいインスタンスを生成します。
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// GetUserPlainError は、標準の errors.New を使って素のエラーを返します。
// スタックトレースは含まれません。
func (r *UserRepository) GetUserPlainError(id int) error {
	// ここで素のerrorを生成
	return errors.New("db connection failed")
}

// GetUserFmtWrappedError は、fmt.Errorf の %w でラップされることを想定した元エラーを返します。
func (r *UserRepository) GetUserFmtWrappedError(id int) error {
	// このエラーは上位のレイヤーでラップされる
	return ErrNotFound
}

// PermissionError は、権限が不足していることを示すカスタムエラー型です。
// errors.As のデモに使用します。
type PermissionError struct {
	UserID int
	Action string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("user %d does not have permission to %s", e.UserID, e.Action)
}

// GetUserPkgErrorsWrappedError は、github.com/pkg/errors でラップされることを想定した元エラーを返します。
// ここではカスタムエラー型 *PermissionError を返します。
func (r *UserRepository) GetUserPkgErrorsWrappedError(id int) error {
	// errors.Asでキャッチされるカスタムエラー
	return &PermissionError{UserID: id, Action: "read:user_profile"}
}

// GetUserCockroachDBError は、github.com/cockroachdb/errors を使ってエラーを生成します。
// このライブラリはスタックトレースや追加情報を自動で付与します。
func (r *UserRepository) GetUserCockroachDBError(id int) error {
	// cockroachdb/errorsでエラーを生成
	return crdb.New("cockroachdb: failed to connect to follower node")
}

// GetUserMultiError は、go.uber.org/multierr で集約されることを想定した複数のエラーを返します。
// バッチ処理などで複数のエラーを一度に返したい場合に便利です。
func (r *UserRepository) GetUserMultiError(id int) (error, error) {
	// 複数のエラーを返す
	err1 := errors.New("validation error: name is required")
	err2 := fmt.Errorf("permission denied for user %d", id)
	return err1, err2
}
