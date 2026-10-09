package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

// RecentChatLimit is how much backlog a client gets on joining. Older messages
// come from the paginated history endpoint.
const RecentChatLimit = 100

const maxMessageLength = 2000

type ChatService struct {
	repo *repository.ChatRepository
}

func NewChatService(repo *repository.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

// Post records a message. The author name is copied from the account rather
// than taken from the client, and kept on the row so history survives the
// account being deleted.
func (s *ChatService) Post(ctx context.Context, user model.User, campaignID int64, body string) (model.ChatMessage, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return model.ChatMessage{}, invalid("body", "is required")
	}
	if utf8.RuneCountInString(body) > maxMessageLength {
		return model.ChatMessage{}, invalid("body", "must be at most 2000 characters")
	}

	id, err := s.repo.Create(ctx, campaignID, user.ID, user.DisplayName, body)
	if err != nil {
		return model.ChatMessage{}, err
	}
	return s.repo.Get(ctx, id)
}

func (s *ChatService) Recent(ctx context.Context, campaignID int64, retentionDays int) ([]model.ChatMessage, error) {
	return s.repo.Recent(ctx, campaignID, RecentChatLimit, retentionDays)
}

func (s *ChatService) List(ctx context.Context, campaignID int64, q model.Query) (model.Page[model.ChatMessage], error) {
	items, total, err := s.repo.List(ctx, campaignID, q)
	if err != nil {
		return model.Page[model.ChatMessage]{}, err
	}
	return model.NewPage(items, total, q), nil
}
