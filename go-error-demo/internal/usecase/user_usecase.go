package usecase

import (
	"fmt"

	crerrors "github.com/cockroachdb/errors"
	pkgerrors "github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/benri/go-error-demo/internal/repository"
)

// UserUsecase orchestrates repository access and adds domain specific context.
type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// ExecutePlainErrorCase simply propagates the plain error from repository.
func (u *UserUsecase) ExecutePlainErrorCase() error {
	err := u.repo.GetUserPlainError(42)
	if err != nil {
		// ここでは特にwrapせずそのまま返す
		return err
	}
	return nil
}

// ExecuteFmtWrappedErrorCase wraps the repository error with fmt.Errorf and %w.
func (u *UserUsecase) ExecuteFmtWrappedErrorCase() error {
	err := u.repo.GetUserFmtWrappedError(99)
	if err != nil {
		// ここでfmt.Errorfによるwrap（%w）
		return fmt.Errorf("usecase: failed to fetch user: %w", err)
	}
	return nil
}

// ExecutePkgErrorsWrappedErrorCase adds additional context using pkg/errors.
func (u *UserUsecase) ExecutePkgErrorsWrappedErrorCase() error {
	err := u.repo.GetUserPkgErrorsWrappedError(7)
	if err != nil {
		// ここでpkg/errorsによるwrap
		return pkgerrors.Wrap(err, "usecase: enriching pkg/errors stack")
	}
	return nil
}

// ExecuteCockroachErrorCase enriches cockroachdb/errors with extra annotations.
func (u *UserUsecase) ExecuteCockroachErrorCase() error {
	err := u.repo.GetUserCockroachError(21)
	if err != nil {
		return crerrors.WithSafeDetails(err, "usecase: cockroach propagation", err.Error())
	}
	return nil
}

// ExecuteMultiErrCase aggregates multiple repository errors using go.uber.org/multierr.
func (u *UserUsecase) ExecuteMultiErrCase() error {
	var combined error
	if err := u.repo.GetUserPlainError(1); err != nil {
		combined = multierr.Append(combined, fmt.Errorf("primary lookup: %w", err))
	}
	if err := u.repo.GetUserFmtWrappedError(2); err != nil {
		combined = multierr.Append(combined, fmt.Errorf("cache lookup: %w", err))
	}
	if combined != nil {
		return combined
	}
	return nil
}
