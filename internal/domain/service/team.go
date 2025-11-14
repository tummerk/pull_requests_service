package service

import (
	"context"
	"errors"
	"fmt"
	"pull_requests_service/internal/domain"
	"pull_requests_service/internal/domain/entity"
	"pull_requests_service/pkg/errcodes"
)

type TeamRepository interface {
	Create(ctx context.Context, team entity.Team) (entity.Team, error)
	Get(ctx context.Context, name string) (entity.Team, error)
}

type TeamService struct {
	teamRepo TeamRepository
	userRepo UserRepository
}

func NewTeamService(teamRepo TeamRepository, userRepo UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) TeamCreate(ctx context.Context, team entity.Team, users []entity.User) (entity.Team, []entity.User, error) {
	createdTeam, err := s.teamRepo.Create(ctx, team)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			return entity.Team{}, nil, err
		}
		return entity.Team{}, nil, domain.WrapError(err, errcodes.InternalServerError, "failed to create team")
	}
	createdUsers := make([]entity.User, len(users))
	for i, user := range users {
		user.Team = createdTeam.Name

		createdUser, err := s.userRepo.Create(ctx, user)
		if err != nil {
			logger(ctx).Error("failed to create user in team creation process", "user_id", user.Id, "error", err)
			var appErr *domain.AppError
			if errors.As(err, &appErr) {
				return entity.Team{}, nil, err
			}
			return entity.Team{}, nil, domain.WrapError(err, errcodes.InternalServerError, fmt.Sprintf("failed to create user '%s'", user.Id))
		}
		createdUsers[i] = createdUser
	}

	return createdTeam, createdUsers, nil
}

func (s *TeamService) TeamGet(ctx context.Context, name string) (entity.Team, []entity.User, error) {
	team, err := s.teamRepo.Get(ctx, name)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			return entity.Team{}, nil, err
		}
		return entity.Team{}, nil, domain.WrapError(err, errcodes.InternalServerError, "failed to get team")
	}

	users, err := s.userRepo.GetByTeam(ctx, name)
	if err != nil {
		return entity.Team{}, nil, domain.WrapError(err, errcodes.InternalServerError, "failed to get users for team")
	}

	return team, users, nil
}
