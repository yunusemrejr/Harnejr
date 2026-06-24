import { z } from "zod";

export const ProviderProtocol = z.enum([
  "openai_chat",
  "openai_responses",
  "anthropic_messages",
  "ollama_native",
  "cli_backed",
  "oauth_backed",
  "custom_http"
]);

export const ProviderRuntime = z.enum([
  "go_native",
  "node_ai_sdk",
  "node_openai_sdk",
  "raw_http",
  "subprocess"
]);

export const AuthMode = z.enum([
  "bearer",
  "api_key_header",
  "x_api_key",
  "custom_headers",
  "none"
]);

export const BillingMode = z.enum([
  "api",
  "subscription",
  "oauth_subscription",
  "local",
  "unknown"
]);

export const CostClass = z.enum(["free", "cheap", "medium", "expensive", "unknown"]);

export const ModelProfile = z.object({
  id: z.string().min(1),
  displayName: z.string().optional(),
  contextWindow: z.number().int().positive().optional(),
  maxOutputTokens: z.number().int().positive().optional(),
  supportsTools: z.boolean().default(false),
  supportsVision: z.boolean().default(false),
  supportsReasoning: z.boolean().default(false),
  supportsPromptCache: z.boolean().default(false),
  supportsStreaming: z.boolean().default(true),
  costClass: CostClass.default("unknown")
});

export const ProviderProfile = z.object({
  id: z.string().min(1),
  displayName: z.string().min(1),
  enabled: z.boolean().default(true),
  protocol: ProviderProtocol,
  runtime: ProviderRuntime,
  billingMode: BillingMode,
  baseURL: z.string().min(1),
  endpoint: z.string().min(1),
  apiKeyEnv: z.string().optional(),
  apiKeySecretRef: z.string().optional(),
  authMode: AuthMode.default("bearer"),
  authHeaderName: z.string().optional(),
  customHeaders: z.record(z.string(), z.string()).default({}),
  defaultModel: z.string().min(1),
  models: z.array(ModelProfile).default([]),
  requestDefaults: z.record(z.string(), z.unknown()).default({}),
  payloadOverrides: z.record(z.string(), z.unknown()).default({}),
  extraBody: z.record(z.string(), z.unknown()).default({}),
  reasoningAdapter: z.string().optional(),
  streamingParser: z.string().default("openai_sse"),
  timeoutMs: z.number().int().positive().default(120000),
  maxRetries: z.number().int().min(0).default(2),
  notes: z.string().optional()
});

export const ProviderRegistry = z.object({
  version: z.number().int().positive(),
  providers: z.array(ProviderProfile)
});

export type ProviderProfile = z.infer<typeof ProviderProfile>;
export type ProviderRegistry = z.infer<typeof ProviderRegistry>;
