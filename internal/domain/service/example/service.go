package example

import (
	"context"
	"fmt"

	"go-backend-example/internal/domain/entity"
	"go-backend-example/internal/domain/value"
)

type exampleRepo interface {
	GetByID(context.Context, value.ExampleID) (entity.Example, error)
	Save(context.Context, entity.Example) error
}

type Service struct {
	exampleRepo exampleRepo
}

func NewService(
	exampleRepo exampleRepo,
) Service {
	return Service{
		exampleRepo: exampleRepo,
	}
}

func (s Service) GetByID(ctx context.Context, id value.ExampleID) (entity.Example, error) {
	example, err := s.exampleRepo.GetByID(ctx, id)
	if err != nil {
		return entity.Example{}, fmt.Errorf("exampleRepo.GetByID: %w", err)
	}

	return example, err
}

func (s Service) Save(ctx context.Context, example entity.Example) error {
	if err := s.exampleRepo.Save(ctx, example); err != nil {
		return fmt.Errorf("exampleRepo.Save: %w", err)
	}

	return nil
}
