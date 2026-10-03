package categories

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInvalidCategory  = errors.New("invalid category")
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryConflict = errors.New("category already exists")
)

type categoryRepository interface {
	CreateDefaults(context.Context, string) (int64, error)
	List(context.Context, string) ([]Category, error)
	Create(context.Context, string, CreateParams) (Category, error)
	Update(context.Context, string, string, UpdateParams) (Category, error)
	Delete(context.Context, string, string) (bool, error)
}

type Service struct {
	repository categoryRepository
}

func NewService(repository categoryRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, userID string) ([]Category, error) {
	if _, err := s.repository.CreateDefaults(ctx, userID); err != nil {
		return nil, err
	}
	return s.repository.List(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (Category, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Type = strings.ToLower(strings.TrimSpace(input.Type))
	input.Icon = strings.TrimSpace(input.Icon)
	input.Color = strings.TrimSpace(input.Color)
	if !validName(input.Name) || !validType(input.Type) || !validOptional(input.Icon) || !validOptional(input.Color) {
		return Category{}, ErrInvalidCategory
	}
	category, err := s.repository.Create(ctx, userID, CreateParams(input))
	if isUniqueViolation(err) {
		return Category{}, ErrCategoryConflict
	}
	return category, err
}

func (s *Service) Update(ctx context.Context, userID, categoryID string, input UpdateInput) (Category, error) {
	if input.Name == nil && input.Type == nil && input.Icon == nil && input.Color == nil {
		return Category{}, ErrInvalidCategory
	}
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if !validName(value) {
			return Category{}, ErrInvalidCategory
		}
		input.Name = &value
	}
	if input.Type != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Type))
		if !validType(value) {
			return Category{}, ErrInvalidCategory
		}
		input.Type = &value
	}
	for _, value := range []*string{input.Icon, input.Color} {
		if value != nil {
			trimmed := strings.TrimSpace(*value)
			if !validOptional(trimmed) || trimmed == "" {
				return Category{}, ErrInvalidCategory
			}
			*value = trimmed
		}
	}
	category, err := s.repository.Update(ctx, userID, categoryID, UpdateParams(input))
	if errors.Is(err, pgx.ErrNoRows) {
		return Category{}, ErrCategoryNotFound
	}
	if isUniqueViolation(err) {
		return Category{}, ErrCategoryConflict
	}
	return category, err
}

func (s *Service) Delete(ctx context.Context, userID, categoryID string) error {
	deleted, err := s.repository.Delete(ctx, userID, categoryID)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrCategoryNotFound
	}
	return nil
}

func validName(value string) bool {
	return value != "" && len([]rune(value)) <= 100
}

func validType(value string) bool {
	return value == "income" || value == "expense"
}

func validOptional(value string) bool {
	return len([]rune(value)) <= 100
}

func isUniqueViolation(err error) bool {
	var databaseError *pgconn.PgError
	return errors.As(err, &databaseError) && databaseError.Code == "23505"
}
