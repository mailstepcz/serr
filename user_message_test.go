package serr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithUserMessage(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		req := require.New(t)
		req.Nil(WithUserMessage(nil, "msg"))
	})

	t.Run("error string is unchanged", func(t *testing.T) {
		req := require.New(t)
		base := errors.New("technical")
		err := WithUserMessage(base, "Do not press that button!")
		req.Equal("technical", err.Error())
	})

	t.Run("UserMessage returns the wrapped message", func(t *testing.T) {
		req := require.New(t)
		err := WithUserMessage(errors.New("technical"), "Do not press that button!")
		um, ok := err.(UserMessager)
		req.True(ok)
		req.Equal("Do not press that button!", um.UserMessage())
	})

	t.Run("errors.Is traverses the wrapper", func(t *testing.T) {
		req := require.New(t)
		base := errors.New("sentinel")
		err := WithUserMessage(base, "Do not press that button!")
		req.True(errors.Is(err, base))
	})

	t.Run("errors.As traverses the wrapper", func(t *testing.T) {
		req := require.New(t)
		base := New("inner", String("k", "v"))
		err := WithUserMessage(base, "Do not press that button!")
		var target *serror
		req.True(errors.As(err, &target))
		req.Equal("inner", target.msg)
	})
}

func TestExtractUserMessage(t *testing.T) {
	t.Run("plain error returns empty", func(t *testing.T) {
		req := require.New(t)
		req.Equal("", ExtractUserMessage(errors.New("plain")))
	})

	t.Run("nil error returns empty", func(t *testing.T) {
		req := require.New(t)
		req.Equal("", ExtractUserMessage(nil))
	})

	t.Run("single user message", func(t *testing.T) {
		req := require.New(t)
		err := WithUserMessage(errors.New("inner"), "Outer message")
		req.Equal("Outer message", ExtractUserMessage(err))
	})

	t.Run("user message survives serr.Wrap", func(t *testing.T) {
		req := require.New(t)
		domain := WithUserMessage(errors.New("inner"), "Domain message")
		wrapped := Wrap("service layer", domain, String("k", "v"))
		req.Equal("Domain message", ExtractUserMessage(wrapped))
	})

	t.Run("multiple user messages join outermost first", func(t *testing.T) {
		req := require.New(t)
		inner := WithUserMessage(errors.New("e"), "Inner detail")
		outer := WithUserMessage(inner, "Outer summary")
		req.Equal("Outer summary: Inner detail", ExtractUserMessage(outer))
	})

	t.Run("user message survives multiple Wrap layers", func(t *testing.T) {
		req := require.New(t)
		domain := WithUserMessage(errors.New("inner"), "Domain message")
		l1 := Wrap("svc l1", domain)
		l2 := Wrap("svc l2", l1)
		req.Equal("Domain message", ExtractUserMessage(l2))
	})

	t.Run("user message inside WrapMulti", func(t *testing.T) {
		req := require.New(t)
		domain := WithUserMessage(errors.New("inner"), "Domain message")
		multi := WrapMulti("aggregated", []error{errors.New("a"), domain})
		req.Equal("Domain message", ExtractUserMessage(multi))
	})

	t.Run("user message inside errors.Join", func(t *testing.T) {
		req := require.New(t)
		domain := WithUserMessage(errors.New("inner"), "Domain message")
		joined := errors.Join(errors.New("a"), domain)
		req.Equal("Domain message", ExtractUserMessage(joined))
	})

	t.Run("user message under multi inside Wrap", func(t *testing.T) {
		req := require.New(t)
		domain := WithUserMessage(errors.New("inner"), "Domain message")
		multi := WrapMulti("aggregated", []error{domain})
		outer := Wrap("svc", multi)
		req.Equal("Domain message", ExtractUserMessage(outer))
	})

	t.Run("multiple user messages across multi branches", func(t *testing.T) {
		req := require.New(t)
		a := WithUserMessage(errors.New("ea"), "Branch A")
		b := WithUserMessage(errors.New("eb"), "Branch B")
		multi := WrapMulti("aggregated", []error{a, b})
		outer := WithUserMessage(multi, "Outer summary")
		req.Equal("Outer summary: Branch A: Branch B", ExtractUserMessage(outer))
	})
}
