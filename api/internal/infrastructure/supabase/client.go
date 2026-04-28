// Package supabase provee un cliente que wrappea el pool de pgx para Supabase.
package supabase

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client es un wrapper sobre pgxpool.Pool que representa la conexión a Supabase.
type Client struct {
	pool *pgxpool.Pool
}

// NewClient crea un nuevo Client de Supabase a partir de un pool existente.
func NewClient(pool *pgxpool.Pool) *Client {
	return &Client{pool: pool}
}

// Pool retorna el pool subyacente de pgx para uso directo en repositorios.
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}
