package serr

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// general errors
var (
	ErrNotPermitted = errors.New("not permitted")
)

// ToGRPC converts an error into a gRPC error.
func ToGRPC(err error) error {
	msg := err.Error()

	switch {

	case errors.Is(err, ErrNotPermitted):
		return status.Error(codes.Unauthenticated, msg)

	case errors.Is(err, sql.ErrNoRows):
		return status.Error(codes.NotFound, msg)

	case isInvalidUUIDMessage(msg):
		return status.Error(codes.InvalidArgument, msg)
	}

	var jsonErr *json.SyntaxError
	if errors.As(err, &jsonErr) {
		return status.Error(codes.InvalidArgument, msg)
	}

	return status.Error(codes.Internal, msg)
}

// isInvalidUUIDMessage reports whether the message comes from a failed uuid parse.
func isInvalidUUIDMessage(msg string) bool {
	return msg == "invalid uuid" || strings.HasPrefix(msg, "invalid UUID")
}
