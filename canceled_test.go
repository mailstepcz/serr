package serr

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestIsCanceled(t *testing.T) {
	tcs := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil",
			err:  nil,
			want: false,
		},
		{
			name: "unrelated error",
			err:  errors.New("boom"),
			want: false,
		},
		{
			name: "plain context.Canceled",
			err:  context.Canceled,
			want: true,
		},
		{
			name: "wrapped context.Canceled",
			err:  Wrap("loading entity", context.Canceled),
			want: true,
		},
		{
			name: "joined error containing context.Canceled",
			err:  errors.Join(errors.New("boom"), context.Canceled),
			want: true,
		},
		{
			name: "gRPC status carrying codes.Canceled",
			err:  status.Error(codes.Canceled, "context canceled"),
			want: true,
		},
		{
			name: "wrapped gRPC status carrying codes.Canceled",
			err:  Wrap("getting users by ids", status.Error(codes.Canceled, "context canceled")),
			want: true,
		},
		{
			name: "postgres query_canceled SQLSTATE",
			err:  &pgconn.PgError{Code: pgerrcode.QueryCanceled},
			want: true,
		},
		{
			name: "wrapped postgres query_canceled SQLSTATE",
			err:  Wrap("listing timelogs", &pgconn.PgError{Code: pgerrcode.QueryCanceled}),
			want: true,
		},
		{
			name: "postgres error with another SQLSTATE",
			err:  Wrap("listing timelogs", &pgconn.PgError{Code: pgerrcode.DeadlockDetected}),
			want: false,
		},
		{
			name: "context.DeadlineExceeded is not a cancellation",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "gRPC status carrying codes.DeadlineExceeded is not a cancellation",
			err:  status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
			want: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, IsCanceled(tc.err))
		})
	}
}

// TestIsCanceledCoversWhatErrorsIsMisses documents the cancellation forms that a plain
// errors.Is(err, context.Canceled) check does not catch.
func TestIsCanceledCoversWhatErrorsIsMisses(t *testing.T) {
	tcs := []struct {
		name string
		err  error
	}{
		{
			name: "gRPC status carrying codes.Canceled",
			err:  status.Error(codes.Canceled, "context canceled"),
		},
		{
			name: "postgres query_canceled SQLSTATE",
			err:  &pgconn.PgError{Code: pgerrcode.QueryCanceled},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			req := require.New(t)

			req.False(errors.Is(tc.err, context.Canceled))
			req.True(IsCanceled(tc.err))
		})
	}
}
