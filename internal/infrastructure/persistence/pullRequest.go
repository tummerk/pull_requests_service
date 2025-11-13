package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"pull_requests_service/internal/domain"
	"pull_requests_service/internal/domain/entity"
	"pull_requests_service/pkg/errcodes"
)

type PullRequestRepository struct {
	db *sqlx.DB
}

func NewPullRequestRepository(db *sqlx.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) CreateWithReviewers(ctx context.Context, pr *entity.PullRequest, reviewerIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return domain.WrapError(err, errcodes.InternalServerError, "failed to begin transaction")
	}
	defer tx.Rollback()

	prQuery := `
        INSERT INTO pull_requests (id, name, author_id, status)
        VALUES ($1, $2, $3, $4)
        RETURNING id, name, author_id, status, created_at, merged_at;
    `
	err = tx.GetContext(ctx, pr, prQuery, pr.Id, pr.Name, pr.AuthorId, pr.Status)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.NewError(errcodes.PullRequestExists, fmt.Sprintf("pull request with id '%s' already exists", pr.Id))
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return domain.NewError(errcodes.NotFound, fmt.Sprintf("author with id '%s' not found", pr.AuthorId))
		}
		return domain.WrapError(err, errcodes.InternalServerError, "failed to create pull request")
	}

	if len(reviewerIDs) > 0 {
		type reviewerLink struct {
			PRID       string `db:"pr_id"`
			ReviewerID string `db:"reviewer_id"`
		}

		links := make([]reviewerLink, len(reviewerIDs))
		for i, reviewerID := range reviewerIDs {
			links[i] = reviewerLink{
				PRID:       pr.Id,
				ReviewerID: reviewerID,
			}
		}

		assignQuery := `INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES (:pr_id, :reviewer_id)`
		_, err = tx.NamedExecContext(ctx, assignQuery, links)
		if err != nil {
			var pgErr *pq.Error
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return domain.NewError(errcodes.NotFound, "one of the reviewers not found")
			}
			return domain.WrapError(err, errcodes.InternalServerError, "failed to assign reviewers")
		}
	}

	if err = tx.Commit(); err != nil {
		return domain.WrapError(err, errcodes.InternalServerError, "failed to commit transaction")
	}
	pr.AssignedReviewers = reviewerIDs
	return nil
}
func (r *PullRequestRepository) Merge(ctx context.Context, prId string) (entity.PullRequest, error) {
	var pr entity.PullRequest
	queryUpdate := `
		UPDATE pull_requests
		SET status = $1, updated_at = NOW(), merged_at = NOW()
		WHERE id = $2 AND status = $3
		RETURNING *`

	err := r.db.QueryRowxContext(ctx, queryUpdate, entity.StatusMerged, prId, entity.StatusOpen).StructScan(&pr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.PullRequest{}, fmt.Errorf("failed to merge: pr %s not found or already merged", prId)
		}
		return entity.PullRequest{}, fmt.Errorf("failed to merge pr %s: %w", prId, err)
	}

	return pr, nil
}

func (r *PullRequestRepository) Reassign(ctx context.Context, prId, oldReviewerId string) (
	entity.PullRequest, string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return entity.PullRequest{}, "", domain.WrapError(err, errcodes.InternalServerError, "failed to begin transaction")
	}
	defer tx.Rollback()

	var pr entity.PullRequest
	prQuery := `SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests WHERE id = $1 FOR UPDATE`
	err = tx.GetContext(ctx, &pr, prQuery, prId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.PullRequest{}, "", domain.NewError(errcodes.NotFound, "pull request not found")
		}
		return entity.PullRequest{}, "", domain.WrapError(err, errcodes.InternalServerError, "failed to get pull request")
	}

	if pr.Status == entity.StatusMerged {
		return entity.PullRequest{}, "", domain.NewError(errcodes.PrMerged, "cannot reassign on merged PR")
	}

	err = tx.SelectContext(ctx, &pr.AssignedReviewers, `SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $1`, prId)
	if err != nil {
		return entity.PullRequest{}, "", domain.WrapError(err, errcodes.InternalServerError, "failed to get current reviewers")
	}
	if pr.AssignedReviewers == nil {
		pr.AssignedReviewers = []string{}
	}

	var newReviewerId string
	findReviewerQuery := `
		SELECT u.id
		FROM users u
		JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_name = (
			SELECT tm2.team_name
			FROM team_members tm2
			WHERE tm2.user_id = $2
		)
		AND u.is_active = TRUE
		AND u.id NOT IN (
			SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $1
			UNION
			SELECT author_id FROM pull_requests WHERE id = $1
		)
		ORDER BY RANDOM()
		LIMIT 1`

	err = tx.GetContext(ctx, &newReviewerId, findReviewerQuery, prId, oldReviewerId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.PullRequest{}, "", domain.NewError(errcodes.NoCandidate, "old reviewer not assigned or no replacement candidate")
		}
		return entity.PullRequest{}, "", domain.WrapError(err, errcodes.InternalServerError, "failed to find replacement")
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = $2`, prId, oldReviewerId)
	if err != nil {
		return entity.PullRequest{}, "", domain.WrapError(err, errcodes.InternalServerError, "failed to remove old reviewer")
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2)`, prId, newReviewerId)
	if err != nil {
		return entity.PullRequest{}, "", domain.WrapError(err, errcodes.InternalServerError, "failed to assign new reviewer")
	}

	updatedReviewers := make([]string, 0, len(pr.AssignedReviewers))
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer != oldReviewerId {
			updatedReviewers = append(updatedReviewers, reviewer)
		}
	}
	pr.AssignedReviewers = append(updatedReviewers, newReviewerId)

	if err = tx.Commit(); err != nil {
		return entity.PullRequest{}, "", domain.WrapError(err, errcodes.InternalServerError, "failed to commit transaction")
	}

	return pr, newReviewerId, nil
}
