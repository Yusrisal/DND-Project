package service

import (
	"context"
	"encoding/json"
	"strings"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

// RecentRollLimit is how much roll history a client gets on joining, merged
// with the chat backlog by timestamp.
const RecentRollLimit = 100

const maxRollLabelLength = 120

type RollService struct {
	repo *repository.RollRepository
}

func NewRollService(repo *repository.RollRepository) *RollService {
	return &RollService{repo: repo}
}

// Roll evaluates the expression on the server and records the result. The
// client sends what to roll, never what was rolled, so a result cannot be
// invented and the breakdown stays checkable afterwards.
func (s *RollService) Roll(ctx context.Context, user model.User, campaignID int64, req model.RollRequest) (model.DiceRoll, error) {
	result, err := EvaluateRoll(req.Expression)
	if err != nil {
		return model.DiceRoll{}, err
	}

	label := strings.TrimSpace(req.Label)
	if len(label) > maxRollLabelLength {
		label = label[:maxRollLabelLength]
	}

	detail, err := json.Marshal(result)
	if err != nil {
		return model.DiceRoll{}, err
	}

	userID := user.ID
	id, err := s.repo.Create(ctx, model.DiceRoll{
		CampaignID:  campaignID,
		UserID:      &userID,
		CharacterID: req.CharacterID,
		AuthorName:  user.DisplayName,
		Label:       label,
		Expression:  result.Expression,
		Detail:      detail,
		Total:       result.Total,
	})
	if err != nil {
		return model.DiceRoll{}, err
	}
	return s.repo.Get(ctx, id)
}

func (s *RollService) Recent(ctx context.Context, campaignID int64, retentionDays int) ([]model.DiceRoll, error) {
	return s.repo.Recent(ctx, campaignID, RecentRollLimit, retentionDays)
}

// Record stores a roll whose result was produced elsewhere, such as an attack
// resolution. It fills in the author and campaign so callers cannot attribute a
// roll to somebody else.
func (s *RollService) Record(ctx context.Context, user model.User, campaignID int64, roll model.DiceRoll) (model.DiceRoll, error) {
	userID := user.ID
	roll.CampaignID = campaignID
	roll.UserID = &userID
	roll.AuthorName = user.DisplayName

	id, err := s.repo.Create(ctx, roll)
	if err != nil {
		return model.DiceRoll{}, err
	}
	return s.repo.Get(ctx, id)
}
