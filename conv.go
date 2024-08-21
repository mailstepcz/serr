package serr

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
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

	case uuid.IsInvalidLengthError(err):
		return status.Error(codes.InvalidArgument, msg)

	case msg == "invalid UUID format":
		return status.Error(codes.InvalidArgument, msg)
	}

	var jsonErr *json.SyntaxError
	if errors.As(err, &jsonErr) {
		return status.Error(codes.InvalidArgument, msg)
	}

	return status.Error(codes.Internal, msg)
}
