package server

import (
	"fmt"

	crerrors "github.com/cockroachdb/errors"
	pkgerrors "github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/benri/go-error-demo/internal/errorsdemo"
	"github.com/benri/go-error-demo/internal/usecase"
)

// Server mimics HTTP handlers for demonstration purposes.
type Server struct {
	uc *usecase.UserUsecase
}

func NewServer(uc *usecase.UserUsecase) *Server {
	return &Server{uc: uc}
}

// HandlePlainErrorCase logs a plain error with almost no context.
func (s *Server) HandlePlainErrorCase() {
	err := s.uc.ExecutePlainErrorCase()
	if err != nil {
		errorsdemo.DumpError("plain", err)
	}
}

// HandleFmtWrappedErrorCase wraps the error twice using fmt.Errorf + %w.
func (s *Server) HandleFmtWrappedErrorCase() {
	err := s.uc.ExecuteFmtWrappedErrorCase()
	if err != nil {
		// server層でもfmt.Errorfによるwrap
		err = fmt.Errorf("server: handler failed: %w", err)
		errorsdemo.DumpError("fmt.Errorf wrap", err)
	}
}

// HandlePkgErrorsWrappedErrorCase demonstrates pkg/errors stack traces through layers.
func (s *Server) HandlePkgErrorsWrappedErrorCase() {
	err := s.uc.ExecutePkgErrorsWrappedErrorCase()
	if err != nil {
		// server層でさらにpkg/errorsでwrap
		err = pkgerrors.Wrap(err, "server: handler failed with pkg/errors")
		errorsdemo.DumpError("pkg/errors wrap", err)
	}
}

// HandleCockroachErrorCase showcases cockroachdb/errors annotations.
func (s *Server) HandleCockroachErrorCase() {
	err := s.uc.ExecuteCockroachErrorCase()
	if err != nil {
		err = crerrors.Wrap(err, "server: cockroach handler failed")
		errorsdemo.DumpError("cockroachdb/errors", err)
	}
}

// HandleMultiErrCase aggregates failures using go.uber.org/multierr and shows its breakdown.
func (s *Server) HandleMultiErrCase() {
	err := s.uc.ExecuteMultiErrCase()
	if err != nil {
		if len(multierr.Errors(err)) > 1 {
			// 複数のエラーをまとめた場合のみ文脈を追加
			err = fmt.Errorf("server: multierr handler failed: %w", err)
		}
		errorsdemo.DumpError("multierr aggregate", err)
	}
}
