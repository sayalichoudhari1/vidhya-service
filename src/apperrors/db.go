package apperrors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// TranslateDBError inspects err for well-known Postgres constraint-violation
// codes and maps them to a client-facing AppError (409 for duplicates, 400
// for missing/invalid references). This is a safety net applied uniformly
// to every entity: even where a service already runs an explicit "does this
// already exist?" pre-check, a concurrent request can still race past it and
// hit the database constraint first - this ensures that case still produces
// a clean, consistent 4xx response instead of a raw SQL error leaking out as
// a 500.
//
// If err does not wrap a *pgconn.PgError, it is returned unchanged.
func TranslateDBError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case "23505": // unique_violation
		return WithMessage(ErrConflict, "a record with the same unique details already exists")
	case "23503": // foreign_key_violation
		return WithMessage(ErrInvalidPayload, "one or more referenced records do not exist")
	case "23502": // not_null_violation
		return WithMessage(ErrInvalidPayload, "a required field is missing")
	case "23514": // check_violation
		return WithMessage(ErrInvalidPayload, "the request violates a data constraint")
	default:
		return err
	}
}
