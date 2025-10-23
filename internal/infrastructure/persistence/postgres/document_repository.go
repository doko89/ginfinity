package postgres

import (
	"context"

	"gin-boilerplate/internal/domain/entity"
	"gin-boilerplate/internal/domain/repository"

	"gorm.io/gorm"
)

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) repository.DocumentRepository {
	return &documentRepository{
		db: db,
	}
}

func (r *documentRepository) Create(ctx context.Context, document *entity.Document) error {
	return r.db.WithContext(ctx).Create(document).Error
}

func (r *documentRepository) FindByID(ctx context.Context, id string) (*entity.Document, error) {
	var document entity.Document
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&document).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &document, nil
}

func (r *documentRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Document, error) {
	var documents []*entity.Document
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&documents).Error
	return documents, err
}

func (r *documentRepository) Update(ctx context.Context, document *entity.Document) error {
	return r.db.WithContext(ctx).Save(document).Error
}

func (r *documentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Document{}, "id = ?", id).Error
}

func (r *documentRepository) GetFileURL(ctx context.Context, id string) (string, error) {
	var fileURL string
	err := r.db.WithContext(ctx).
		Model(&entity.Document{}).
		Where("id = ?", id).
		Select("file_url").
		Scan(&fileURL).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", gorm.ErrRecordNotFound
		}
		return "", err
	}
	return fileURL, nil
}

func (r *documentRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Document{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}