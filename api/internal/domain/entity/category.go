package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Category representa una categoría para clasificar transacciones.
// Puede ser raíz (ParentID == nil) o subcategoría.
type Category struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	ParentID  *uuid.UUID
	Icon      string
	CreatedAt time.Time
}

// NewCategory crea una nueva Category. parentID puede ser nil para categorías raíz.
func NewCategory(userID uuid.UUID, name string, parentID *uuid.UUID) (Category, error) {
	if name == "" {
		return Category{}, fmt.Errorf("category name cannot be empty")
	}

	return Category{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		ParentID:  parentID,
		CreatedAt: time.Now().UTC(),
	}, nil
}

// IsRoot retorna true si la categoría no tiene padre.
func (c Category) IsRoot() bool {
	return c.ParentID == nil
}
