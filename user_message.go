package serr

import (
	"errors"
	"strings"
)

// UserMessager wraps an error with a user-friendly message.
type UserMessager interface {
	UserMessage() string
}

type userMessage struct {
	err     error
	userMsg string
}

func (e *userMessage) Error() string       { return e.err.Error() }
func (e *userMessage) Unwrap() error       { return e.err }
func (e *userMessage) UserMessage() string { return e.userMsg }

// WithUserMessage wraps err with a static user-friendly message.
func WithUserMessage(err error, userMsg string) error {
	if err == nil {
		return nil
	}
	return &userMessage{err: err, userMsg: userMsg}
}

// ExtractUserMessage walks the error chain and joins user messages.
func ExtractUserMessage(err error) string {
	var msgs []string
	for e := err; e != nil; e = errors.Unwrap(e) {
		if um, ok := e.(UserMessager); ok && len(um.UserMessage()) > 0 {
			msgs = append(msgs, um.UserMessage())
		}
	}
	if len(msgs) == 0 {
		return ""
	}
	return strings.Join(msgs, ": ")
}
