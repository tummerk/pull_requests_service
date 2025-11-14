package service

import (
	"context"
	"pull_requests_service/internal/domain/entity"
)

type TeamRepository interface {
	Create(ctx context.Context, team entity.Team) (entity.Team, error)
	Get(ctx context.Context, name string) (entity.Team, error)
}

type TeamService struct {
	Repository  TeamRepository
	userService UserService
}

func (t *TeamService) TeamCreate(ctx context.Context, team entity.Team, users []entity.User) (
	entity.Team, []entity.User, error) {
	team, err := t.Repository.Create(ctx, team)
	if err != nil {
		logger(ctx).Error(err.Error())
		return entity.Team{}, []entity.User{}, err
	}
	for i := range users {
		users[i].Team = team.Name
		users[i], err = t.userService.CreateUser(ctx, users[i])
		if err != nil {
			logger(ctx).Error(err.Error())
			return entity.Team{}, []entity.User{}, err
		}
	}
	return team, users, nil
}

func (t *TeamService) TeamGet(ctx context.Context, name string) (entity.Team, []entity.User, error) {
	team, err := t.Repository.Get(ctx, name)
	if err != nil {
		logger(ctx).Error(err.Error())
		return entity.Team{}, []entity.User{}, err
	}
	users, err := t.userService.GetByTeam(ctx, name)
	if err != nil {
		logger(ctx).Error(err.Error())
		return entity.Team{}, []entity.User{}, err
	}
	return team, users, nil
}
