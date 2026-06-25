import { useEffect, useMemo, useState } from "react";

type DoctorReport = { status: string; checks: Array<{ id: string; status: string; message: string }> };
type UserPrompt = { content: string; path: string; updatedAt?: string };
type TranscriptItem = { role: "user" | "harness"; text: string };
type Panel = "session" | "providers" | "runtime" | "diagnostics" | "prompts";
type ModelProfile = { id: string; displayName?: string; contextWindow?: number; maxOutputTokens?: number };
type ProviderProfile = {
  id: string; displayName: string; enabled: boolean; protocol: string; runtime: string; billingMode: string;
  baseURL: string; endpoint: string; apiKeyEnv?: string; apiKeySecretRef?: string; apiKeyFileHint?: string;
  authMode: string; authHeaderName?: string; defaultModel: string; models: ModelProfile[]; notes?: string;
};
type ProviderRegistry = { version: number; providers: ProviderProfile[] };
type ProviderRegistryResponse = { registry: ProviderRegistry; secrets: Record<string, boolean> };

const panels: Array<{ id: Panel; label: string; detail: string }> = [
  { id: "session", label: "Session", detail: "Task and transcript" },
  { id: "providers", label: "Providers", detail: "Keys, models, endpoints" },
  { id: "runtime", label: "Runtime", detail: "Workspace, goals, workers" },
  { id: "diagnostics", label: "Diagnostics", detail: "Doctor, MCP, skills" },
  { id: "prompts", label: "Prompts", detail: "System rules" }
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
  const [panel, setPanel] = useState<Panel>("session");
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

  const activeProvider = useMemo(() => registry.providers.find((p) => p.id === activeProviderId) ?? registry.providers[0], [registry.providers, activeProviderId]);
  const modelOptions = useMemo(() => registry.providers.flatMap((p) => modelsOf(p).map((m) => ({ providerId: p.id, modelId: m.id, label: `${p.displayName} / ${m.displayName ?? m.id}` }))), [registry]);
  const selected = useMemo(() => parseSelectedModel(selectedModel), [selectedModel]);
  const selectedProvider = registry.providers.find((p) => p.id === selected.providerId);
  const selectedRoute = modelsOf(selectedProvider).find((m) => m.id === selected.modelId);
  const savedKeyCount = Object.values(secrets).filter(Boolean).length;
  const failingChecks = doctor?.checks.filter((check) => check.status !== "pass") ?? [];

  async function loadAll() {
    try {
      const [doctorReport, promptReport, providerReport] = await Promise.all([
        api.get<DoctorReport>("/api/doctor"), api.get<UserPrompt>("/api/prompts/user"), api.get<ProviderRegistryResponse>("/api/providers/registry")
      ]);
      setDoctor(doctorReport); setPrompt(promptReport); setPromptDraft(promptReport.content ?? "");
      setRegistry(providerReport.registry); setSecrets(providerReport.secrets ?? {});
      const first = providerReport.registry.providers[0];
      if (first) { setActiveProviderId(first.id); setSelectedModel(`${first.id}::${first.defaultModel || modelsOf(first)[0]?.id || ""}`); }
      setMessage("Ready");
    } catch (error) { setMessage(error instanceof Error ? error.message : "Unable to load Harnejr"); }
  }

  function add(role: TranscriptItem["role"], text: string) { setTranscript((items) => [...items, { role, text }]); }
  function chooseProvider(id: string) {
    const provider = registry.providers.find((p) => p.id === id); setActiveProviderId(id);
    if (provider) setSelectedModel(`${provider.id}::${provider.defaultModel || modelsOf(provider)[0]?.id || ""}`);
  }
  function updateProvider(patch: Partial<ProviderProfile>) {
    if (!activeProvider) return;
    const before = activeProvider.id; const next = { ...activeProvider, ...patch };
    setRegistry((current) => ({ ...current, providers: current.providers.map((p) => (p.id === before ? next : p)) }));
    if (patch.id && patch.id !== before) { setActiveProviderId(patch.id); setSelectedModel(`${patch.id}::${parseSelectedModel(selectedModel).modelId}`); }
  }
  function addProvider() {
    const id = uniqueId("custom-openai-compatible", registry.providers.map((p) => p.id));
    const provider: ProviderProfile = { id, displayName: "Custom OpenAI-compatible provider", enabled: true, protocol: "openai_chat", runtime: "raw_http", billingMode: "manual", baseURL: "https://example.com/v1", endpoint: "/chat/completions", apiKeyEnv: "CUSTOM_LLM_API_KEY", authMode: "bearer", authHeaderName: "Authorization", defaultModel: "model-id", models: [{ id: "model-id", displayName: "model-id", contextWindow: 128000, maxOutputTokens: 8192 }], notes: "Replace this template with the exact provider contract." };
    setRegistry((current) => ({ ...current, providers: [provider, ...current.providers] })); setActiveProviderId(id); setSelectedModel(`${id}::model-id`); setMessage("Provider added. Save config to persist it.");
  }
  function duplicateProvider() {
    if (!activeProvider) return;
    const id = uniqueId(`${activeProvider.id}-copy`, registry.providers.map((p) => p.id)); const copy = { ...activeProvider, id, displayName: `${activeProvider.displayName} copy` };
    setRegistry((current) => ({ ...current, providers: [copy, ...current.providers] })); setActiveProviderId(id); setSelectedModel(`${id}::${copy.defaultModel}`); setMessage("Provider duplicated. Save config to persist it.");
  }
  function removeProvider() {
    if (!activeProvider || registry.providers.length <= 1) { setMessage("At least one provider must remain"); return; }
    const providers = registry.providers.filter((p) => p.id !== activeProvider.id); const first = providers[0];
    setRegistry((current) => ({ ...current, providers })); setActiveProviderId(first?.id ?? ""); setSelectedModel(first ? `${first.id}::${first.defaultModel || modelsOf(first)[0]?.id || ""}` : ""); setMessage("Provider removed. Save config to persist it.");
  }
  function addModel() { if (activeProvider) updateProvider({ models: [...modelsOf(activeProvider), { id: uniqueId("model-id", modelsOf(activeProvider).map((m) => m.id)), contextWindow: 128000, maxOutputTokens: 8192 }] }); }
  function updateModel(oldId: string, patch: Partial<ModelProfile>) { if (activeProvider) updateProvider({ models: modelsOf(activeProvider).map((m) => (m.id === oldId ? { ...m, ...patch } : m)) }); }
  function removeModel(id: string) {
    if (!activeProvider || modelsOf(activeProvider).length <= 1) { setMessage("At least one model must remain"); return; }
    const models = modelsOf(activeProvider).filter((m) => m.id !== id); updateProvider({ models, defaultModel: activeProvider.defaultModel === id ? models[0]?.id ?? "" : activeProvider.defaultModel });
  }

  async function saveProviderRegistry() { await act("Saving provider configuration", "Provider configuration saved", "Provider save failed", async () => { const r = await api.put<ProviderRegistryResponse>("/api/providers/registry", registry); setRegistry(r.registry); setSecrets(r.secrets ?? {}); }); }
  async function saveProviderSecret() {
    if (!activeProvider || apiKeyDraft.trim() === "") { setMessage("Enter an API key first"); return; }
    await act("Saving API key locally", "API key saved to local secret file", "API key save failed", async () => { await api.put("/api/providers/secret", { providerId: activeProvider.id, apiKey: apiKeyDraft }); setApiKeyDraft(""); const r = await api.get<ProviderRegistryResponse>("/api/providers/registry"); setRegistry(r.registry); setSecrets(r.secrets ?? {}); });
  }
  async function savePrompt() { await act("Saving user prompt", "User-level system prompt saved", "Unable to save user prompt", async () => { const saved = await api.put<UserPrompt>("/api/prompts/user", { content: promptDraft }); setPrompt(saved); setPromptDraft(saved.content ?? ""); }); }
  async function applyControls() { await act("Applying session controls", "Session controls saved", "Unable to apply controls", async () => { await api.post("/api/control/apply", { workspaceRoot, sessionId, topic, goal, yolo }); }); }
  async function prepareWorkspace() { await postReport("Preparing workspace", "Workspace prepared", "Workspace preparation failed", "/api/workspaces/prepare", { workspaceRoot, sessionId, userRequest: topic || goal || "Harnejr web session" }); }
  async function startCheckpointedGoal() { if (!goal.trim()) { setMessage("Enter a goal first"); return; } await postReport("Starting checkpointed goal", "Checkpointed goal started", "Goal start failed", "/api/goals/start", { workspaceRoot, sessionId, goal }); }
  async function summarizeMemory() { await postReport("Building compact memory", "Compact memory updated", "Memory summary failed", "/api/memory/summary", { workspaceRoot, sessionId }); }
  async function runWorkers() { const task = sessionPrompt.trim() || goal.trim(); if (!task) { setMessage("Enter a task or goal first"); return; } await postReport("Running workers", "Workers finished", "Worker run failed", "/api/workers/run", { workspaceRoot, sessionId, task, mode: goal ? "goal" : "task", providerId: selected.providerId, model: selected.modelId, allowBillingChange }); }
  async function runReview() { await postReport("Running review", "Review finished", "Review failed", "/api/review/run", { providerId: selected.providerId, model: selected.modelId, input: { goal: goal || sessionPrompt, evidence: transcript.map((i) => i.text).slice(-5), tests: [], subagentReviews: 2, qualityGatePass: false, providerPlanPass: false } }); }
  async function probeActiveProvider() { if (activeProvider) await postReport("Probing selected provider", "Provider probe finished", "Provider probe failed", "/api/providers/probe", { providerId: activeProvider.id, model: activeProvider.defaultModel }); }
  async function discoverSkills() { await postReport("Discovering skills and agents", "Skills and agents discovered", "Skill discovery failed", "/api/skills/discover", { workspaceRoot }); }
  async function runDiagnostic(path: string, label: string) { await act(`Checking ${label}`, `${label} checked`, `${label} check failed`, async () => add("harness", JSON.stringify(await api.get<Record<string, unknown>>(path), null, 2))); }
  async function postReport(start: string, done: string, fail: string, path: string, payload: Record<string, unknown>) { await act(start, done, fail, async () => add("harness", JSON.stringify(await api.post<Record<string, unknown>>(path, payload), null, 2))); }
  async function act(start: string, done: string, fail: string, fn: () => Promise<void>) { setMessage(start); try { await fn(); setMessage(done); } catch (error) { setMessage(error instanceof Error ? error.message : fail); } }
  async function sendMessage() {
    const text = sessionPrompt.trim(); if (!text) { setMessage("Enter a message first"); return; }
    if (!selected.providerId || !selected.modelId) { setMessage("Select a provider model first"); return; }
    add("user", text); setSessionPrompt("");
    await act("Calling provider", "Provider call finished", "Provider call failed", async () => {
      const result = await api.post<Record<string, unknown>>("/api/llm/generate", { providerId: selected.providerId, model: selected.modelId, prompt: text, maxTokens: 2048, allowBillingChange });
      add("harness", typeof result.text === "string" && result.text ? result.text : JSON.stringify(result, null, 2));
    });
  }
  async function saveToMemory() {
    const text = sessionPrompt.trim(); if (!text) { setMessage("Enter text first"); return; }
    await act("Saving to memory", "Saved to workspace memory", "Memory save failed", async () => {
      const result = await api.post<{ message: string }>("/api/session/message", { workspaceRoot, sessionId, providerId: selected.providerId, modelId: selected.modelId, prompt: text }); add("harness", result.message);
    });
  }

  return (
    <main className="appFrame">
      <aside className="leftRail">
        <div className="productBlock"><span className="productKicker">Local web harness</span><h1>Harnejr</h1><p>Provider-aware agent control for Ubuntu workspaces.</p></div>
        <nav className="panelNav" aria-label="Harnejr sections">{panels.map((item) => <button key={item.id} className={panel === item.id ? "navButton selected" : "navButton"} onClick={() => setPanel(item.id)} type="button"><span>{item.label}</span><small>{item.detail}</small></button>)}</nav>
        <div className="railCard systemCard"><span className="railLabel">System state</span><strong>{message}</strong><small>Doctor: {doctor?.status ?? "unknown"}</small><small>{registry.providers.length} providers, {savedKeyCount} local keys</small></div>
      </aside>

      <section className="mainColumn">
        <header className="topBar"><div className="topContext"><span>Workspace</span><strong>{workspaceRoot}</strong></div><label className="topModel">Model<select value={selectedModel} onChange={(event) => setSelectedModel(event.target.value)}>{modelOptions.map((option) => <option key={`${option.providerId}:${option.modelId}`} value={`${option.providerId}::${option.modelId}`}>{option.label}</option>)}</select></label></header>
        <div className="panelCanvas">
          {panel === "session" ? renderSession() : null}
          {panel === "providers" ? renderProviders() : null}
          {panel === "runtime" ? renderRuntime() : null}
          {panel === "diagnostics" ? renderDiagnostics() : null}
          {panel === "prompts" ? renderPrompts() : null}
        </div>
      </section>

      <aside className="inspectorRail">
        <RailCard title="Selected route" facts={{ "Provider ID": selectedProvider?.id ?? "unknown", Model: selectedRoute?.displayName || selected.modelId || "unknown", Protocol: selectedProvider?.protocol ?? "unknown", Billing: selectedProvider?.billingMode ?? "unknown" }} />
        <RailCard title="Runtime scope" facts={{ Workspace: workspaceRoot, Session: sessionId, Yolo: yolo ? "enabled for safe actions" : "disabled", "Billing fallback": allowBillingChange ? "allowed for selected call" : "blocked" }} />
        <RailCard title="Config locations" facts={{ "Local keys": String(savedKeyCount), "User prompt": prompt.path || "not created", Goal: goal ? "defined" : "not set" }} />
      </aside>
    </main>
  );

  function renderSession() {
    return <section className="screen sessionScreen"><Intro label="Session" title="Work surface" text="Write the task once, send it through the selected provider, and keep runtime actions explicit." />
      <div className="conversationPane">{transcript.length === 0 ? <div className="quietState"><h3>No transcript yet</h3><p>Set provider keys, prepare a workspace, then send a task. Provider outputs and daemon reports stay here.</p></div> : transcript.map((item, index) => <article className={item.role === "user" ? "messageBubble userBubble" : "messageBubble harnessBubble"} key={`${item.role}-${index}`}><span>{item.role === "user" ? "User" : "Harnejr"}</span><p>{item.text}</p></article>)}</div>
      <footer className="composerPanel"><textarea value={sessionPrompt} onChange={(event) => setSessionPrompt(event.target.value)} placeholder="Ask Harnejr to inspect, patch, test, review, or plan against the current workspace." /><div className="composerFooter"><label className="checkLine"><input type="checkbox" checked={allowBillingChange} onChange={(event) => setAllowBillingChange(event.target.checked)} />Allow billing-mode fallback for this call</label><div className="buttonRow"><button className="secondaryButton" type="button" onClick={saveToMemory}>Save memory</button><button className="secondaryButton" type="button" onClick={runWorkers}>Workers</button><button className="secondaryButton" type="button" onClick={runReview}>Review</button><button className="primaryButton" type="button" onClick={sendMessage}>Send</button></div></div></footer>
    </section>;
  }

  function renderProviders() {
    const p = activeProvider;
    return <section className="screen providerScreen"><header className="screenIntro withActions"><div><span className="sectionLabel">Providers</span><h2>Transport contracts</h2><p>Providers are real endpoint, auth, billing, runtime, model namespace, and local-key contracts.</p></div><div className="buttonRow"><button className="secondaryButton" type="button" onClick={addProvider}>Add provider</button><button className="primaryButton" type="button" onClick={saveProviderRegistry}>Save config</button></div></header>
      <div className="providerWorkbench"><aside className="providerIndex">{registry.providers.map((item) => <button key={item.id} className={item.id === activeProviderId ? "providerTile selected" : "providerTile"} onClick={() => chooseProvider(item.id)} type="button"><strong>{item.displayName}</strong><span>{item.id}</span><small>{item.billingMode || "billing unknown"}</small><small>{secrets[item.id] ? "local key saved" : "local key missing"}</small></button>)}</aside>
        {p ? <div className="providerEditor"><div className="contractHeader"><div><span className="sectionLabel">Selected provider</span><h3>{p.displayName}</h3></div><div className="buttonRow"><button className="secondaryButton" type="button" onClick={probeActiveProvider}>Probe</button><button className="secondaryButton" type="button" onClick={duplicateProvider}>Duplicate</button><button className="secondaryButton" type="button" onClick={removeProvider}>Remove</button></div></div>
          <div className="editorSection"><h4>Identity</h4><div className="formGrid twoCols"><Field label="Provider ID" value={p.id} onChange={(v) => updateProvider({ id: v })} /><Field label="Display name" value={p.displayName} onChange={(v) => updateProvider({ displayName: v })} /><label>Enabled<select value={p.enabled ? "true" : "false"} onChange={(event) => updateProvider({ enabled: event.target.value === "true" })}><option value="true">Enabled</option><option value="false">Disabled</option></select></label><Field label="Billing mode" value={p.billingMode} onChange={(v) => updateProvider({ billingMode: v })} /></div></div>
          <div className="editorSection"><h4>Endpoint and call format</h4><div className="formGrid twoCols"><Field label="Base URL" value={p.baseURL} onChange={(v) => updateProvider({ baseURL: v })} /><Field label="Endpoint" value={p.endpoint} onChange={(v) => updateProvider({ endpoint: v })} /><Field label="Protocol" value={p.protocol} onChange={(v) => updateProvider({ protocol: v })} /><Field label="Runtime" value={p.runtime} onChange={(v) => updateProvider({ runtime: v })} /></div></div>
          <div className="editorSection"><h4>Authentication</h4><div className="formGrid twoCols"><Field label="Auth mode" value={p.authMode} onChange={(v) => updateProvider({ authMode: v })} /><Field label="Auth header" value={p.authHeaderName ?? ""} onChange={(v) => updateProvider({ authHeaderName: v })} /><Field label="Environment variable" value={p.apiKeyEnv ?? ""} onChange={(v) => updateProvider({ apiKeyEnv: v })} /><Field label="Secret file hint" value={p.apiKeyFileHint ?? ""} onChange={(v) => updateProvider({ apiKeyFileHint: v })} /></div><div className="keyVault"><div><strong>Local API key</strong><p>{secrets[p.id] ? "The daemon has a saved local key for this provider." : "No local key is saved for this provider."}</p></div><input type="password" value={apiKeyDraft} onChange={(event) => setApiKeyDraft(event.target.value)} placeholder="Paste API key. Browser storage is not used." /><button className="primaryButton" type="button" onClick={saveProviderSecret}>Save key</button></div></div>
          <div className="editorSection"><div className="sectionTitleRow"><h4>Models</h4><button className="secondaryButton" type="button" onClick={addModel}>Add model</button></div><Field label="Default model" value={p.defaultModel} onChange={(v) => updateProvider({ defaultModel: v })} /> <div className="modelEditorList">{modelsOf(p).map((m) => <div className="modelEditorRow" key={m.id}><Field label="Model ID" value={m.id} onChange={(v) => updateModel(m.id, { id: v })} /><Field label="Name" value={m.displayName ?? ""} onChange={(v) => updateModel(m.id, { displayName: v })} /><Field label="Context" value={m.contextWindow ?? ""} type="number" onChange={(v) => updateModel(m.id, { contextWindow: parseOptionalNumber(v) })} /><Field label="Output" value={m.maxOutputTokens ?? ""} type="number" onChange={(v) => updateModel(m.id, { maxOutputTokens: parseOptionalNumber(v) })} /><button className="secondaryButton" type="button" onClick={() => removeModel(m.id)}>Remove</button></div>)}</div></div>
          <div className="editorSection"><h4>Notes</h4><textarea value={p.notes ?? ""} onChange={(event) => updateProvider({ notes: event.target.value })} placeholder="Document endpoint quirks, subscription path, tool support, or model caveats." /></div>
        </div> : null}</div>
    </section>;
  }

  function renderRuntime() {
    return <section className="screen runtimeScreen"><Intro label="Runtime" title="Workspace and goal control" text="Session state, goal checkpoints, compact memory, workers, and review are separated from chat." />
      <div className="runtimeGrid"><div className="editorSection"><h4>Session scope</h4><div className="formGrid twoCols"><Field label="Workspace root" value={workspaceRoot} onChange={setWorkspaceRoot} /><Field label="Session ID" value={sessionId} onChange={setSessionId} /><Field label="Topic" value={topic} onChange={setTopic} wide /><label className="checkLine wideField"><input type="checkbox" checked={yolo} onChange={(event) => setYolo(event.target.checked)} />Yolo for safe workspace work</label></div></div><div className="editorSection goalForm"><h4>Checkpointed goal</h4><textarea value={goal} onChange={(event) => setGoal(event.target.value)} placeholder="Define the goal. The daemon turns this into deterministic checkpoints." /></div></div>
      <div className="actionMatrix"><Action title="Prepare workspace" text="Find or initialize the safe project root and create memory files." action="Prepare" onClick={prepareWorkspace} /><Action title="Apply controls" text="Persist workspace, session, topic, goal, and safe autonomy state." action="Apply" onClick={applyControls} /><Action title="Start goal" text="Create scope, plan, implement, verify, review, and complete checkpoints." action="Start" onClick={startCheckpointedGoal} /><Action title="Compact memory" text="Write a concise continuation summary without raw history bloat." action="Compact" onClick={summarizeMemory} /><Action title="Worker pass" text="Run provider-backed workers against the current task or goal." action="Run" onClick={runWorkers} /><Action title="Completion review" text="Challenge the current evidence through the review path." action="Review" onClick={runReview} /></div>
    </section>;
  }

  function renderDiagnostics() {
    return <section className="screen diagnosticsScreen"><header className="screenIntro withActions"><div><span className="sectionLabel">Diagnostics</span><h2>Runtime proof</h2><p>The daemon result is the source of truth. Interface labels are not proof.</p></div><div className="buttonRow"><button className="secondaryButton" type="button" onClick={() => runDiagnostic("/api/doctor", "Doctor")}>Doctor</button><button className="secondaryButton" type="button" onClick={() => runDiagnostic("/api/providers/probe", "providers")}>Providers</button><button className="secondaryButton" type="button" onClick={() => runDiagnostic("/api/mcp/check", "MCP")}>MCP</button><button className="secondaryButton" type="button" onClick={discoverSkills}>Skills</button></div></header>
      <div className="diagnosticSummary"><Metric label="Doctor" value={doctor?.status ?? "unknown"} /><Metric label="Checks" value={String(doctor?.checks.length ?? 0)} /><Metric label="Failing" value={String(failingChecks.length)} /></div>
      <div className="checkTable">{doctor?.checks.map((check) => <div key={check.id} className="checkRowItem"><strong>{check.id}</strong><span>{check.status}</span><p>{check.message}</p></div>)}</div>
    </section>;
  }

  function renderPrompts() {
    return <section className="screen promptsScreen"><header className="screenIntro withActions"><div><span className="sectionLabel">Prompts</span><h2>User-level additive prompt</h2><p>This prompt stays below daemon policy. It cannot override safety, routing, or workspace boundaries.</p><small>{prompt.path || "Prompt file has not been created yet"}</small></div><button className="primaryButton" type="button" onClick={savePrompt}>Save prompt</button></header><textarea className="promptEditor" value={promptDraft} onChange={(event) => setPromptDraft(event.target.value)} placeholder="Add coding standards, review expectations, naming rules, project preferences, and personal operating constraints." /></section>;
  }
}

function Intro(props: { label: string; title: string; text: string }) { return <header className="screenIntro"><span className="sectionLabel">{props.label}</span><h2>{props.title}</h2><p>{props.text}</p></header>; }
function Field(props: { label: string; value: string | number; onChange: (value: string) => void; type?: string; wide?: boolean }) { return <label className={props.wide ? "wideField" : undefined}>{props.label}<input type={props.type ?? "text"} value={props.value} onChange={(event) => props.onChange(event.target.value)} /></label>; }
function Action(props: { title: string; text: string; action: string; onClick: () => void }) { return <article className="actionCard"><h3>{props.title}</h3><p>{props.text}</p><button className="secondaryButton" type="button" onClick={props.onClick}>{props.action}</button></article>; }
function Metric(props: { label: string; value: string }) { return <article className="metricCard"><span>{props.label}</span><strong>{props.value}</strong></article>; }
function RailCard(props: { title: string; facts: Record<string, string> }) { return <div className="railCard"><span className="railLabel">{props.title}</span><dl className="factList">{Object.entries(props.facts).map(([key, value]) => <div key={key}><dt>{key}</dt><dd>{value}</dd></div>)}</dl></div>; }
function modelsOf(provider?: ProviderProfile): ModelProfile[] { if (!provider) return []; return Array.isArray(provider.models) && provider.models.length > 0 ? provider.models : [{ id: provider.defaultModel }]; }
function parseSelectedModel(value: string): { providerId: string; modelId: string } { const [providerId = "", modelId = ""] = value.split("::"); return { providerId, modelId }; }
function uniqueId(base: string, existing: string[]): string { const taken = new Set(existing); if (!taken.has(base)) return base; for (let i = 2; i < 1000; i += 1) { const candidate = `${base}-${i}`; if (!taken.has(candidate)) return candidate; } return `${base}-${Date.now()}`; }
function parseOptionalNumber(value: string): number | undefined { const trimmed = value.trim(); if (!trimmed) return undefined; const parsed = Number(trimmed); return Number.isFinite(parsed) ? parsed : undefined; }
