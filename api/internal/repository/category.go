package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neto-app/neto/api/internal/domain/entity"
)

// CategoryRepository implementa domain/repository.CategoryRepository usando pgx.
type CategoryRepository struct {
	pool *pgxpool.Pool
}

func newCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

// Create inserta una nueva categoría.
func (r *CategoryRepository) Create(ctx context.Context, category entity.Category) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO categories (id, user_id, name, parent_id, icon, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		category.ID,
		category.UserID,
		category.Name,
		category.ParentID,
		category.Icon,
		category.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create category: %w", err)
	}
	return nil
}

// GetByID busca una categoría por ID.
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (entity.Category, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, parent_id, icon, created_at
		FROM categories
		WHERE id = $1
	`, id)
	return scanCategory(row)
}

// GetByUserID retorna todas las categorías del usuario.
func (r *CategoryRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, parent_id, icon, created_at
		FROM categories
		WHERE user_id = $1
		ORDER BY name ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: get categories by user: %w", err)
	}
	defer rows.Close()

	var categories []entity.Category
	for rows.Next() {
		cat, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, rows.Err()
}

// GetChildren retorna todas las subcategorías de un padre.
func (r *CategoryRepository) GetChildren(ctx context.Context, parentID uuid.UUID) ([]entity.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, name, parent_id, icon, created_at
		FROM categories
		WHERE parent_id = $1
		ORDER BY name ASC
	`, parentID)
	if err != nil {
		return nil, fmt.Errorf("repository: get children categories: %w", err)
	}
	defer rows.Close()

	var categories []entity.Category
	for rows.Next() {
		cat, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, rows.Err()
}

// Update actualiza el nombre e ícono de una categoría.
func (r *CategoryRepository) Update(ctx context.Context, category entity.Category) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE categories SET name = $1, icon = $2
		WHERE id = $3 AND user_id = $4
	`, category.Name, category.Icon, category.ID, category.UserID)
	if err != nil {
		return fmt.Errorf("repository: update category: %w", err)
	}
	return nil
}

// Delete elimina una categoría por ID.
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("repository: delete category: %w", err)
	}
	return nil
}

// scanCategory lee una fila de categories.
func scanCategory(row scanner) (entity.Category, error) {
	var (
		id        uuid.UUID
		userID    uuid.UUID
		name      string
		parentID  *uuid.UUID
		icon      string
		createdAt time.Time
	)

	if err := row.Scan(&id, &userID, &name, &parentID, &icon, &createdAt); err != nil {
		return entity.Category{}, fmt.Errorf("repository: scan category: %w", err)
	}

	return entity.Category{
		ID:        id,
		UserID:    userID,
		Name:      name,
		ParentID:  parentID,
		Icon:      icon,
		CreatedAt: createdAt,
	}, nil
}
