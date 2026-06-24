import { useEffect, useMemo, useState } from "react";

type DoctorReport = { status: string; checks: Array<{ id: string; status: string; message: string }> };
type UserPrompt = { content: string; path: string; updatedAt?: string };
type ModelOption = { providerId: string; modelId: string; label: string };
type TranscriptItem = { role: "user" | "harness"; text: string };
type Panel = "chat" | "providers" | "commands" | "diagnostics" | "prompts";

type ModelProfile = { id: string; displayName?: string; contextWindow?: number; maxOutputTokens?: number };
type ProviderProfile = {
  id: string;
  displayName: string;
  enabled: boolean;
  protocol: string;
  runtime: string;
  billingMode: string;
  baseURL: string;
  endpoint: string;
  apiKeyEnv?: string;
  apiKeySecretRef?: string;
  apiKeyFileHint?: string;
  authMode: string;
  authHeaderName?: string;
  defaultModel: string;
  models: ModelProfile[];
  notes?: string;
};
type ProviderRegistry = { version: number; providers: ProviderProfile[] };
type ProviderRegistryResponse = { registry: ProviderRegistry; secrets: Record<string, boolean> };

const panels: Array<{ id: Panel; label: string; description: string }> = [
  { id: "chat", label: "Session", description: "Prompt and send" },
  { id: "providers", label: "Providers", description: "Keys and endpoints" },
  { id: "commands", label: "Commands", description: "Goal and runtime" },
  { id: "diagnostics", label: "Diagnostics", description: "Doctor and probes" },
  { id: "prompts", label: "System prompt", description: "Persistent user rules" }
];

const api = {
  get: async <T,>(path: string): Promise<T> => {
    const response = await fetch(path);
    if (!response.ok) throw new Error(`${path} failed`);
    return response.json();
  },
  post: async <T,>(path: string, payload: Record<string, unknown>): Promise<T> => {
    const response = await fetch(path, { method: "POST", headers: { "content-type": "application/json" }, body: JSON.stringify(payload) });
    const body = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(typeof body.error === "string" ? body.error : `${path} failed`);
    return body as T;
  },
  put: async <T,>(path: string, payload: unknown): Promise<T> => {
    const response = await fetch(path, { method: "PUT", headers: { "content-type": "application/json" }, body: JSON.stringify(payload) });
    const body = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(typeof body.error === "string" ? body.error : `${path} failed`);
    return body as T;
  }
};

export function App() {
  const [panel, setPanel] = useState<Panel>("chat");
  const [doctor, setDoctor] = useState<DoctorReport | null>(null);
  const [registry, setRegistry] = useState<ProviderRegistry>({ version: 1, providers: [] });
  const [secrets, setSecrets] = useState<Record<string, boolean>>({});
  const [activeProviderId, setActiveProviderId] = useState("");
  const [apiKeyDraft, setApiKeyDraft] = useState("");
  const [prompt, setPrompt] = useState<UserPrompt>({ content: "", path: "" });
  const [promptDraft, setPromptDraft] = useState("");
  const [workspaceRoot, setWorkspaceRoot] = useState(".");
  const [sessionId, setSessionId] = useState(() => `web-${Date.now()}`);
  const [topic, setTopic] = useState("");
  const [goal, setGoal] = useState("");
  const [selectedModel, setSelectedModel] = useState("");
  const [sessionPrompt, setSessionPrompt] = useState("");
  const [transcript, setTranscript] = useState<TranscriptItem[]>([]);
  const [yolo, setYolo] = useState(false);
  const [allowBillingChange, setAllowBillingChange] = useState(false);
  const [message, setMessage] = useState("Loading Harnejr");

  useEffect(() => { void loadAll(); }, []);

  const modelOptions = useMemo(() => extractModelOptions(registry), [registry]);
  const activeProvider = useMemo(() => registry.providers.find((provider) => provider.id === activeProviderId) ?? registry.providers[0], [registry.providers, activeProviderId]);
  const selected = useMemo(() => parseSelectedModel(selectedModel), [selectedModel]);
  const failingChecks = useMemo(() => doctor?.checks.filter((check) => check.status !== "pass") ?? [], [doctor]);

  async function loadAll() {
    try {
      const [doctorReport, promptReport, providerReport] = await Promise.all([
        api.get<DoctorReport>("/api/doctor"),
        api.get<UserPrompt>("/api/prompts/user"),
        api.get<ProviderRegistryResponse>("/api/providers/registry")
      ]);
      setDoctor(doctorReport);
      setPrompt(promptReport);
      setPromptDraft(promptReport.content ?? "");
      setRegistry(providerReport.registry);
      setSecrets(providerReport.secrets ?? {});
      const firstProvider = providerReport.registry.providers[0];
      if (firstProvider) {
        setActiveProviderId(firstProvider.id);
        const firstModel = firstProvider.models?.[0]?.id ?? firstProvider.defaultModel;
        setSelectedModel(`${firstProvider.id}::${firstModel}`);
      }
      setMessage("Ready");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Unable to load Harnejr");
    }
  }

  function add(role: TranscriptItem["role"], text: string) { setTranscript((items) => [...items, { role, text }]); }

  function updateProvider(patch: Partial<ProviderProfile>) {
    if (!activeProvider) return;
    const providers = registry.providers.map((provider) => provider.id === activeProvider.id ? { ...provider, ...patch } : provider);
    setRegistry({ ...registry, providers });
  }

  async function saveProviderRegistry() {
    setMessage("Saving provider configuration");
    try {
      const result = await api.put<ProviderRegistryResponse>("/api/providers/registry", registry);
      setRegistry(result.registry);
      setSecrets(result.secrets ?? {});
      setMessage("Provider configuration saved");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Provider save failed");
    }
  }

  async function saveProviderSecret() {
    if (!activeProvider || apiKeyDraft.trim() === "") { setMessage("Enter an API key first"); return; }
    setMessage("Saving API key locally");
    try {
      await api.put("/api/providers/secret", { providerId: activeProvider.id, apiKey: apiKeyDraft });
      setApiKeyDraft("");
      await loadProviderRegistryOnly();
      setMessage("API key saved to local secret file");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "API key save failed");
    }
  }

  async function loadProviderRegistryOnly() {
    const result = await api.get<ProviderRegistryResponse>("/api/providers/registry");
    setRegistry(result.registry);
    setSecrets(result.secrets ?? {});
  }

  async function savePrompt() {
    setMessage("Saving user prompt");
    try {
      const saved = await api.put<UserPrompt>("/api/prompts/user", { content: promptDraft });
      setPrompt(saved);
      setPromptDraft(saved.content ?? "");
      setMessage("User-level system prompt saved");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Unable to save user prompt"); }
  }

  async function applyControls() {
    setMessage("Applying session controls");
    try {
      await api.post("/api/control/apply", { workspaceRoot, sessionId, topic, goal, yolo });
      setMessage("Session controls saved");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Unable to apply controls"); }
  }

  async function startCheckpointedGoal() {
    if (!goal.trim()) { setMessage("Enter a goal first"); return; }
    setMessage("Starting checkpointed goal");
    try {
      const result = await api.post<Record<string, unknown>>("/api/goals/start", { workspaceRoot, sessionId, goal });
      add("harness", JSON.stringify(result, null, 2));
      setMessage("Checkpointed goal started");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Goal start failed"); }
  }

  async function summarizeMemory() {
    setMessage("Building compact memory");
    try {
      const result = await api.post<Record<string, unknown>>("/api/memory/summary", { workspaceRoot, sessionId });
      add("harness", JSON.stringify(result, null, 2));
      setMessage("Compact memory updated");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Memory summary failed"); }
  }

  async function sendMessage() {
    const text = sessionPrompt.trim();
    if (!text) { setMessage("Enter a message first"); return; }
    add("user", text);
    setSessionPrompt("");
    setMessage("Calling provider");
    try {
      const result = await api.post<Record<string, unknown>>("/api/llm/generate", { providerId: selected.providerId, model: selected.modelId, prompt: text, maxTokens: 2048, allowBillingChange });
      add("harness", typeof result.text === "string" && result.text ? result.text : JSON.stringify(result, null, 2));
      setMessage("Provider call finished");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Provider call failed"); }
  }

  async function saveToMemory() {
    const text = sessionPrompt.trim();
    if (!text) { setMessage("Enter text first"); return; }
    setMessage("Saving to memory");
    try {
      const result = await api.post<{ message: string }>("/api/session/message", { workspaceRoot, sessionId, providerId: selected.providerId, modelId: selected.modelId, prompt: text });
      add("harness", result.message);
      setMessage("Saved to workspace memory");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Memory save failed"); }
  }

  async function runWorkers() {
    const task = sessionPrompt.trim() || goal.trim();
    if (!task) { setMessage("Enter a task or goal first"); return; }
    setMessage("Running workers");
    try {
      const result = await api.post<Record<string, unknown>>("/api/workers/run", { workspaceRoot, sessionId, task, mode: goal ? "goal" : "task", providerId: selected.providerId, model: selected.modelId, allowBillingChange });
      add("harness", JSON.stringify(result, null, 2));
      setMessage("Workers finished");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Worker run failed"); }
  }

  async function runReview() {
    setMessage("Running review");
    try {
      const result = await api.post<Record<string, unknown>>("/api/review/run", { providerId: selected.providerId, model: selected.modelId, input: { goal: goal || sessionPrompt, evidence: transcript.map((item) => item.text).slice(-5), tests: [], subagentReviews: 2, qualityGatePass: false, providerPlanPass: false } });
      add("harness", JSON.stringify(result, null, 2));
      setMessage("Review finished");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Review failed"); }
  }

  async function runDiagnostic(path: string, label: string) {
    setMessage(`Checking ${label}`);
    try {
      const result = await api.get<Record<string, unknown>>(path);
      add("harness", JSON.stringify(result, null, 2));
      setMessage(`${label} checked`);
    } catch (error) { setMessage(error instanceof Error ? error.message : `${label} check failed`); }
  }

  return (
    <main className="appShell">
      <aside className="sidebar">
        <div className="brandBlock"><h1>Harnejr</h1><p>Local coding harness</p></div>
        <nav className="navList">{panels.map((item) => <button key={item.id} className={panel === item.id ? "navItem active" : "navItem"} onClick={() => setPanel(item.id)}><span>{item.label}</span><small>{item.description}</small></button>)}</nav>
        <div className="sideStatus"><span>{message}</span><small>Doctor: {doctor?.status ?? "unknown"}</small></div>
      </aside>

      <section className="mainStage">
        {panel === "chat" ? <ChatPanel modelOptions={modelOptions} selectedModel={selectedModel} setSelectedModel={setSelectedModel} transcript={transcript} sessionPrompt={sessionPrompt} setSessionPrompt={setSessionPrompt} sendMessage={sendMessage} saveToMemory={saveToMemory} allowBillingChange={allowBillingChange} setAllowBillingChange={setAllowBillingChange} /> : null}
        {panel === "providers" ? <ProviderPanel registry={registry} activeProvider={activeProvider} activeProviderId={activeProviderId} setActiveProviderId={setActiveProviderId} updateProvider={updateProvider} saveProviderRegistry={saveProviderRegistry} saveProviderSecret={saveProviderSecret} apiKeyDraft={apiKeyDraft} setApiKeyDraft={setApiKeyDraft} secrets={secrets} /> : null}
        {panel === "commands" ? <CommandsPanel workspaceRoot={workspaceRoot} setWorkspaceRoot={setWorkspaceRoot} sessionId={sessionId} setSessionId={setSessionId} topic={topic} setTopic={setTopic} goal={goal} setGoal={setGoal} yolo={yolo} setYolo={setYolo} applyControls={applyControls} startCheckpointedGoal={startCheckpointedGoal} summarizeMemory={summarizeMemory} runWorkers={runWorkers} runReview={runReview} /> : null}
        {panel === "diagnostics" ? <DiagnosticsPanel doctor={doctor} failingChecks={failingChecks} runDiagnostic={runDiagnostic} /> : null}
        {panel === "prompts" ? <PromptsPanel prompt={prompt} promptDraft={promptDraft} setPromptDraft={setPromptDraft} savePrompt={savePrompt} /> : null}
      </section>
    </main>
  );
}

function ChatPanel(props: { modelOptions: ModelOption[]; selectedModel: string; setSelectedModel: (v: string) => void; transcript: TranscriptItem[]; sessionPrompt: string; setSessionPrompt: (v: string) => void; sendMessage: () => void; saveToMemory: () => void; allowBillingChange: boolean; setAllowBillingChange: (v: boolean) => void }) {
  return <section className="screen chatScreen"><header className="screenHeader"><div><p className="eyebrow">Session</p><h2>Message Harnejr</h2></div><label className="compactField">Model<select value={props.selectedModel} onChange={(event) => props.setSelectedModel(event.target.value)}>{props.modelOptions.map((option) => <option key={`${option.providerId}:${option.modelId}`} value={`${option.providerId}::${option.modelId}`}>{option.label}</option>)}</select></label></header><div className="conversation">{props.transcript.length === 0 ? <div className="emptyState"><h3>No conversation yet</h3><p>Configure providers first, then send a task here. Runtime tools live outside the composer.</p></div> : props.transcript.map((item, index) => <article className={item.role === "user" ? "bubble userBubble" : "bubble harnessBubble"} key={`${item.role}-${index}`}><strong>{item.role === "user" ? "You" : "Harnejr"}</strong><p>{item.text}</p></article>)}</div><footer className="composer"><textarea value={props.sessionPrompt} onChange={(event) => props.setSessionPrompt(event.target.value)} placeholder="Describe the coding task, inspection, patch, or review." /><div className="composerActions"><label className="checkRow"><input type="checkbox" checked={props.allowBillingChange} onChange={(event) => props.setAllowBillingChange(event.target.checked)} />Allow billing-mode fallback</label><button className="secondaryButton" type="button" onClick={props.saveToMemory}>Save to memory</button><button className="primaryButton" type="button" onClick={props.sendMessage}>Send</button></div></footer></section>;
}

function ProviderPanel(props: { registry: ProviderRegistry; activeProvider?: ProviderProfile; activeProviderId: string; setActiveProviderId: (v: string) => void; updateProvider: (patch: Partial<ProviderProfile>) => void; saveProviderRegistry: () => void; saveProviderSecret: () => void; apiKeyDraft: string; setApiKeyDraft: (v: string) => void; secrets: Record<string, boolean> }) {
  const provider = props.activeProvider;
  return <section className="screen providerScreen"><header className="screenHeader"><div><p className="eyebrow">Providers</p><h2>Models, endpoints, and local API keys</h2><p>Keys are saved to local files by the daemon. They are not stored in browser localStorage.</p></div><button className="primaryButton" onClick={props.saveProviderRegistry}>Save provider config</button></header><div className="providerLayout"><aside className="providerList">{props.registry.providers.map((item) => <button key={item.id} className={item.id === props.activeProviderId ? "providerItem active" : "providerItem"} onClick={() => props.setActiveProviderId(item.id)}><strong>{item.displayName}</strong><span>{item.id}</span><small>{props.secrets[item.id] ? "key saved" : "no local key"}</small></button>)}</aside>{provider ? <div className="providerEditor"><div className="formGrid twoCols"><label>Display name<input value={provider.displayName} onChange={(event) => props.updateProvider({ displayName: event.target.value })} /></label><label>Enabled<select value={provider.enabled ? "true" : "false"} onChange={(event) => props.updateProvider({ enabled: event.target.value === "true" })}><option value="true">Enabled</option><option value="false">Disabled</option></select></label><label>Base URL<input value={provider.baseURL} onChange={(event) => props.updateProvider({ baseURL: event.target.value })} /></label><label>Endpoint<input value={provider.endpoint} onChange={(event) => props.updateProvider({ endpoint: event.target.value })} /></label><label>Protocol<input value={provider.protocol} onChange={(event) => props.updateProvider({ protocol: event.target.value })} /></label><label>Billing mode<input value={provider.billingMode} onChange={(event) => props.updateProvider({ billingMode: event.target.value })} /></label><label>Auth mode<input value={provider.authMode} onChange={(event) => props.updateProvider({ authMode: event.target.value })} /></label><label>Auth header<input value={provider.authHeaderName ?? ""} onChange={(event) => props.updateProvider({ authHeaderName: event.target.value })} /></label><label>Default model<input value={provider.defaultModel} onChange={(event) => props.updateProvider({ defaultModel: event.target.value })} /></label><label>Env var<input value={provider.apiKeyEnv ?? ""} onChange={(event) => props.updateProvider({ apiKeyEnv: event.target.value })} /></label></div><div className="keyBox"><div><h3>API key</h3><p>{props.secrets[provider.id] ? "A local key is saved for this provider." : "No local key saved for this provider."}</p></div><input type="password" value={props.apiKeyDraft} onChange={(event) => props.setApiKeyDraft(event.target.value)} placeholder="Paste API key. It will be written to the daemon config secrets folder." /><button className="primaryButton" onClick={props.saveProviderSecret}>Save API key</button></div><div className="modelList"><h3>Models</h3>{provider.models.map((model) => <div className="modelRow" key={model.id}><strong>{model.displayName || model.id}</strong><span>{model.id}</span><small>{model.contextWindow ? `${model.contextWindow} ctx` : "context unknown"} {model.maxOutputTokens ? `· ${model.maxOutputTokens} out` : ""}</small></div>)}</div></div> : null}</div></section>;
}

function CommandsPanel(props: { workspaceRoot: string; setWorkspaceRoot: (v: string) => void; sessionId: string; setSessionId: (v: string) => void; topic: string; setTopic: (v: string) => void; goal: string; setGoal: (v: string) => void; yolo: boolean; setYolo: (v: boolean) => void; applyControls: () => void; startCheckpointedGoal: () => void; summarizeMemory: () => void; runWorkers: () => void; runReview: () => void }) {
  return <section className="screen"><header className="screenHeader"><div><p className="eyebrow">Commands</p><h2>Goal and runtime controls</h2><p>Controls are separate from the message composer so sending text remains obvious.</p></div></header><div className="formGrid twoCols"><label>Workspace root<input value={props.workspaceRoot} onChange={(event) => props.setWorkspaceRoot(event.target.value)} /></label><label>Session ID<input value={props.sessionId} onChange={(event) => props.setSessionId(event.target.value)} /></label><label>Topic<input value={props.topic} onChange={(event) => props.setTopic(event.target.value)} /></label><label className="checkRow inlineCheck"><input type="checkbox" checked={props.yolo} onChange={(event) => props.setYolo(event.target.checked)} />Yolo for safe workspace work</label><label className="wideField">Goal<textarea value={props.goal} onChange={(event) => props.setGoal(event.target.value)} rows={6} placeholder="Define a goal that will be converted into checkpoints." /></label></div><div className="actionDeck"><button className="primaryButton" onClick={props.startCheckpointedGoal}>Start checkpointed goal</button><button className="secondaryButton" onClick={props.applyControls}>Apply controls</button><button className="secondaryButton" onClick={props.summarizeMemory}>Compact memory</button><button className="secondaryButton" onClick={props.runWorkers}>Run workers</button><button className="secondaryButton" onClick={props.runReview}>Run review</button></div></section>;
}

function DiagnosticsPanel(props: { doctor: DoctorReport | null; failingChecks: Array<{ id: string; status: string; message: string }>; runDiagnostic: (path: string, label: string) => void }) {
  return <section className="screen"><header className="screenHeader"><div><p className="eyebrow">Diagnostics</p><h2>Runtime health</h2><p>Use these to prove runtime state instead of trusting labels.</p></div></header><div className="diagnosticGrid"><article className="statusCard"><h3>Doctor</h3><p>{props.doctor?.status ?? "unknown"}</p><small>{props.failingChecks.length} failing checks</small></article><button className="secondaryButton" onClick={() => props.runDiagnostic("/api/doctor", "Doctor")}>Run Doctor</button><button className="secondaryButton" onClick={() => props.runDiagnostic("/api/providers/probe", "providers")}>Provider probe</button><button className="secondaryButton" onClick={() => props.runDiagnostic("/api/mcp/check", "MCP")}>MCP check</button></div><div className="checkList">{props.doctor?.checks.map((check) => <div key={check.id} className="checkItem"><strong>{check.id}</strong><span>{check.status}</span><p>{check.message}</p></div>)}</div></section>;
}

function PromptsPanel(props: { prompt: UserPrompt; promptDraft: string; setPromptDraft: (v: string) => void; savePrompt: () => void }) {
  return <section className="screen"><header className="screenHeader"><div><p className="eyebrow">System prompt</p><h2>User-level additive prompt</h2><p>Saved by the daemon and added to Harnejr's core prompt. It cannot override daemon safety policy.</p><small>{props.prompt.path || "not created yet"}</small></div><button className="primaryButton" onClick={props.savePrompt}>Save prompt</button></header><textarea className="largeEditor" value={props.promptDraft} onChange={(event) => props.setPromptDraft(event.target.value)} placeholder="Add your personal operating rules, coding standards, and review expectations." /></section>;
}

function extractModelOptions(registry: ProviderRegistry): ModelOption[] {
  return registry.providers.flatMap((provider) => {
    const models = Array.isArray(provider.models) && provider.models.length > 0 ? provider.models : [{ id: provider.defaultModel }];
    return models.map((model) => ({ providerId: provider.id, modelId: model.id, label: `${provider.displayName} / ${model.displayName ?? model.id}` }));
  });
}

function parseSelectedModel(value: string): { providerId: string; modelId: string } {
  const [providerId = "", modelId = ""] = value.split("::");
  return { providerId, modelId };
}
