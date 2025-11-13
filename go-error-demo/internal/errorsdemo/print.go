package errorsdemo

import (
	"errors"
	"fmt"

	crerrors "github.com/cockroachdb/errors"
	"go.uber.org/multierr"

	"github.com/benri/go-error-demo/internal/repository"
)

// DumpError prints verbose information about the provided error.
func DumpError(title string, err error) {
	fmt.Printf("=== %s ===\n", title)
	fmt.Printf("error: %v\n", err)

	// %+v でpkg/errors利用時にはスタックトレースが表示される。
	fmt.Println("error with %+v (pkg/errorsならstack付き):")
	fmt.Printf("%+v\n", err)

	fmt.Println("-- errors.Is / errors.As checks --")
	fmt.Printf("errors.Is(err, repository.ErrDBConnection): %v\n", errors.Is(err, repository.ErrDBConnection))

	var notFound *repository.NotFoundError
	if errors.As(err, &notFound) {
		fmt.Printf("errors.As -> *repository.NotFoundError (ID=%d)\n", notFound.ID)
	} else {
		fmt.Println("errors.As -> *repository.NotFoundError: false")
	}

	// デモとしてcontext deadline exceededと比較
	fmt.Printf("errors.Is(err, context deadline): %v\n", errors.Is(err, contextDeadlineError()))

	if details := crerrors.GetSafeDetails(err); len(details) > 0 {
		fmt.Println("cockroachdb/errors safe details:")
		for i, d := range details {
			fmt.Printf("  [%d] %s\n", i, d)
		}
	}

	if unwrapped := multierr.Errors(err); len(unwrapped) > 0 {
		fmt.Println("multierr breakdown:")
		for i, inner := range unwrapped {
			fmt.Printf("  [%d] %v\n", i, inner)
		}
	}
	fmt.Println()
}

// contextDeadlineError mimics a sentinel error from another package.
func contextDeadlineError() error {
	return contextDeadlineSentinel
}

var contextDeadlineSentinel = errors.New("context deadline exceeded (fake)")
