// Package client provides types used by the Neto HTTP client.
package client

// SupabaseTokenResponse is the response from Supabase Auth token endpoint.
type SupabaseTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// ChatRequest is the body sent to POST /api/v1/chat.
type ChatRequest struct {
	ConversationID string      `json:"conversation_id,omitempty"`
	Message        string      `json:"message"`
	Confirm        bool        `json:"confirm"`
	PendingTool    interface{} `json:"pending_tool,omitempty"`
}

// PendingConfirmation represents a tool call awaiting user confirmation.
type PendingConfirmation struct {
	Tool    string                 `json:"tool"`
	Preview map[string]interface{} `json:"preview"`
}

// ChatResponse is the response from POST /api/v1/chat.
type ChatResponse struct {
	ConversationID      string               `json:"conversation_id"`
	Reply               string               `json:"reply"`
	PendingConfirmation *PendingConfirmation `json:"pending_confirmation,omitempty"`
}

// Account represents a user account returned by GET /api/v1/accounts.
type Account struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}
