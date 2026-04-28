package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/neto-app/neto/api/internal/domain/entity"
	domainrepo "github.com/neto-app/neto/api/internal/domain/repository"
)

// CategoryUseCase contiene la lógica de aplicación para categorías.
type CategoryUseCase struct {
	categories domainrepo.CategoryRepository
}

// NewCategoryUseCase crea un nuevo CategoryUseCase.
func NewCategoryUseCase(categories domainrepo.CategoryRepository) *CategoryUseCase {
	return &CategoryUseCase{categories: categories}
}

// CreateCategory crea una nueva categoría para el usuario.
func (uc *CategoryUseCase) CreateCategory(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID) (entity.Category, error) {
	category, err := entity.NewCategory(userID, name, parentID)
	if err != nil {
		return entity.Category{}, fmt.Errorf("usecase: create category entity: %w", err)
	}

	if err := uc.categories.Create(ctx, category); err != nil {
		return entity.Category{}, fmt.Errorf("usecase: persist category: %w", err)
	}

	return category, nil
}

// ListCategories retorna todas las categorías del usuario.
func (uc *CategoryUseCase) ListCategories(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	categories, err := uc.categories.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("usecase: list categories: %w", err)
	}
	return categories, nil
}

// GetCategoryTree retorna las categorías raíz del usuario con sus hijos.
func (uc *CategoryUseCase) GetCategoryTree(ctx context.Context, userID uuid.UUID) ([]entity.Category, error) {
	// Retorna todas las categorías — el árbol se construye en el handler si se necesita
	return uc.ListCategories(ctx, userID)
}
