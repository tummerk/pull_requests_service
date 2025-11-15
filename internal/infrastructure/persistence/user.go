package persistence

import (
	"context"
	"database/sql" // Нужен для проверки sql.ErrNoRows
	"errors"       // Нужен для errors.Is
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq" // Нужен для проверки кода ошибки Postgres
	"pull_requests_service/internal/domain"
	"pull_requests_service/internal/domain/entity"
	"pull_requests_service/pkg/errcodes"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func isUniqueConstraintError(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func (r *UserRepository) Create(ctx context.Context, user entity.User) (entity.User, error) {
	query := `
        INSERT INTO users (id, name, is_active, team_id)
        VALUES (:id, :name, :is_active, :team_id)
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            is_active = EXCLUDED.is_active,
            team_id = EXCLUDED.team_id
        RETURNING id, name, is_active, team_id, created_at;
    `

	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		if isUniqueConstraintError(err) {
			return entity.User{}, domain.NewError(errcodes.UserAlreadyExists, fmt.Sprintf("user with id '%s' already exists", user.Id))
		}
		return entity.User{}, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to create/update user")
	}
	defer rows.Close()

	if rows.Next() {
		var createdUser entity.User
		if err := rows.StructScan(&createdUser); err != nil {
			return entity.User{}, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to scan created user")
		}
		return createdUser, nil
	}

	return entity.User{}, domain.NewError(errcodes.InternalServerError, "repository: user was not created or updated, no rows returned")
}

func (r *UserRepository) GetById(ctx context.Context, userId string) (entity.User, error) {
	query := `SELECT id, name, is_active, team_id, created_at FROM users WHERE id = $1`

	var foundUser entity.User
	err := r.db.GetContext(ctx, &foundUser, query, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, domain.NewError(errcodes.NotFound, fmt.Sprintf("user with id '%s' not found", userId))
		}
		return entity.User{}, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to get user by id")
	}

	return foundUser, nil
}

func (r *UserRepository) GetByTeam(ctx context.Context, teamName string) ([]entity.User, error) {
	query := `SELECT id, name, is_active, team_id, created_at FROM users WHERE team_id = $1`

	var users []entity.User
	err := r.db.SelectContext(ctx, &users, query, teamName)
	if err != nil {
		return nil, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to get user by team")
	}

	return users, nil
}

func (r *UserRepository) SetIsActive(ctx context.Context, userId string, isActive bool) (entity.User, error) {
	query := `
        UPDATE users
        SET is_active = $1
        WHERE id = $2
        RETURNING id, name, is_active, team_id, created_at;
    `

	var updatedUser entity.User
	err := r.db.GetContext(ctx, &updatedUser, query, isActive, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, domain.NewError(errcodes.NotFound, fmt.Sprintf("user with id '%s' not found for update", userId))
		}
		return entity.User{}, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to set user active status")
	}

	return updatedUser, nil
}

func (r *UserRepository) GetActiveTeamCandidatesId(ctx context.Context, authorID string) ([]string, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, domain.WrapError(err, errcodes.InternalServerError, "failed to begin transaction")
	}
	defer tx.Rollback()

	var authorExists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err = tx.GetContext(ctx, &authorExists, checkQuery, authorID)
	if err != nil {
		return nil, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to check author existence")
	}

	if !authorExists {
		return nil, domain.NewError(errcodes.NotFound, fmt.Sprintf("author with id '%s' not found", authorID))
	}

	query := `
        SELECT id
        FROM users
        WHERE team_id = (SELECT team_id FROM users WHERE id = $1)
          AND is_active = TRUE
          AND id != $1
        ORDER BY RANDOM()
        LIMIT 2;
    `
	var candidateIDs []string
	err = tx.SelectContext(ctx, &candidateIDs, query, authorID)
	if err != nil {
		return nil, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to get active team candidates")
	}

	// Вся остальная логика остается без изменений
	if err = tx.Commit(); err != nil {
		return nil, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to commit transaction")
	}

	return candidateIDs, nil
}

func (r *UserRepository) GetUserReviews(ctx context.Context, userId string) ([]entity.PullRequest, error) {
	const query = `
        SELECT
            pr.id,
            pr.name,
            pr.author_id,
            pr.status
        FROM
            pull_requests pr
        WHERE
            EXISTS (
                SELECT 1
                FROM pr_reviewers r
                WHERE r.pull_request_id = pr.id AND r.reviewer_id = $1
            )
    `

	var reviews []entity.PullRequest

	err := r.db.SelectContext(ctx, &reviews, query, userId)
	if err != nil {
		return nil, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to get user reviews")
	}

	return reviews, nil
}

func (r *UserRepository) GetUserAssignmentStats(ctx context.Context) ([]entity.UserAssignmentStat, error) {
	const query = `
        SELECT
            u.id AS user_id,
            u.name AS username,
            COUNT(pr.reviewer_id) AS assignment_count
        FROM
            users u
        LEFT JOIN
            pr_reviewers pr ON u.id = pr.reviewer_id
        GROUP BY
            u.id, u.name
        ORDER BY
            assignment_count DESC, u.name ASC;
    `

	var stats []entity.UserAssignmentStat

	err := r.db.SelectContext(ctx, &stats, query)
	if err != nil {
		return nil, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to get user assignment stats")
	}
	return stats, nil
}
