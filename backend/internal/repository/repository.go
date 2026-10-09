package repository

import "database/sql"

// affected turns an UPDATE/DELETE that matched nothing into sql.ErrNoRows,
// so services can treat "missing row" the same way they do for SELECT.
func affected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
