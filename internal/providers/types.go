package providers

type Protocol string
type Runtime string
type AuthMode string
type BillingMode string
type CostClass string

type ReasoningLevel string

const (
	ProtocolOpenAIChat        Protocol = "openai_chat"
	ProtocolOpenAIResponses   Protocol = "openai_responses"
	ProtocolAnthropicMessages Protocol = "anthropic_messages"
	ProtocolOllamaNative      Protocol = "ollama_native"
	ProtocolCLIBacked         Protocol = "cli_backed"
	ProtocolOAuthBacked       Protocol = "oauth_backed"

	RuntimeGoNative   Runtime = "go_native"
	RuntimeNodeAISDK  Runtime = "node_ai_sdk"
	RuntimeNodeOpenAI Runtime = "node_openai_sdk"
	RuntimeRawHTTP    Runtime = "raw_http"
	RuntimeSubprocess Runtime = "subprocess"

	AuthBearer        AuthMode = "bearer"
	AuthAPIKeyHeader  AuthMode = "api_key_header"
	AuthXAPIKey       AuthMode = "x_api_key"
	AuthCustomHeaders AuthMode = "custom_headers"
	AuthNone          AuthMode = "none"

	BillingAPI          BillingMode = "api"
	BillingSubscription BillingMode = "subscription"
	BillingLocal        BillingMode = "local"
	BillingOAuth        BillingMode = "oauth_subscription"
	BillingUnknown      BillingMode = "unknown"

	CostFree      CostClass = "free"
	CostCheap     CostClass = "cheap"
	CostMedium    CostClass = "medium"
	CostExpensive CostClass = "expensive"
	CostUnknown   CostClass = "unknown"

	ReasoningOff    ReasoningLevel = "off"
	ReasoningLow    ReasoningLevel = "low"
	ReasoningMedium ReasoningLevel = "medium"
	ReasoningHigh   ReasoningLevel = "high"
	ReasoningMax    ReasoningLevel = "max"
)

type ModelProfile struct {
	ID                  string    `json:"id"`
	DisplayName         string    `json:"displayName,omitempty"`
	ContextWindow       int       `json:"contextWindow,omitempty"`
	MaxOutputTokens     int       `json:"maxOutputTokens,omitempty"`
	SupportsTools       bool      `json:"supportsTools"`
	SupportsVision      bool      `json:"supportsVision"`
	SupportsReasoning   bool      `json:"supportsReasoning"`
	SupportsPromptCache bool      `json:"supportsPromptCache"`
	SupportsStreaming   bool      `json:"supportsStreaming"`
	CostClass           CostClass `json:"costClass"`
}

type ProviderProfile struct {
	ID               string            `json:"id"`
	DisplayName      string            `json:"displayName"`
	Enabled          bool              `json:"enabled"`
	Protocol         Protocol          `json:"protocol"`
	Runtime          Runtime           `json:"runtime"`
	BillingMode      BillingMode       `json:"billingMode"`
	BaseURL          string            `json:"baseURL"`
	Endpoint         string            `json:"endpoint"`
	APIKeyEnv        string            `json:"apiKeyEnv,omitempty"`
	APIKeySecretRef  string            `json:"apiKeySecretRef,omitempty"`
	AuthMode         AuthMode          `json:"authMode"`
	AuthHeaderName   string            `json:"authHeaderName,omitempty"`
	CustomHeaders    map[string]string `json:"customHeaders"`
	DefaultModel     string            `json:"defaultModel"`
	Models           []ModelProfile    `json:"models"`
	ReasoningAdapter string            `json:"reasoningAdapter,omitempty"`
	StreamingParser  string            `json:"streamingParser"`
	TimeoutMs        int               `json:"timeoutMs"`
	MaxRetries       int               `json:"maxRetries"`
	Notes            string            `json:"notes,omitempty"`
}

type RouteRequest struct {
	TaskType          string    `json:"taskType"`
	NeedsReasoning   bool      `json:"needsReasoning"`
	NeedsTools       bool      `json:"needsTools"`
	NeedsLongContext bool      `json:"needsLongContext"`
	MaxCostClass     CostClass `json:"maxCostClass"`
	Preferred        []string  `json:"preferred"`
	Denied           []string  `json:"denied"`
}

type RouteResult struct {
	ProviderID string  `json:"providerId"`
	Model      string  `json:"model"`
	Runtime    Runtime `json:"runtime"`
	Reason     string  `json:"reason"`
}
