package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"pull_requests_service/internal/domain"
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.Team{}, domain.NewError(errcodes.TeamAlreadyExists, fmt.Sprintf("team with name '%s' already exists", team.Name))
		}
		return entity.Team{}, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to create team")
	}

	return createdTeam, nil
}

func (r *TeamRepository) Get(ctx context.Context, name string) (entity.Team, error) {
	query := `SELECT name, created_at FROM teams WHERE name = $1`

	var foundTeam entity.Team

	err := r.db.GetContext(ctx, &foundTeam, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Team{}, domain.NewError(errcodes.NotFound, fmt.Sprintf("team with name '%s' not found", name))
		}
		return entity.Team{}, domain.WrapError(err, errcodes.InternalServerError, "repository: failed to get team")
	}
	return foundTeam, nil
}
