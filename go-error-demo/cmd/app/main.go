package main

import (
	"fmt"

	"github.com/benri/go-error-demo/internal/repository"
	"github.com/benri/go-error-demo/internal/server"
	"github.com/benri/go-error-demo/internal/usecase"
)

func main() {
	repo := repository.NewDemoUserRepository()
	uc := usecase.NewUserUsecase(repo)
	srv := server.NewServer(uc)

	srv.HandlePlainErrorCase()
	fmt.Println()

	srv.HandleFmtWrappedErrorCase()
	fmt.Println()

	srv.HandlePkgErrorsWrappedErrorCase()
	fmt.Println()

	srv.HandleCockroachErrorCase()
	fmt.Println()

	srv.HandleMultiErrCase()
}

/*
解説まとめ:
- 素のerror: errors.Newで生成されたエラーのみなのでスタックトレースは保持されず、%+vで表示してもメッセージのみ。
- fmt.Errorf + %w: ラップしてもerrors.Is/Asでオリジナルのエラーを判定できる。スタックトレースは保持されないが、文脈を積み重ねられる。
- pkg/errors.Wrap / WithStack: ラップ時にスタックトレースを採取するため、%+vで各レイヤーの呼び出し元が確認できる。errors.Is/Asも機能する。
- cockroachdb/errors: stackに加えて安全な詳細情報（SafeDetails）を付与でき、 %+v / DumpError内で構造化情報を取得できる。
- go.uber.org/multierr: 複数のエラーをまとめたまま1つのerrorとして返却でき、multierr.Errorsで個々の要素を確認できる。
*/
