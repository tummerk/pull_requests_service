package service

import (
	"context"
	"pull_requests_service/internal/domain/entity"
)

type StatisticsService struct {
	userRepo UserRepository
}

func NewStatisticsService(userRepo UserRepository) *StatisticsService {
	return &StatisticsService{userRepo: userRepo}
}

func (s *StatisticsService) GetUserAssignmentStats(ctx context.Context) ([]entity.UserAssignmentStat, error) {
	return s.userRepo.GetUserAssignmentStats(ctx)
}
