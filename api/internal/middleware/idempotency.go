package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// cachedResponse almacena una respuesta HTTP para reutilizarla.
type cachedResponse struct {
	Status int
	Body   []byte
}

// responseRecorder captura la respuesta del handler para poder cachearla.
type responseRecorder struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (rr *responseRecorder) WriteHeader(status int) {
	rr.status = status
	// No propagamos el header todavía — esperamos capturar el body
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	return rr.buf.Write(b)
}

// Idempotency retorna un middleware que garantiza idempotencia para requests POST.
// Si el header Idempotency-Key está presente y la key ya existe en DB sin expirar,
// retorna la respuesta cacheada. Si no existe, ejecuta el handler y guarda el resultado.
// Si el header no se envía, continúa sin idempotencia.
func Idempotency(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Solo aplica a POST
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				// Sin header → continuar sin idempotencia
				next.ServeHTTP(w, r)
				return
			}

			// Buscar respuesta cacheada
			cached, err := getIdempotencyRecord(r.Context(), db, key)
			if err == nil && cached != nil {
				// Respuesta encontrada → retornar directamente
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Idempotent-Replayed", "true")
				w.WriteHeader(cached.Status)
				_, _ = w.Write(cached.Body)
				return
			}

			// Ejecutar el handler y capturar la respuesta
			rec := newResponseRecorder(w)
			next.ServeHTTP(rec, r)

			// Guardar en DB solo si fue exitoso (2xx)
			if rec.status >= 200 && rec.status < 300 {
				_ = saveIdempotencyRecord(r.Context(), db, key, rec.status, rec.buf.Bytes())
			}

			// Propagar la respuesta real
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(rec.status)
			_, _ = w.Write(rec.buf.Bytes())
		})
	}
}

// getIdempotencyRecord busca un registro de idempotencia en la DB que no haya expirado.
func getIdempotencyRecord(ctx context.Context, db *pgxpool.Pool, key string) (*cachedResponse, error) {
	var status int
	var body []byte

	err := db.QueryRow(ctx, `
		SELECT response_status, response_body
		FROM idempotency_keys
		WHERE key = $1 AND expires_at > NOW()
	`, key).Scan(&status, &body)
	if err != nil {
		return nil, err
	}

	return &cachedResponse{Status: status, Body: body}, nil
}

// saveIdempotencyRecord persiste la respuesta en la tabla idempotency_keys.
// La expiración se fija en 24 horas (estándar de la industria).
func saveIdempotencyRecord(ctx context.Context, db *pgxpool.Pool, key string, status int, body []byte) error {
	bodyJSON := json.RawMessage(body)
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	_, err := db.Exec(ctx, `
		INSERT INTO idempotency_keys (key, response_status, response_body, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key) DO NOTHING
	`, key, status, bodyJSON, expiresAt)

	return err
}
