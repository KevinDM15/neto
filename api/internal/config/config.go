// Package config gestiona la configuración de la aplicación desde variables de entorno.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config contiene toda la configuración necesaria para levantar el servidor.
type Config struct {
	Port            string
	Env             string
	DatabaseURL     string
	SupabaseURL     string
	SupabaseAnonKey string
	AnthropicAPIKey string
}

// JWKSUrl retorna el endpoint JWKS de Supabase derivado del SupabaseURL.
// Supabase publica sus claves públicas en /auth/v1/.well-known/jwks.json.
func (c Config) JWKSUrl() string {
	return c.SupabaseURL + "/auth/v1/.well-known/jwks.json"
}

// Load lee la configuración desde variables de entorno.
// Intenta cargar un archivo .env si existe (útil en desarrollo).
// Retorna error si alguna variable obligatoria no está definida.
func Load() (Config, error) {
	// godotenv no falla si el archivo no existe — útil en CI/prod
	_ = godotenv.Load()

	cfg := Config{
		Port:            getEnv("PORT", "8080"),
		Env:             getEnv("ENV", "development"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		SupabaseURL:     os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey: os.Getenv("SUPABASE_ANON_KEY"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// validate verifica que todas las variables obligatorias estén presentes.
func (c Config) validate() error {
	required := map[string]string{
		"DATABASE_URL":      c.DatabaseURL,
		"SUPABASE_URL":      c.SupabaseURL,
		"SUPABASE_ANON_KEY": c.SupabaseAnonKey,
		"ANTHROPIC_API_KEY": c.AnthropicAPIKey,
	}

	for name, val := range required {
		if val == "" {
			return fmt.Errorf("required environment variable %s is not set", name)
		}
	}

	return nil
}

// getEnv retorna el valor de la variable de entorno o el default si no existe.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
