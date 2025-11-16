module github.com/benri/go-error-demo

go 1.22

require (
	github.com/cockroachdb/errors v1.11.1
	github.com/pkg/errors v0.9.1
	go.uber.org/multierr v1.11.0
)

replace github.com/cockroachdb/errors => ./third_party/cockroacherrors

replace go.uber.org/multierr => ./third_party/multierr
