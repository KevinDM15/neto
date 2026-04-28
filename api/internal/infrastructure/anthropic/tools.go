package anthropic

import (
	"encoding/json"

	"github.com/neto-app/neto/api/internal/ai"
)

// toolSchema construye un JSON Schema básico para los inputs de los tools.
func mustRawJSON(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic("anthropic: mustRawJSON: " + err.Error())
	}
	return b
}

// RequiresConfirmation retorna true si el tool dado requiere confirmación del usuario
// antes de ejecutarse.
func RequiresConfirmation(toolName string) bool {
	switch toolName {
	case "create_transaction", "create_account", "record_debt", "delete_transaction":
		return true
	default:
		return false
	}
}

// ToolCatalog retorna el catálogo completo de tools disponibles para el agente Neto.
func ToolCatalog() []ai.Tool {
	return []ai.Tool{
		{
			Name:        "create_transaction",
			Description: "Registra un nuevo movimiento de dinero (ingreso o egreso) en una cuenta del usuario. Requiere confirmación antes de ejecutarse.",
			InputSchema: mustRawJSON(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account_id": map[string]interface{}{
						"type":        "string",
						"format":      "uuid",
						"description": "UUID de la cuenta donde se registra la transacción.",
					},
					"amount": map[string]interface{}{
						"type":        "string",
						"description": "Monto de la transacción como string decimal (ej: '50000.00').",
					},
					"currency_code": map[string]interface{}{
						"type":        "string",
						"description": "Código de moneda ISO 4217 (ej: 'ARS', 'USD'). Si se omite, se usa la moneda de la cuenta.",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"income", "expense", "transfer"},
						"description": "Tipo de transacción.",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Descripción breve del movimiento.",
					},
					"category_id": map[string]interface{}{
						"type":        "string",
						"format":      "uuid",
						"description": "UUID de la categoría (opcional).",
					},
				},
				"required": []string{"account_id", "amount", "type", "description"},
			}),
		},
		{
			Name:        "list_transactions",
			Description: "Lista las transacciones del usuario con filtros opcionales por cuenta, categoría y rango de fechas.",
			InputSchema: mustRawJSON(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account_id": map[string]interface{}{
						"type":        "string",
						"format":      "uuid",
						"description": "Filtrar por cuenta (opcional).",
					},
					"category_id": map[string]interface{}{
						"type":        "string",
						"format":      "uuid",
						"description": "Filtrar por categoría (opcional).",
					},
					"from": map[string]interface{}{
						"type":        "string",
						"format":      "date",
						"description": "Fecha de inicio del rango, formato YYYY-MM-DD (opcional).",
					},
					"to": map[string]interface{}{
						"type":        "string",
						"format":      "date",
						"description": "Fecha de fin del rango, formato YYYY-MM-DD (opcional).",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Máximo de resultados a retornar (default: 20).",
					},
				},
				"required": []string{},
			}),
		},
		{
			Name:        "get_balance",
			Description: "Retorna el saldo actual de todas las cuentas del usuario o de una cuenta específica.",
			InputSchema: mustRawJSON(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account_id": map[string]interface{}{
						"type":        "string",
						"format":      "uuid",
						"description": "UUID de la cuenta a consultar (opcional — si se omite retorna todas).",
					},
				},
				"required": []string{},
			}),
		},
		{
			Name:        "create_account",
			Description: "Crea una nueva cuenta financiera para el usuario (ej: cuenta bancaria, billetera). Requiere confirmación.",
			InputSchema: mustRawJSON(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Nombre descriptivo de la cuenta.",
					},
					"currency_code": map[string]interface{}{
						"type":        "string",
						"description": "Código de moneda ISO 4217 (ej: 'ARS', 'USD').",
					},
				},
				"required": []string{"name", "currency_code"},
			}),
		},
		{
			Name:        "list_categories",
			Description: "Lista todas las categorías de gastos/ingresos disponibles para el usuario.",
			InputSchema: mustRawJSON(map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			}),
		},
		{
			Name:        "get_monthly_summary",
			Description: "Retorna un resumen financiero mensual: total ingresos, total egresos y balance neto.",
			InputSchema: mustRawJSON(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"year": map[string]interface{}{
						"type":        "integer",
						"description": "Año del resumen (ej: 2025).",
					},
					"month": map[string]interface{}{
						"type":        "integer",
						"description": "Mes del resumen 1-12.",
					},
				},
				"required": []string{"year", "month"},
			}),
		},
		{
			Name:        "record_debt",
			Description: "Registra una deuda (a pagar o a cobrar). Requiere confirmación.",
			InputSchema: mustRawJSON(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"counterpart_name": map[string]interface{}{
						"type":        "string",
						"description": "Nombre de la persona o entidad con quien es la deuda.",
					},
					"amount": map[string]interface{}{
						"type":        "string",
						"description": "Monto de la deuda como string decimal.",
					},
					"currency_code": map[string]interface{}{
						"type":        "string",
						"description": "Código de moneda ISO 4217.",
					},
					"direction": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"owed_by_me", "owed_to_me"},
						"description": "Dirección de la deuda: 'owed_by_me' si le debo, 'owed_to_me' si me deben.",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Descripción del motivo de la deuda.",
					},
					"due_date": map[string]interface{}{
						"type":        "string",
						"format":      "date",
						"description": "Fecha de vencimiento YYYY-MM-DD (opcional).",
					},
				},
				"required": []string{"counterpart_name", "amount", "currency_code", "direction", "description"},
			}),
		},
		{
			Name:        "update_goal_progress",
			Description: "Actualiza el progreso de una meta de ahorro registrando un nuevo aporte.",
			InputSchema: mustRawJSON(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"goal_id": map[string]interface{}{
						"type":        "string",
						"format":      "uuid",
						"description": "UUID de la meta de ahorro.",
					},
					"amount": map[string]interface{}{
						"type":        "string",
						"description": "Monto del aporte como string decimal.",
					},
				},
				"required": []string{"goal_id", "amount"},
			}),
		},
		{
			Name:        "delete_transaction",
			Description: "Elimina una transacción existente. Operación destructiva — requiere confirmación explícita.",
			InputSchema: mustRawJSON(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"transaction_id": map[string]interface{}{
						"type":        "string",
						"format":      "uuid",
						"description": "UUID de la transacción a eliminar.",
					},
				},
				"required": []string{"transaction_id"},
			}),
		},
	}
}
