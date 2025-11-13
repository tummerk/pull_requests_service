package persistence

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"pull_requests_service/internal/domain"
	"pull_requests_service/internal/domain/entity"
	"pull_requests_service/pkg/errcodes"
	"time"
)

type PullRequestRepository struct {
	db *sqlx.DB
}

func NewPullRequestRepository(db *sqlx.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) CreateWithReviewers(ctx context.Context, pr *entity.PullRequest, reviewerIDs []string) (*entity.PullRequest, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, domain.WrapError(err, errcodes.InternalServerError, "failed to begin transaction")
	}
	defer tx.Rollback()

	prQuery := `
        INSERT INTO pull_requests (id, name, author_id, status)
        VALUES ($1, $2, $3, $4)
        RETURNING id, name, author_id, status, created_at, merged_at;
    `
	var createdPR entity.PullRequest
	err = tx.GetContext(ctx, &createdPR, prQuery, pr.Id, pr.Name, pr.AuthorId, pr.Status)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.NewError(errcodes.PullRequestExists, fmt.Sprintf("pull request with id '%s' already exists", pr.Id))
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, domain.NewError(errcodes.NotFound, fmt.Sprintf("author with id '%s' not found", pr.AuthorId))
		}
		return nil, domain.WrapError(err, errcodes.InternalServerError, "failed to create pull request")
	}

	if len(reviewerIDs) > 0 {

		type reviewerLink struct {
			PRID       string `db:"pr_id"`
			ReviewerID string `db:"reviewer_id"`
		}

		links := make([]reviewerLink, len(reviewerIDs))
		for i, reviewerID := range reviewerIDs {
			links[i] = reviewerLink{
				PRID:       createdPR.Id,
				ReviewerID: reviewerID,
			}
		}

		assignQuery := `INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES (:pr_id, :reviewer_id)`
		_, err = tx.NamedExecContext(ctx, assignQuery, links)
		if err != nil {
			var pgErr *pq.Error

			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, domain.NewError(errcodes.NotFound, "one of the reviewers not found")
			}
			return nil, domain.WrapError(err, errcodes.InternalServerError, "failed to assign reviewers")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, domain.WrapError(err, errcodes.InternalServerError, "failed to commit transaction")
	}
	createdPR.AssignedReviewers = reviewerIDs

	return &createdPR, nil
}

func (r *pullRequestRepository) Merge(ctx context.Context, prId string) (entity.PullRequest, error) {
	var pr entity.PullRequest

	queryGet := `SELECT id, is_merged, merged_at FROM pull_requests WHERE id = $1`
	err := r.db.GetContext(ctx, &pr, queryGet, prId)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) { // sqlx имеет собственный ErrNoRows (совместим с стандартным)
			return entity.PullRequest{}, fmt.Errorf("merge: %w", errcodes.NotFound)
		}
		return entity.PullRequest{}, fmt.Errorf("failed to fetch pr %s: %w", prId, err)
	}

	if pr.IsMerged {
		return entity.PullRequest{}, fmt.Errorf("pr %s is already merged", prId)
	}


	now := time.Now()
	pr.Status =
	pr. = &now

	queryUpdate := `
		UPDATE pull_requests
		SET is_merged = \(1, merged_at = \)2
		WHERE id = $3
		RETURNING id, is_merged, merged_at` // Возвращаем обновленные данные

	// Используем QueryRowxContext для сканирования обновленного PR
	err = r.db.QueryRowxContext(ctx, queryUpdate, pr.IsMerged, pr.MergedAt, prId).StructScan(&pr)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("failed to merge pr %s: %w", prId, err)
	}

	// Возвращаем уже слиянный PR
	return pr, nil
}