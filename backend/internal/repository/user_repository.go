package repository

import (
	"context"
	"database/sql"
	"fmt"

	"dndproject/backend/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

var userSortColumns = []string{"id", "display_name", "created_at", "updated_at"}

func UserSortColumns() []string { return userSortColumns }

const userColumns = `id, email, display_name, role, created_at, updated_at`

func scanUser(row interface{ Scan(...any) error }) (model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *UserRepository) List(ctx context.Context, q model.Query) ([]model.User, int64, error) {
	where, args := "", []any{}
	if q.Search != "" {
		where = "WHERE display_name LIKE ? OR email LIKE ?"
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM users %s ORDER BY %s %s LIMIT ? OFFSET ?`, userColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, u)
	}
	return items, total, rows.Err()
}

func (r *UserRepository) Get(ctx context.Context, id int64) (model.User, error) {
	return scanUser(r.db.QueryRowContext(ctx, `SELECT `+userColumns+` FROM users WHERE id = ?`, id))
}

// GetCredentials returns the user alongside the stored hash, which is the only
// place the hash leaves the database.
func (r *UserRepository) GetCredentials(ctx context.Context, email string) (model.User, string, error) {
	var u model.User
	var hash string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, display_name, role, created_at, updated_at, password_hash FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.Role, &u.CreatedAt, &u.UpdatedAt, &hash)
	return u, hash, err
}

func (r *UserRepository) Create(ctx context.Context, email, displayName, passwordHash string, role model.Role) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (email, display_name, password_hash, role) VALUES (?, ?, ?, ?)`,
		email, displayName, passwordHash, role)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (r *UserRepository) EmailTaken(ctx context.Context, email string) (bool, error) {
	var found int
	err := r.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE email = ?`, email).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// CreateWithInvite inserts the user and spends the invite in one transaction,
// so a signup that fails on a duplicate email does not burn the code. It is the
// one place the user repository touches another table, and it is deliberate:
// the two writes have to succeed or fail together.
//
// It also covers the bootstrap case. While no users exist the first account is
// created as the DM and no code is required, which is how the first DM gets in.
// Register that account before the hostname is shared.
func (r *UserRepository) CreateWithInvite(ctx context.Context, code, email, displayName, passwordHash string) (model.User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.User{}, err
	}
	defer tx.Rollback()

	var users int64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&users); err != nil {
		return model.User{}, err
	}

	role := model.RoleDM
	if users > 0 {
		// Spending the invite first means the UPDATE is the race guard: a second
		// registration on the same code matches no rows and is rejected.
		if err := tx.QueryRowContext(ctx,
			`SELECT role FROM invites WHERE code = ? AND used_by IS NULL AND expires_at > NOW()`, code,
		).Scan(&role); err != nil {
			return model.User{}, err
		}
	}

	res, err := tx.ExecContext(ctx,
		`INSERT INTO users (email, display_name, password_hash, role) VALUES (?, ?, ?, ?)`,
		email, displayName, passwordHash, role)
	if err != nil {
		return model.User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.User{}, err
	}

	if users > 0 {
		spent, err := tx.ExecContext(ctx,
			`UPDATE invites SET used_by = ?, used_at = NOW() WHERE code = ? AND used_by IS NULL AND expires_at > NOW()`,
			id, code)
		if err != nil {
			return model.User{}, err
		}
		if err := affected(spent); err != nil {
			return model.User{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.User{}, err
	}
	return r.Get(ctx, id)
}
