package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"pull_requests_service/internal/domain/entity"
	"pull_requests_service/pkg/errcodes"
)

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team entity.Team) (entity.Team, error) {
	query := `
        INSERT INTO teams (name)
        VALUES ($1)
        RETURNING name, created_at;
    `
	var createdTeam entity.Team

	err := r.db.GetContext(ctx, &createdTeam, query, team.Name)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.Team{}, fmt.Errorf("repository.Create: %w. Team %s", errcodes.TeamAlreadyExists, team.Name)
		}
		return entity.Team{}, fmt.Errorf("repository.Create: failed to create team: %w", err)
	}

	return createdTeam, nil
}

func (r *TeamRepository) Get(ctx context.Context, name string) (entity.Team, error) {
	query := `SELECT name, created_at FROM teams WHERE name = $1`

	var foundTeam entity.Team

	err := r.db.GetContext(ctx, &foundTeam, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Team{}, fmt.Errorf("repository.Get: %w: team '%s'", errcodes.NotFound, name)
		}
		return entity.Team{}, fmt.Errorf("repository.Get: failed to get team: %w", err)
	}
	return foundTeam, nil
}
