package service

import (
	"context"
	"pull_requests_service/internal/domain/entity"
	service "pull_requests_service/internal/domain/service/user"
	"pull_requests_service/pkg/errcodes"
)

type PullRequestRepository interface {
	CreateWithReviewers(ctx context.Context, pr *entity.PullRequest, reviewerIDs []string) error
	Merge(ctx context.Context, prId string) (entity.PullRequest, error)
	Reassign(ctx context.Context, prId, oldReviewerId string) (entity.PullRequest, string, error)
}

type PullRequestService struct {
	UserRepository         service.UserRepository
	PullRequestsRepository PullRequestRepository
}

func (s *PullRequestService) CreatePullRequest(ctx context.Context, pr entity.PullRequest) (entity.PullRequest, error) {
	candidateIds, err := s.UserRepository.GetActiveTeamCandidatesId(ctx, pr.AuthorId)
	if err != nil {
		logger(ctx).Error("error getting active team candidates: %s", errcodes.NotFound)
		return entity.PullRequest{}, err
	}
	err = s.PullRequestsRepository.CreateWithReviewers(ctx, &pr, candidateIds)
	if err != nil {
		logger(ctx).Error("error creating pull requests: %s", errcodes.NotFound)
		return entity.PullRequest{}, err
	}
	return pr, nil
}

func (s *PullRequestService) Merge(ctx context.Context, prId string) (entity.PullRequest, error) {
	pr, err := s.PullRequestsRepository.Merge(ctx, prId)
	if err != nil {
		logger(ctx).Error("error creating pull requests: %s", errcodes.NotFound)
		return entity.PullRequest{}, err
	}
	return pr, nil
}

func (s *PullRequestService) Reassign(ctx context.Context, prId string, oldId string) (entity.PullRequest, string, error) {
	pr, newId, err := s.PullRequestsRepository.Reassign(ctx, prId, oldId)
	if err != nil {
		logger(ctx).Error("error creating pull requests: %s", errcodes.NotFound)
	}
	return pr, newId, err
}
