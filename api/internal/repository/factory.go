// Package repository provee implementaciones concretas de las interfaces de dominio.
package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositories agrupa todos los repositorios de la aplicación.
type Repositories struct {
	Account        *AccountRepository
	Category       *CategoryRepository
	Transaction    *TransactionRepository
	Currency       *CurrencyRepository
	AIConversation *AIConversationRepository
	AIMessage      *AIMessageRepository
}

// NewRepositories instancia todos los repositorios con el pool dado.
func NewRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Account:        newAccountRepository(pool),
		Category:       newCategoryRepository(pool),
		Transaction:    newTransactionRepository(pool),
		Currency:       newCurrencyRepository(pool),
		AIConversation: newAIConversationRepository(pool),
		AIMessage:      newAIMessageRepository(pool),
	}
}
