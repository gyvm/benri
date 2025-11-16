package usecase

import (
	"fmt"

	"github.com/cockroachdb/errors"
	pkgErrors "github.com/pkg/errors"
	"go.uber.org/multierr"
)

// UserRepositorer は、UserRepositoryが満たすべきインターフェースを定義します。
type UserRepositorer interface {
	GetUserPlainError(id int) error
	GetUserFmtWrappedError(id int) error
	GetUserPkgErrorsWrappedError(id int) error
	GetUserCockroachDBError(id int) error
	GetUserMultiError(id int) (error, error)
}

// UserUsecase は、ユーザー関連のビジネスロジックを担当します。
type UserUsecase struct {
	userRepo UserRepositorer
}

// NewUserUsecase は、UserUsecaseの新しいインスタンスを生成します。
func NewUserUsecase(repo UserRepositorer) *UserUsecase {
	return &UserUsecase{userRepo: repo}
}

// ExecutePlainErrorCase は、Repositoryから返された素のエラーをそのまま返します。
func (u *UserUsecase) ExecutePlainErrorCase() error {
	return u.userRepo.GetUserPlainError(1)
}

// ExecuteFmtWrappedErrorCase は、fmt.Errorfと%wを使ってエラーをラップします。
func (u *UserUsecase) ExecuteFmtWrappedErrorCase() error {
	err := u.userRepo.GetUserFmtWrappedError(2)
	if err != nil {
		// ここでfmt.Errorfによるwrap（%w）
		return fmt.Errorf("usecase layer: failed to get user: %w", err)
	}
	return nil
}

// ExecutePkgErrorsWrappedErrorCase は、pkg/errorsを使ってエラーをラップします。
func (u *UserUsecase) ExecutePkgErrorsWrappedErrorCase() error {
	err := u.userRepo.GetUserPkgErrorsWrappedError(3)
	if err != nil {
		// ここでpkg/errorsによるwrap
		return pkgErrors.Wrap(err, "usecase layer: permission check failed")
	}
	return nil
}

// ExecuteCockroachDBErrorCase は、cockroachdb/errorsを使ってエラーをラップします。
func (u *UserUsecase) ExecuteCockroachDBErrorCase() error {
	err := u.userRepo.GetUserCockroachDBError(4)
	if err != nil {
		// cockroachdb/errorsでさらにラップ
		return errors.Wrap(err, "usecase layer: could not reach user service")
	}
	return nil
}

// ExecuteMultiErrorCase は、複数のエラーをmultierrで集約します。
func (u *UserUsecase) ExecuteMultiErrorCase() error {
	err1, err2 := u.userRepo.GetUserMultiError(5)

	// ここでmultierrを使い、複数のエラーを一つにまとめる
	var multiErr error
	multiErr = multierr.Append(multiErr, err1)
	multiErr = multierr.Append(multiErr, err2)

	// multierrのデモを明確にするため、ここではラップせずに直接エラーを返す
	return multiErr
}
