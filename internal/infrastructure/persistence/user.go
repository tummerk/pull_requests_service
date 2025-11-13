package persistence

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"os/user"
	"pull_requests_service/internal/domain/entity"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user entity.User) (entity.User, error) {
	query := `
        INSERT INTO user (id, username, is_active, team_name)
        VALUES (:id, :username, :is_active, :team_name)
        ON CONFLICT (id) DO UPDATE SET
            username = EXCLUDED.username,
            is_active = EXCLUDED.is_active,
            team_name = EXCLUDED.team_name
        RETURNING id, username, is_active, team_name, created_at;
    `

	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return entity.User{}, fmt.Errorf("repository: failed to create/update user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var createdUser entity.User
		if err := rows.StructScan(&createdUser); err != nil {
			return entity.User{}, fmt.Errorf("repository: failed to scan created user: %w", err)
		}
		return createdUser, nil
	}

	return entity.User{}, fmt.Errorf("repository: user was not created or updated, no rows returned")
}

func (r *UserRepository) GetByTeam(ctx context.Context, teamName string) ([]user.User, error) {
	query := `SELECT id, username, is_active, team_name, created_at FROM user WHERE team_name = $1`

	var users []user.User
	err := r.db.SelectContext(ctx, &users, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get user by team: %w", err)
	}

	return users, nil
}

func (r *UserRepository) SetIsActive(ctx context.Context, userId string, isActive bool) (user.User, error) {
	query := `
        UPDATE user
        SET is_active = $1
        WHERE id = $2
        RETURNING id, username, is_active, team_name, created_at;
    `

	var updatedUser user.User
	err := r.db.GetContext(ctx, &updatedUser, query, isActive, userId)
	if err != nil {
		return user.User{}, fmt.Errorf("repository: failed to set user active status: %w", err)
	}

	return updatedUser, nil
}

func (r *UserRepository) GetReviewers(ctx context.Context, authorID string) ([]user.User, error) {
	query := `
        SELECT
			u2.id,
			u2.username,
			u2.is_active,
			u2.team_name,
			u2.created_at
		FROM
			users AS u1
		JOIN
			users AS u2 ON u1.team_name = u2.team_name
		WHERE
			u1.id = $1
			AND u2.is_active = TRUE
			AND u2.id != u1.id
		ORDER BY
			RANDOM() 
		LIMIT 2;
    `

	var candidates []user.User

	err := r.db.SelectContext(ctx, &candidates, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get active team candidates: %w", err)
	}

	return candidates, nil
}
