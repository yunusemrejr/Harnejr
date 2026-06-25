import { readFile } from "node:fs/promises";
import { generateText } from "ai";
import { createAnthropic } from "@ai-sdk/anthropic";
import { createOpenAICompatible } from "@ai-sdk/openai-compatible";
import type { ProviderProfile } from "@harnejr/shared";
import { z } from "zod";

const UnknownRecord = z.record(z.string(), z.unknown());

const ProviderSchema = z.object({
  id: z.string(),
  aliases: z.array(z.string()).default([]),
  openCodeProviderId: z.string().optional(),
  displayName: z.string(),
  enabled: z.boolean().default(true),
  protocol: z.enum(["openai_chat", "openai_responses", "anthropic_messages", "ollama_native", "cli_backed", "oauth_backed", "custom_http"]),
  runtime: z.enum(["go_native", "node_ai_sdk", "node_openai_sdk", "raw_http", "subprocess"]),
  billingMode: z.enum(["api", "subscription", "oauth_subscription", "local", "unknown"]),
  baseURL: z.string(),
  endpoint: z.string(),
  apiKeyEnv: z.string().optional(),
  apiKeySecretRef: z.string().optional(),
  apiKeyFileHint: z.string().optional(),
  authMode: z.enum(["bearer", "api_key_header", "x_api_key", "custom_headers", "none"]).default("bearer"),
  authHeaderName: z.string().optional(),
  customHeaders: z.record(z.string(), z.string()).default({}),
  defaultModel: z.string(),
  models: z.array(z.object({
    id: z.string(),
    displayName: z.string().optional(),
    contextWindow: z.number().optional(),
    maxOutputTokens: z.number().optional(),
    supportsTools: z.boolean().default(false),
    supportsVision: z.boolean().default(false),
    supportsReasoning: z.boolean().default(false),
    supportsPromptCache: z.boolean().default(false),
    supportsStreaming: z.boolean().default(true),
    costClass: z.enum(["free", "cheap", "medium", "expensive", "unknown"]).default("unknown")
  })).default([]),
  requestDefaults: UnknownRecord.default({}),
  payloadOverrides: UnknownRecord.default({}),
  extraBody: UnknownRecord.default({}),
  modelRequestDefaults: z.record(z.string(), UnknownRecord).default({}),
  reasoningAdapter: z.string().optional(),
  streamingParser: z.string().default("openai_sse"),
  timeoutMs: z.number().default(120000),
  maxRetries: z.number().default(2),
  notes: z.string().optional()
});

const GeneratePayload = z.object({
  provider: ProviderSchema,
  model: z.string().optional(),
  request: z.object({
    providerId: z.string().optional(),
    model: z.string().optional(),
    system: z.string().optional(),
    prompt: z.string(),
    maxTokens: z.number().int().positive().optional(),
    allowBillingChange: z.boolean().optional()
  })
});

export type SmokeResult = {
  providerId: string;
  ok: boolean;
  reason: string;
};

export function sdkInventory() {
  return {
    runtimes: ["node_ai_sdk"],
    sdks: ["ai", "@ai-sdk/openai-compatible", "@ai-sdk/anthropic", "zod"],
    protocols: ["openai_chat", "anthropic_messages", "custom_http"]
  };
}

export async function smokeTestProvider(profile: ProviderProfile): Promise<SmokeResult> {
  if (!profile.enabled) {
    return { providerId: profile.id, ok: false, reason: "provider is disabled" };
  }
  if (profile.authMode !== "none" && !profile.apiKeyEnv && !profile.apiKeySecretRef) {
    return { providerId: profile.id, ok: false, reason: "provider has no apiKeyEnv or apiKeySecretRef" };
  }
  return { providerId: profile.id, ok: true, reason: "static profile validation passed" };
}

export async function generateWithSDK(raw: unknown) {
  const payload = GeneratePayload.parse(raw);
  const provider = payload.provider as ProviderProfile;
  const model = payload.model || payload.request.model || provider.defaultModel;
  const key = await resolveKey(provider);
  if (provider.authMode !== "none" && !key) {
    return errorResult(provider.id, model, "auth", `missing auth for ${provider.apiKeyEnv || provider.apiKeySecretRef || provider.id}`);
  }
  try {
    const maxOutputTokens = payload.request.maxTokens || 1024;
    const messages = buildMessages(payload.request);
    const defaults = flatDefaults(provider, model);
    if (provider.protocol === "anthropic_messages") {
      const anthropic = createAnthropic({ apiKey: key, baseURL: provider.baseURL });
      const result = await generateText({ model: anthropic(model), messages: messages as never, maxOutputTokens, ...defaults });
      return okResult(provider.id, model, result.text);
    }
    const compatible = createOpenAICompatible({ name: provider.id, apiKey: key, baseURL: provider.baseURL });
    const result = await generateText({ model: compatible(model), messages: messages as never, maxOutputTokens, ...defaults });
    return okResult(provider.id, model, result.text);
  } catch (error) {
    return errorResult(provider.id, model, "sdk-ai", error instanceof Error ? error.message : String(error));
  }
}

function buildMessages(request: z.infer<typeof GeneratePayload>["request"]) {
  const messages: Array<{ role: "system" | "user"; content: string }> = [];
  if (request.system) messages.push({ role: "system", content: request.system });
  messages.push({ role: "user", content: request.prompt });
  return messages;
}

function flatDefaults(provider: ProviderProfile, model: string) {
  return {
    ...provider.requestDefaults,
    ...provider.extraBody,
    ...(provider.modelRequestDefaults[model] || {}),
    ...provider.payloadOverrides
  };
}

async function resolveKey(provider: ProviderProfile) {
  if (provider.apiKeySecretRef) {
    try {
      const value = (await readFile(provider.apiKeySecretRef, "utf8")).trim();
      if (value) return value;
    } catch {
      // fall through to environment lookup
    }
  }
  return provider.apiKeyEnv ? process.env[provider.apiKeyEnv] || "" : "";
}

function okResult(providerId: string, model: string, text: string) {
  return { providerId, model, text };
}

function errorResult(providerId: string, model: string, errorClass: string, error: string) {
  return { providerId, model, text: "", errorClass, error };
}

async function readStdin() {
  const chunks: Buffer[] = [];
  for await (const chunk of process.stdin) {
    chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
  }
  return Buffer.concat(chunks).toString("utf8");
}

async function main() {
  const command = process.argv[2] || "inventory";
  if (command === "inventory") {
    console.log(JSON.stringify(sdkInventory()));
    return;
  }
  if (command === "generate") {
    const input = await readStdin();
    const result = await generateWithSDK(JSON.parse(input));
    console.log(JSON.stringify(result));
    return;
  }
  throw new Error(`unknown provider-node command: ${command}`);
}

if (process.argv[1] && import.meta.url.endsWith(process.argv[1].replace(/\\/g, "/"))) {
  main().catch((error) => {
    console.error(error instanceof Error ? error.message : String(error));
    process.exit(1);
  });
}
