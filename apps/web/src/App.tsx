import { useEffect, useMemo, useState } from "react";

type DoctorReport = {
  status: string;
  checks: Array<{ id: string; status: string; message: string }>;
};

type MCPSystem = {
  id: string;
  displayName: string;
  status: string;
  description: string;
  tools?: Array<{ id: string; name: string }>;
};

type UserPrompt = {
  content: string;
  path: string;
  updatedAt?: string;
};

type ModelOption = {
  providerId: string;
  modelId: string;
  label: string;
};

type TranscriptItem = {
  role: "user" | "harness";
  text: string;
};

const commands = ["/goal", "/yolo", "/loop", "/swarm", "/export"];

const api = {
  async doctor(): Promise<DoctorReport> {
    const response = await fetch("/api/doctor");
    if (!response.ok) throw new Error("doctor request failed");
    return response.json();
  },
  async configDefaults(): Promise<unknown> {
    const response = await fetch("/api/config/defaults");
    if (!response.ok) throw new Error("config defaults request failed");
    return response.json();
  },
  async mcpSystems(): Promise<{ systems: MCPSystem[] }> {
    const response = await fetch("/api/mcp/systems");
    if (!response.ok) throw new Error("MCP request failed");
    return response.json();
  },
  async userPrompt(): Promise<UserPrompt> {
    const response = await fetch("/api/prompts/user");
    if (!response.ok) throw new Error("prompt request failed");
    return response.json();
  },
  async saveUserPrompt(content: string): Promise<UserPrompt> {
    const response = await fetch("/api/prompts/user", {
      method: "PUT",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ content })
    });
    if (!response.ok) throw new Error("prompt save failed");
    return response.json();
  },
  async applyControl(payload: Record<string, unknown>): Promise<unknown> {
    const response = await fetch("/api/control/apply", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(payload)
    });
    if (!response.ok) {
      const body = await response.json().catch(() => ({}));
      throw new Error(typeof body.error === "string" ? body.error : "control update failed");
    }
    return response.json();
  },
  async submitSessionMessage(payload: Record<string, unknown>): Promise<{ message: string }> {
    const response = await fetch("/api/session/message", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(payload)
    });
    const body = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(typeof body.error === "string" ? body.error : "session message failed");
    }
    return body as { message: string };
  }
};

export function App() {
  const [doctor, setDoctor] = useState<DoctorReport | null>(null);
  const [systems, setSystems] = useState<MCPSystem[]>([]);
  const [modelOptions, setModelOptions] = useState<ModelOption[]>([]);
  const [selectedModel, setSelectedModel] = useState("");
  const [prompt, setPrompt] = useState<UserPrompt>({ content: "", path: "" });
  const [promptDraft, setPromptDraft] = useState("");
  const [workspaceRoot, setWorkspaceRoot] = useState(".");
  const [sessionId, setSessionId] = useState(() => `web-${Date.now()}`);
  const [topic, setTopic] = useState("");
  const [goal, setGoal] = useState("");
  const [commandInput, setCommandInput] = useState("");
  const [sessionPrompt, setSessionPrompt] = useState("");
  const [transcript, setTranscript] = useState<TranscriptItem[]>([]);
  const [yolo, setYolo] = useState(false);
  const [message, setMessage] = useState("Loading daemon state");

  useEffect(() => {
    let alive = true;
    Promise.all([api.doctor(), api.mcpSystems(), api.userPrompt(), api.configDefaults()])
      .then(([doctorReport, mcpReport, userPrompt, defaults]) => {
        if (!alive) return;
        const options = extractModelOptions(defaults);
        setDoctor(doctorReport);
        setSystems(mcpReport.systems);
        setPrompt(userPrompt);
        setPromptDraft(userPrompt.content ?? "");
        setModelOptions(options);
        setSelectedModel(options[0] ? optionValue(options[0]) : "");
        setMessage("Daemon state loaded");
      })
      .catch((error: unknown) => {
        if (!alive) return;
        setMessage(error instanceof Error ? error.message : "Unable to load daemon state");
      });
    return () => {
      alive = false;
    };
  }, []);

  const failingChecks = useMemo(() => doctor?.checks.filter((check) => check.status !== "pass") ?? [], [doctor]);
  const selected = useMemo(() => parseSelectedModel(selectedModel), [selectedModel]);

  async function savePrompt() {
    setMessage("Saving user prompt");
    try {
      const saved = await api.saveUserPrompt(promptDraft);
      setPrompt(saved);
      setPromptDraft(saved.content ?? "");
      setMessage("User-level system prompt saved");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Unable to save user prompt");
    }
  }

  async function applyControls() {
    setMessage("Applying session controls");
    try {
      await api.applyControl({ workspaceRoot, sessionId, topic, goal, yolo });
      setMessage("Goal, topic, and mode controls saved");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Unable to apply controls");
    }
  }

  async function submitPrompt() {
    const promptText = sessionPrompt.trim();
    const commandText = commandInput.trim();
    if (!promptText && !commandText) {
      setMessage("Enter a prompt or command first");
      return;
    }
    setTranscript((items) => [...items, { role: "user", text: [commandText, promptText].filter(Boolean).join("\n\n") }]);
    setMessage("Saving session prompt");
    try {
      const result = await api.submitSessionMessage({
        workspaceRoot,
        sessionId,
        providerId: selected.providerId,
        modelId: selected.modelId,
        command: commandText,
        prompt: promptText
      });
      setTranscript((items) => [...items, { role: "harness", text: result.message }]);
      setSessionPrompt("");
      setCommandInput("");
      setMessage("Session prompt stored");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Unable to store session prompt");
    }
  }

  return (
    <main className="shell">
      <section className="hero" aria-labelledby="title">
        <div className="heroTopline">
          <p className="eyebrow">Ubuntu-native web harness</p>
          <p className="statusText">{message}</p>
        </div>
        <h1 id="title">Harnejr</h1>
        <p className="lede">
          Local web control for agentic coding sessions, workspace safety, provider routing, MCP systems, goal control, quality gates, prompts, commands, and evidence-backed completion.
        </p>
        <div className="commandRow" aria-label="engineered commands">
          {commands.map((command) => (
            <button type="button" key={command} onClick={() => setCommandInput(command)}>{command}</button>
          ))}
        </div>
      </section>

      <section className="panel split" aria-label="session console">
        <div>
          <p className="eyebrow">Session console</p>
          <h2>Prompt, command, and model</h2>
          <p>
            Enter a prompt, choose a configured model, or use an engineered command. Until provider execution is implemented, submissions are saved into workspace memory for traceability.
          </p>
          <div className="transcriptBox" aria-label="session transcript">
            {transcript.length === 0 ? <p className="metaText">No messages yet.</p> : null}
            {transcript.map((item, index) => (
              <article key={`${item.role}-${index}`} className="transcriptItem">
                <strong>{item.role === "user" ? "You" : "Harnejr"}</strong>
                <p>{item.text}</p>
              </article>
            ))}
          </div>
        </div>
        <div className="formGrid">
          <label>
            Model
            <select value={selectedModel} onChange={(event) => setSelectedModel(event.target.value)}>
              {modelOptions.length === 0 ? <option value="">No configured models found</option> : null}
              {modelOptions.map((option) => (
                <option key={optionValue(option)} value={optionValue(option)}>{option.label}</option>
              ))}
            </select>
          </label>
          <label>
            Command
            <input value={commandInput} onChange={(event) => setCommandInput(event.target.value)} placeholder="/goal, /yolo, /swarm, /export, or a plain command" />
          </label>
          <label>
            Prompt
            <textarea value={sessionPrompt} onChange={(event) => setSessionPrompt(event.target.value)} rows={10} placeholder="Ask Harnejr to inspect, plan, edit, review, verify, or continue a goal." />
          </label>
          <button type="button" onClick={submitPrompt}>Submit to session</button>
        </div>
      </section>

      <section className="grid gridThree" aria-label="daemon readiness">
        <article className="card">
          <h2>Doctor</h2>
          <p>Readiness: {doctor?.status ?? "unknown"}</p>
          <p>{failingChecks.length === 0 ? "No failing checks reported." : `${failingChecks.length} check needs attention.`}</p>
        </article>
        <article className="card">
          <h2>Built-in MCP</h2>
          <p>{systems.length} local harness systems are registered out of the box.</p>
          <p>Doctor, LoC, workspace memory, healing, goal/topic control, and context efficiency.</p>
        </article>
        <article className="card">
          <h2>Quality gates</h2>
          <p>LoC controller and healing planner are daemon endpoints, not prompt reminders.</p>
          <p>Oversized source files can block completion review.</p>
        </article>
      </section>

      <section className="panel split" aria-label="session controls">
        <div>
          <p className="eyebrow">Session controls</p>
          <h2>Goal, topic, and autonomy</h2>
          <p>Controls are persisted into the selected workspace memory so the session does not lose its objective or topic.</p>
        </div>
        <div className="formGrid">
          <label>Workspace root<input value={workspaceRoot} onChange={(event) => setWorkspaceRoot(event.target.value)} /></label>
          <label>Session ID<input value={sessionId} onChange={(event) => setSessionId(event.target.value)} /></label>
          <label>Topic<input value={topic} onChange={(event) => setTopic(event.target.value)} placeholder="Provider routing, web UI, policy engine" /></label>
          <label>Goal<textarea value={goal} onChange={(event) => setGoal(event.target.value)} rows={4} placeholder="Define the session objective." /></label>
          <label className="checkRow"><input type="checkbox" checked={yolo} onChange={(event) => setYolo(event.target.checked)} />Enable yolo for safe workspace work</label>
          <button type="button" onClick={applyControls}>Apply controls</button>
        </div>
      </section>

      <section className="panel split" aria-label="user prompt editor">
        <div>
          <p className="eyebrow">User-level prompt</p>
          <h2>Permanent additive system prompt</h2>
          <p>This prompt is saved by the daemon and added to Harnejr's fundamental harness prompt for every session. It does not replace core safety or runtime policy.</p>
          <p className="metaText">Storage: {prompt.path || "not created yet"}</p>
          {prompt.updatedAt ? <p className="metaText">Last updated: {new Date(prompt.updatedAt).toLocaleString()}</p> : null}
        </div>
        <div className="editorBox">
          <textarea value={promptDraft} onChange={(event) => setPromptDraft(event.target.value)} placeholder="Add personal operating preferences, coding standards, naming conventions, review expectations, or project habits." rows={12} />
          <button type="button" onClick={savePrompt}>Save prompt</button>
        </div>
      </section>

      <section className="panel" aria-label="mcp systems">
        <p className="eyebrow">Local systems</p>
        <h2>Ready out-of-the-box MCP layer</h2>
        <div className="systemList">
          {systems.map((system) => (
            <article key={system.id} className="systemItem">
              <div><h3>{system.displayName}</h3><p>{system.description}</p></div>
              <span>{system.status}</span>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}

function extractModelOptions(defaults: unknown): ModelOption[] {
  const root = defaults as { providers?: { providers?: unknown[] } };
  const providers = Array.isArray(root.providers?.providers) ? root.providers.providers : [];
  return providers.flatMap((provider): ModelOption[] => {
    const p = provider as { id?: string; displayName?: string; defaultModel?: string; models?: Array<{ id?: string; displayName?: string }> };
    const providerId = p.id ?? "unknown-provider";
    const providerName = p.displayName ?? providerId;
    const models = Array.isArray(p.models) && p.models.length > 0 ? p.models : [{ id: p.defaultModel ?? "default" }];
    return models.map((model) => {
      const modelId = model.id ?? p.defaultModel ?? "default";
      const modelName = model.displayName ?? modelId;
      return { providerId, modelId, label: `${providerName} / ${modelName}` };
    });
  });
}

function optionValue(option: ModelOption): string {
  return `${option.providerId}::${option.modelId}`;
}

function parseSelectedModel(value: string): { providerId: string; modelId: string } {
  const [providerId = "", modelId = ""] = value.split("::");
  return { providerId, modelId };
}
