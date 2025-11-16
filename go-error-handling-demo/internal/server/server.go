package server

import (
	"fmt"
	"go-error-demo/internal/errorsdemo"
)

// UserUsecaser は、UserUsecaseが満たすべきインターフェースを定義します。
type UserUsecaser interface {
	ExecutePlainErrorCase() error
	ExecuteFmtWrappedErrorCase() error
	ExecutePkgErrorsWrappedErrorCase() error
	ExecuteCockroachDBErrorCase() error
	ExecuteMultiErrorCase() error
}

// Server は、ハンドラロジックをまとめるための構造体です。
type Server struct {
	userUsecase UserUsecaser
}

// NewServer は、Serverの新しいインスタンスを生成します。
func NewServer(uc UserUsecaser) *Server {
	return &Server{userUsecase: uc}
}

// HandlePlainErrorCase は、素のerrorを処理するハンドラです。
func (s *Server) HandlePlainErrorCase() {
	err := s.userUsecase.ExecutePlainErrorCase()
	if err != nil {
		// Server層でコンテキストを追加するが、%vを使いエラーチェーンを切断する。
		// これにより「ラップされていない」エラーの挙動を正確にデモする。
		err = fmt.Errorf("server layer: request failed: %v", err)
		errorsdemo.DumpError("1. Plain Error (errors.New)", err)
	}
}

// HandleFmtWrappedErrorCase は、fmt.Errorfでラップされたエラーを処理するハンドラです。
func (s *Server) HandleFmtWrappedErrorCase() {
	err := s.userUsecase.ExecuteFmtWrappedErrorCase()
	if err != nil {
		// Server層でもfmt.Errorfでラップ
		err = fmt.Errorf("server layer: user retrieval failed: %w", err)
		errorsdemo.DumpError("2. fmt.Errorf Wrapped Error", err)
	}
}

// HandlePkgErrorsWrappedErrorCase は、pkg/errorsでラップされたエラーを処理するハンドラです。
func (s *Server) HandlePkgErrorsWrappedErrorCase() {
	err := s.userUsecase.ExecutePkgErrorsWrappedErrorCase()
	if err != nil {
		// pkg/errors.Wrapはnilを返さないので、ここでは追加ラップしない
		errorsdemo.DumpError("3. pkg/errors Wrapped Error", err)
	}
}

// HandleCockroachDBErrorCase は、cockroachdb/errorsでラップされたエラーを処理するハンドラです。
func (s *Server) HandleCockroachDBErrorCase() {
	err := s.userUsecase.ExecuteCockroachDBErrorCase()
	if err != nil {
		errorsdemo.DumpError("4. cockroachdb/errors Wrapped Error", err)
	}
}

// HandleMultiErrorCase は、multierrで集約されたエラーを処理するハンドラです。
func (s *Server) HandleMultiErrorCase() {
	err := s.userUsecase.ExecuteMultiErrorCase()
	if err != nil {
		errorsdemo.DumpError("5. go.uber.org/multierr Aggregated Error", err)
	}
}
