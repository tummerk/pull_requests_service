package service

import (
	"context"
	"errors"
	"fmt"
	"pull_requests_service/internal/domain"
	"pull_requests_service/internal/domain/entity"
	"pull_requests_service/pkg/errcodes"
)

type PullRequestRepository interface {
	CreateWithReviewers(ctx context.Context, pr *entity.PullRequest, reviewerIDs []string) error
	Merge(ctx context.Context, prId string) (entity.PullRequest, error)
	Reassign(ctx context.Context, prId, oldReviewerId string) (entity.PullRequest, string, error)
	GetUserReviews(ctx context.Context, userId string) ([]entity.PullRequest, error)
	AssignToNeedyPRs(ctx context.Context, userID string) error
	ReassignFromAllPRs(ctx context.Context, userID string) error
}

type PullRequestService struct {
	userRepo  UserRepository
	prRepo    PullRequestRepository
	eventChan <-chan PrWorkerJob
}

func NewPullRequestService(userRepo UserRepository, prRepo PullRequestRepository, eventChan <-chan PrWorkerJob) *PullRequestService {
	return &PullRequestService{
		userRepo:  userRepo,
		prRepo:    prRepo,
		eventChan: eventChan,
	}
}

func (s *PullRequestService) CreatePullRequest(ctx context.Context, pr entity.PullRequest) (entity.PullRequest, error) {
	candidateIds, err := s.userRepo.GetActiveTeamCandidatesId(ctx, pr.AuthorId)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			return entity.PullRequest{}, err
		}
		return entity.PullRequest{}, domain.WrapError(err, errcodes.InternalServerError, "failed to get team candidates")
	}
	pr.Status = entity.StatusOpen
	err = s.prRepo.CreateWithReviewers(ctx, &pr, candidateIds)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			return entity.PullRequest{}, err
		}
		return entity.PullRequest{}, domain.WrapError(err, errcodes.InternalServerError, "failed to create pull request with reviewers")
	}
	return pr, nil
}

func (s *PullRequestService) Merge(ctx context.Context, prId string) (entity.PullRequest, error) {
	mergedPR, err := s.prRepo.Merge(ctx, prId)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			return entity.PullRequest{}, err
		}
		return entity.PullRequest{}, domain.WrapError(err, errcodes.InternalServerError,
			fmt.Sprintf("failed to merge pull request %s", prId))
	}
	return mergedPR, nil
}

func (s *PullRequestService) Reassign(ctx context.Context, prId string, oldId string) (entity.PullRequest, string, error) {
	pr, newId, err := s.prRepo.Reassign(ctx, prId, oldId)
	if err != nil {
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			return entity.PullRequest{}, "", err
		}
		return entity.PullRequest{}, "", domain.WrapError(err, errcodes.InternalServerError,
			fmt.Sprintf("failed to reassign reviewer for pull request %s", prId))
	}
	return pr, newId, nil
}

func (s *PullRequestService) GetUserReviews(ctx context.Context, userId string) ([]entity.PullRequest, error) {
	return s.prRepo.GetUserReviews(ctx, userId)
}

func (s *PullRequestService) StartEventWorker(ctx context.Context) {
	logger(ctx).Info("Starting PR event worker...")
	for {
		select {
		case <-ctx.Done():
			logger(ctx).Info("Stopping PR event worker...")
			return
		case job := <-s.eventChan:
			if job.value {
				err := s.prRepo.AssignToNeedyPRs(ctx, job.userId)
				if err != nil {
					logger(ctx).Error("Failed to assign needy PRs to needy PRs", err.Error())
				}

			} else {
				err := s.prRepo.ReassignFromAllPRs(ctx, job.userId)
				if err != nil {
					logger(ctx).Error("Failed to reassign from all PRs to needy PRs", err.Error())
				}
			}
		}
	}
}
