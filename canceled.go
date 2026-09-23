package serr

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsCanceled reports whether err is, or wraps, a cancellation caused by the caller going away.
//
// Covers context.Canceled, a gRPC status carrying codes.Canceled (which never unwraps to
// context.Canceled), and the Postgres query_canceled SQLSTATE. context.DeadlineExceeded is
// deliberately not a cancellation - a timeout is a real signal worth reporting.
func IsCanceled(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) {
		return true
	}

	if s, ok := status.FromError(err); ok && s.Code() == codes.Canceled {
		return true
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgerrcode.QueryCanceled {
		return true
	}

	return false
}
