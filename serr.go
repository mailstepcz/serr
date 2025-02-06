// Package serr provides types and functions for structured errors.
// A structured error can contain named attributes which can in turn be passed on
// to a structured logger. The current version supports slog.
package serr

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/fealsamh/go-utils/nocopy"
	"github.com/google/uuid"
)

type serror struct {
	msg   string
	attrs []Attributed
}

func (se *serror) Error() string {
	var sb strings.Builder
	sb.WriteString(se.msg)
	for _, attr := range se.attrs {
		for _, attr := range attr.Attributes() {
			sb.WriteByte(' ')
			sb.WriteString(attr.key)
			sb.WriteByte('=')
			if logstr, ok := logString(attr.value); ok {
				sb.WriteString(logstr)
			} else {
				fmt.Fprintf(&sb, "%v", attr.value)
			}
		}
	}
	return sb.String()
}

type wrapped struct {
	msg   string
	err   error
	attrs []Attributed
}

func (se *wrapped) message() string {
	if se.msg == "" {
		return se.err.Error()
	}
	return se.msg + ": " + se.err.Error()
}

func (se *wrapped) Error() string {
	var sb strings.Builder
	sb.WriteString(se.message())
	for _, attr := range se.attrs {
		for _, attr := range attr.Attributes() {
			sb.WriteByte(' ')
			sb.WriteString(attr.key)
			sb.WriteByte('=')
			if logstr, ok := logString(attr.value); ok {
				sb.WriteString(logstr)
			} else {
				fmt.Fprintf(&sb, "%v", attr.value)
			}
		}
	}
	return sb.String()
}

func (se *wrapped) Unwrap() error {
	return se.err
}

type wrappedMulti struct {
	msg   string
	errs  []error
	attrs []Attributed
}

func (se *wrappedMulti) message() string {
	sfx := make([]string, 0, len(se.errs))
	for _, err := range se.errs {
		sfx = append(sfx, err.Error())
	}
	if se.msg == "" {
		return strings.Join(sfx, "/")
	}
	return se.msg + ": " + strings.Join(sfx, "/")
}

func (se *wrappedMulti) Error() string {
	var sb strings.Builder
	sb.WriteString(se.message())
	for _, attr := range se.attrs {
		for _, attr := range attr.Attributes() {
			sb.WriteByte(' ')
			sb.WriteString(attr.key)
			sb.WriteByte('=')
			if logstr, ok := logString(attr.value); ok {
				sb.WriteString(logstr)
			} else {
				fmt.Fprintf(&sb, "%v", attr.value)
			}
		}
	}
	return sb.String()
}

func (se *wrappedMulti) Unwrap() []error {
	return se.errs
}

// Attributed provides custom attributes for structured errors.
type Attributed interface {
	Attributes() []Attr
}

// Attr is a named attribute associated with an error.
type Attr struct {
	key   string
	value interface{}
}

// Attributes returns the attribute as a slice in order to conform to [Attributed].
func (a Attr) Attributes() []Attr {
	return []Attr{a}
}

// String is a string-valued attribute.
func String(key, value string) Attr { return Attr{key: key, value: value} }

// Int is an integer-valued attribute.
func Int(key string, value int) Attr { return Attr{key: key, value: value} }

// UUID is a uuid-valued attribute.
func UUID(key string, value uuid.UUID) Attr { return Attr{key: key, value: value} }

// Time is a time-valued attribute.
func Time(key string, value time.Time) Attr { return Attr{key: key, value: value} }

// Error is an error-valued attribute.
func Error(key string, value error) Attr { return Attr{key: key, value: value} }

// Any is an untyped attribute.
func Any(key string, value interface{}) Attr { return Attr{key: key, value: value} }

// New returns a new structured error.
func New(msg string, attrs ...Attributed) error {
	return &serror{msg: msg, attrs: attrs}
}

// Wrap returns a new structured error which wraps the provided error.
func Wrap(msg string, err error, attrs ...Attributed) error {
	return &wrapped{msg: msg, err: err, attrs: attrs}
}

// WrapMulti returns a new structured error which wraps the provided errors.
func WrapMulti(msg string, errs []error, attrs ...Attributed) error {
	return &wrappedMulti{msg: msg, errs: errs, attrs: attrs}
}

// LogDebug logs a structured error at the debug level.
func LogDebug(ctx context.Context, logger *slog.Logger, err error) {
	Log(ctx, logger, slog.LevelDebug, err)
}

// LogInfo logs a structured error at the info level.
func LogInfo(ctx context.Context, logger *slog.Logger, err error) {
	Log(ctx, logger, slog.LevelInfo, err)
}

// LogWarn logs a structured error at the warn level.
func LogWarn(ctx context.Context, logger *slog.Logger, err error) {
	Log(ctx, logger, slog.LevelWarn, err)
}

// LogError logs a structured error at the error level.
func LogError(ctx context.Context, logger *slog.Logger, err error) {
	Log(ctx, logger, slog.LevelError, err)
}

// Log logs a structured error at the provided level.
func Log(ctx context.Context, logger *slog.Logger, level slog.Level, err error) {
	switch err := err.(type) {
	case *serror:
		logger.Log(ctx, level, err.msg, attrsToSlog(err.attrs)...)
	case *wrapped:
		logger.Log(ctx, level, err.message(), attrsToSlog(err.attrs)...)
	case *wrappedMulti:
		logger.Log(ctx, level, err.message(), attrsToSlog(err.attrs)...)
	default:
		logger.Log(ctx, level, err.Error())
	}
}

func attrsToSlog(errAttrs []Attributed) []interface{} {
	attrs := make([]interface{}, 0, len(errAttrs))
	for _, attr := range errAttrs {
		for _, attr := range attr.Attributes() {
			switch val := attr.value.(type) {
			case string:
				attrs = append(attrs, slog.String(attr.key, val))
			case int:
				attrs = append(attrs, slog.Int(attr.key, val))
			case uuid.UUID:
				attrs = append(attrs, slog.String(attr.key, val.String()))
			case time.Time:
				attrs = append(attrs, slog.Time(attr.key, val))
			case error:
				attrs = append(attrs, slog.String(attr.key, val.Error()))
			default:
				if logstr, ok := logString(val); ok {
					attrs = append(attrs, slog.String(attr.key, logstr))
				} else {
					attrs = append(attrs, slog.Any(attr.key, val))
				}
			}
		}
	}
	return attrs
}

func logString(val interface{}) (string, bool) {
	switch val := val.(type) {
	case string:
		return val, true
	case int:
		return strconv.Itoa(val), true
	case uuid.UUID:
		return val.String(), true
	case time.Time:
		return val.String(), true
	case error:
		return val.Error(), true
	}

	if logval, ok := val.(Loggable); ok {
		return logval.LogString(), true
	} else if b, err := json.MarshalIndent(val, "", " "); err != nil {
		return "", false
	} else {
		return nocopy.String(b), true
	}
}
