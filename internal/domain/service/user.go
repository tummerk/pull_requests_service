package service

import (
	"context"
	"pull_requests_service/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user entity.User) (entity.User, error)
	GetByTeam(ctx context.Context, team string) ([]entity.User, error)
	SetIsActive(ctx context.Context, userId string, isActive bool) (entity.User, error)
	GetById(ctx context.Context, userId string) (entity.User, error)
	GetActiveTeamCandidatesId(ctx context.Context, authorID string) ([]string, error)
	GetUserAssignmentStats(ctx context.Context) ([]entity.UserAssignmentStat, error)
}

type PrWorkerJob struct {
	userId string
	value  bool
}

type UserService struct {
	repository UserRepository
	eventChan  chan<- PrWorkerJob
}

func NewUserService(repository UserRepository, eventChan chan<- PrWorkerJob) *UserService {
	return &UserService{
		repository: repository,
		eventChan:  eventChan,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userId string, isActive bool) (entity.User, error) {
	user, err := s.repository.SetIsActive(ctx, userId, isActive)
	if err != nil {
		return entity.User{}, err
	}
	select {
	case s.eventChan <- PrWorkerJob{userId, isActive}:
		return user, nil
	default:
		logger(ctx).Warn("Failed to publish user status change event: channel is full", "user_id", user.Id)
	}
	return user, nil
}
