package serr

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type attributed struct {
	id uuid.UUID
}

func (a *attributed) Attributes() []Attr {
	return []Attr{UUID("id", a.id), Int("num", 1234), Uint("wheels", 3)}
}

func TestAttributed(t *testing.T) {
	req := require.New(t)

	id := uuid.New()
	var a Attributed = &attributed{id: id}

	req.Equal([]Attr{UUID("id", id), Int("num", 1234), Uint("wheels", 3)}, a.Attributes())

	err := New("dummy error", String("attr", "abcd"), a)
	req.Equal("dummy error attr=abcd id="+id.String()+" num=1234 wheels=3", err.Error())
}

func TestErrorAttributes(t *testing.T) {
	req := require.New(t)

	err := New("msg", String("a", "1"), String("b", "2"))
	req.Equal("msg a=1 b=2", err.Error())
}

type custom struct {
	Data string
}

func (c *custom) LogString() string { return "custom: " + c.Data }

func TestAnyAttributes(t *testing.T) {
	req := require.New(t)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	err := New("msg", Any("attr", &custom{Data: "data"}))
	LogError(context.Background(), logger, err)
	req.Contains(buf.String(), `"level":"ERROR","msg":"msg","attr":"custom: data"`)
}

func TestWrappedErrors(t *testing.T) {
	t.Run("message & wrapped error", func(t *testing.T) {
		req := require.New(t)

		err := Wrap("msg", errors.New("malheur"), String("a", "1"), String("b", "2"))
		req.Equal("msg: malheur a=1 b=2", err.Error())
	})

	t.Run("message & wrapped errors", func(t *testing.T) {
		req := require.New(t)

		err := WrapMulti("msg", []error{errors.New("malheur"), errors.New("catastrophe")}, String("a", "1"), String("b", "2"))
		req.Equal("msg: malheur/catastrophe a=1 b=2", err.Error())
	})

	t.Run("no message & error", func(t *testing.T) {
		req := require.New(t)

		err := Wrap("", errors.New("malheur"), String("a", "1"), String("b", "2"))
		req.Equal("malheur a=1 b=2", err.Error())
	})

	t.Run("error unwrapping", func(t *testing.T) {
		req := require.New(t)

		ErrSome := errors.New("some error")

		err := Wrap("", ErrSome, String("a", "1"), String("b", "2"))
		req.True(errors.Is(err, ErrSome))
	})
}

type object1 struct {
	Data string
}

func (obj *object1) LogString() string { return "log string: " + obj.Data }

type object2 struct {
	Data string
}

func TestLogString(t *testing.T) {
	req := require.New(t)

	obj1 := &object1{"OBJ1"}
	logstr, ok := logString(obj1)
	req.True(ok)
	req.Equal(logstr, "log string: OBJ1")

	obj2 := &object2{"OBJ2"}
	logstr, ok = logString(obj2)
	req.True(ok)
	req.Equal(logstr, `{
 "Data": "OBJ2"
}`)
}

var gr interface{}

func BenchmarkAttrSlice(b *testing.B) {
	a, bb, c := "abcd", 1234, 12.34
	var lr interface{}
	for i := 0; i < b.N; i++ {
		sl := []interface{}{a, bb, c}
		lr = sl
	}
	gr = lr
}

func BenchmarkAttrSliceFunc(b *testing.B) {
	a, bb, c := "abcd", 1234, 12.34
	var lr interface{}
	for i := 0; i < b.N; i++ {
		f := func() []interface{} { return []interface{}{a, bb, c} }
		lr = f
	}
	gr = lr
}
