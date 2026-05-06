package serr

import (
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

// ExtractUserMessage walks the error tree and joins all user messages,
// outermost first. Both single- and multi-error wrappers are traversed.
func ExtractUserMessage(err error) string {
	var msgs []string
	walkErrors(err, func(e error) {
		if um, ok := e.(UserMessager); ok && len(um.UserMessage()) > 0 {
			msgs = append(msgs, um.UserMessage())
		}
	})
	if len(msgs) == 0 {
		return ""
	}
	return strings.Join(msgs, ": ")
}

// walkErrors visits err and every error reachable through Wrapper or
// MultiWrapper, in a pre-order traversal. errors.Unwrap does not descend
// into MultiWrapper, which is why this helper is needed.
func walkErrors(err error, visit func(error)) {
	if err == nil {
		return
	}
	visit(err)
	switch e := err.(type) {
	case Wrapper:
		walkErrors(e.Unwrap(), visit)
	case MultiWrapper:
		for _, inner := range e.Unwrap() {
			walkErrors(inner, visit)
		}
	}
}

var (
	_ Wrapper      = new(userMessage)
	_ UserMessager = new(userMessage)
)
