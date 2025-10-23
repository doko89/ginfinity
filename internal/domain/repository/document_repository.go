package repository

import (
	"context"

	"gin-boilerplate/internal/domain/entity"
)

type DocumentRepository interface {
	Create(ctx context.Context, document *entity.Document) error
	FindByID(ctx context.Context, id string) (*entity.Document, error)
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Document, error)
	Update(ctx context.Context, document *entity.Document) error
	Delete(ctx context.Context, id string) error
	GetFileURL(ctx context.Context, id string) (string, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
}