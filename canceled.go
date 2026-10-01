package serr

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// sqlStateError is a database driver error carrying a Postgres SQLSTATE code.
type sqlStateError interface {
	error
	SQLState() string
}

// IsCanceled reports whether err is, or wraps, a cancellation caused by the caller going away.
//
// Covers context.Canceled, a gRPC status carrying codes.Canceled (which never unwraps to
// context.Canceled), and the Postgres query_canceled SQLSTATE from any driver exposing
// SQLState() (pgx *pgconn.PgError, lib/pq *pq.Error). context.DeadlineExceeded is
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

	if sqlErr, ok := errors.AsType[sqlStateError](err); ok && sqlErr.SQLState() == pgerrcode.QueryCanceled {
		return true
	}

	return false
}
