package repository

import (
	"context"
	"database/sql"
	"fmt"
	"slices"

	"dndproject/backend/internal/model"
)

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

var chatSortColumns = []string{"id", "created_at"}

func ChatSortColumns() []string { return chatSortColumns }

const chatColumns = `id, campaign_id, user_id, author_name, body, created_at`

func scanChatMessage(row interface{ Scan(...any) error }) (model.ChatMessage, error) {
	var m model.ChatMessage
	err := row.Scan(&m.ID, &m.CampaignID, &m.UserID, &m.AuthorName, &m.Body, &m.CreatedAt)
	return m, err
}

func (r *ChatRepository) Create(ctx context.Context, campaignID int64, userID int64, authorName, body string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO chat_messages (campaign_id, user_id, author_name, body) VALUES (?, ?, ?, ?)`,
		campaignID, userID, authorName, body)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ChatRepository) Get(ctx context.Context, id int64) (model.ChatMessage, error) {
	return scanChatMessage(r.db.QueryRowContext(ctx, `SELECT `+chatColumns+` FROM chat_messages WHERE id = ?`, id))
}

// Recent returns the tail of the log in chronological order.
// If retentionDays > 0, it only returns messages within that window.
func (r *ChatRepository) Recent(ctx context.Context, campaignID int64, limit int, retentionDays int) ([]model.ChatMessage, error) {
	where := "WHERE campaign_id = ?"
	args := []any{campaignID}
	if retentionDays > 0 {
		where += " AND created_at >= NOW() - INTERVAL ? DAY"
		args = append(args, retentionDays)
	}

	query := `SELECT ` + chatColumns + ` FROM chat_messages ` + where + ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.ChatMessage{}
	for rows.Next() {
		m, err := scanChatMessage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	slices.Reverse(items)
	return items, nil
}

// List is the paginated history behind the scrollback endpoint.
func (r *ChatRepository) List(ctx context.Context, campaignID int64, q model.Query) ([]model.ChatMessage, int64, error) {
	where, args := "WHERE 1 = 1", []any{}
	if campaignID > 0 {
		where += " AND campaign_id = ?"
		args = append(args, campaignID)
	}
	if q.Search != "" {
		where += " AND body LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM chat_messages "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM chat_messages %s ORDER BY %s %s LIMIT ? OFFSET ?`, chatColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.ChatMessage
	for rows.Next() {
		m, err := scanChatMessage(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, rows.Err()
}
