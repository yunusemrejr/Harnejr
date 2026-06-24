const surfaces = [
  {
    title: "Daemon control",
    body: "Local Go process owns sessions, policy, workspace safety, provider routing, and export state."
  },
  {
    title: "Provider registry",
    body: "Providers are modeled as transport contracts with protocol, auth, billing mode, endpoint, model namespace, parser, and fallback policy."
  },
  {
    title: "Autonomy gates",
    body: "Goal, loop, yolo, swarm, and export are deterministic harness actions, not prompt-only reminders."
  },
  {
    title: "Ubuntu safety",
    body: "Dangerous shell and filesystem operations are denied before a model can execute them. Safe alternatives should continue without user prompts."
  }
];

const commands = ["/goal", "/yolo", "/loop", "/swarm", "/export"];

export function App() {
  return (
    <main className="shell">
      <section className="hero" aria-labelledby="title">
        <p className="eyebrow">Ubuntu-native web harness</p>
        <h1 id="title">Harnejr</h1>
        <p className="lede">
          A local web GUI for autonomous coding sessions with enforced safety, provider routing, subagents, judges, skills, MCPs, cache discipline, and proof-backed completion.
        </p>
        <div className="commandRow" aria-label="engineered commands">
          {commands.map((command) => (
            <span key={command}>{command}</span>
          ))}
        </div>
      </section>

      <section className="grid" aria-label="core surfaces">
        {surfaces.map((surface) => (
          <article className="card" key={surface.title}>
            <h2>{surface.title}</h2>
            <p>{surface.body}</p>
          </article>
        ))}
      </section>

      <section className="panel" aria-label="next implementation focus">
        <h2>Next implementation focus</h2>
        <p>
          The scaffold now needs persistent session state, provider health probes, workspace file APIs, the real command dispatcher, and a bounded subagent scheduler.
        </p>
      </section>
    </main>
  );
}
