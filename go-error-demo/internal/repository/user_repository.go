package repository

import (
	"errors"
	"fmt"

	crerrors "github.com/cockroachdb/errors"
	pkgerrors "github.com/pkg/errors"
)

var (
	// ErrDBConnection represents a low level connectivity issue.
	ErrDBConnection = errors.New("db connection failed")
)

// NotFoundError is returned when a user does not exist.
type NotFoundError struct {
	ID int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("user %d not found", e.ID)
}

// UserRepository defines the behaviour required by the usecase layer.
type UserRepository interface {
	GetUserPlainError(id int) error
	GetUserFmtWrappedError(id int) error
	GetUserPkgErrorsWrappedError(id int) error
	GetUserCockroachError(id int) error
}

// DemoUserRepository is a small stub that always returns errors.
type DemoUserRepository struct{}

func NewDemoUserRepository() *DemoUserRepository {
	return &DemoUserRepository{}
}

// GetUserPlainError demonstrates returning a plain error without wrapping.
func (r *DemoUserRepository) GetUserPlainError(id int) error {
	// ここで素のerrorを生成
	return ErrDBConnection
}

// GetUserFmtWrappedError returns a typed error for later wrapping.
func (r *DemoUserRepository) GetUserFmtWrappedError(id int) error {
	// このエラーは上位レイヤーでfmt.Errorfによりwrapされます
	return &NotFoundError{ID: id}
}

// GetUserPkgErrorsWrappedError returns an error already wrapped with pkg/errors.
func (r *DemoUserRepository) GetUserPkgErrorsWrappedError(id int) error {
	base := errors.New("timeout from replica set")
	// ここでpkg/errorsによるwrap + stack付与
	return pkgerrors.WithStack(pkgerrors.Wrap(base, "repository: query failed"))
}

// GetUserCockroachError demonstrates github.com/cockroachdb/errors with details and stacks.
func (r *DemoUserRepository) GetUserCockroachError(id int) error {
	base := crerrors.Newf("replica %d returned stale data", id%3)
	annotated := crerrors.Wrap(base, "repository: cockroach annotated error")
	// WithDetailを使って構造化された詳細を付与
	return crerrors.WithDetail(annotated, "rebalance pending: follower lag exceeded")
}

var _ UserRepository = (*DemoUserRepository)(nil)
