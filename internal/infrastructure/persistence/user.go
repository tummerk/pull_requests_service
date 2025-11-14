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
        INSERT INTO users (id, username, is_active, team_name)
        VALUES (:id, :username, :is_active, :team_name)
        ON CONFLICT (id) DO UPDATE SET
            username = EXCLUDED.username,
            is_active = EXCLUDED.is_active,
            team_name = EXCLUDED.team_name,
			updated_at = CURRENT_TIMESTAMP
        RETURNING id, username, is_active, team_name, created_at, updated_at;
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
	query := `SELECT id, username, is_active, team_name, created_at, updated_at FROM users WHERE id = $1`

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
	query := `SELECT id, username, is_active, team_name, created_at, updated_at FROM users WHERE team_name = $1`

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
        SET is_active = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        RETURNING id, username, is_active, team_name, created_at, updated_at;
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
	query := `
        WITH author_team AS (
            SELECT team_name FROM users WHERE id = $1
        )
        SELECT id
        FROM users
        WHERE team_name = (SELECT team_name FROM author_team)
          AND is_active = TRUE
          AND id != $1
        ORDER BY RANDOM()
        LIMIT 2;
    `

	var candidateIDs []string
	err := r.db.SelectContext(ctx, &candidateIDs, query, authorID)
	if err != nil {
		return nil, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to get active team candidates")
	}

	return candidateIDs, nil
}
