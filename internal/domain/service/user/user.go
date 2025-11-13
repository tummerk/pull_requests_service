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
}

type UserService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{
		repository: repository,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user entity.User) (entity.User, error) {
	return s.repository.Create(ctx, user)
}

func (s *UserService) GetByTeam(ctx context.Context, team string) ([]entity.User, error) {
	return s.repository.GetByTeam(ctx, team)
}

func (s *UserService) SetIsActive(ctx context.Context, userId string, isActive bool) (entity.User, error) {
	return s.repository.SetIsActive(ctx, userId, isActive)
}
