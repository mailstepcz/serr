# serr
Structured errors in Go

## Detecting a cancelled caller

`IsCanceled(err) bool` reports whether an error is, or wraps, a cancellation caused
by the caller going away. A cancelled caller does not arrive as one single error
type, so `errors.Is(err, context.Canceled)` alone misses most real cases:

| Boundary | Error returned | caught by `errors.Is(err, context.Canceled)` |
| --- | --- | --- |
| outbound gRPC call | `*status.Error` with `codes.Canceled` | **no** |
| Postgres, server cancels the statement | `*pgconn.PgError`, SQLSTATE `57014` | **no** |
| Postgres, driver aborts first | `context.Canceled` | yes |
| HTTP client | `*url.Error` | yes |
| AWS SDK / smithy | `*smithy.OperationError` | yes |

The gRPC status is the trap: grpc-go gives `*status.Error` its own `Is` that only
matches another status, so a `codes.Canceled` status never unwraps to
`context.Canceled`.

```go
// in a Sentry or logging middleware: a cancelled caller is not a failure
if serr.IsCanceled(err) {
    logger.WarnContext(ctx, "call cancelled by caller", slog.String("error", err.Error()))
    return
}
```

`context.DeadlineExceeded` is deliberately **not** treated as a cancellation - a
timeout is a real signal worth reporting.

The predicate answers only "was this a cancellation". It does not know about codes
deliberately attached further up; a package that layers its own status on an error
(such as `grpcerr` or `httperr`) resolves that first and consults `IsCanceled` only
when nothing else applies.
