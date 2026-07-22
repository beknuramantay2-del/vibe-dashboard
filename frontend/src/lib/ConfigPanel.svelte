<script>
  export let agents = []
  export let theme = 'dark'
  export let onToggleTheme = () => {}
</script>

<div class="config">
  <div class="header">
    <h2>Configuration</h2>
  </div>

  <div class="card">
    <div class="card-title">Theme</div>
    <div class="card-body">
      <div class="theme-toggle">
        <span>Appearance</span>
        <button class="toggle-btn" on:click={onToggleTheme}>
          {theme === 'dark' ? '☀ Light' : '☾ Dark'}
        </button>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-title">Connected Agents</div>
    <div class="card-body">
      {#if agents.length}
        {#each agents as agent}
          <div class="agent-card">
            <span class="dot" class:claude={agent === 'Claude Code'} class:opencode={agent === 'OpenCode'} class:codex={agent === 'Codex CLI'}></span>
            <span class="agent-name">{agent}</span>
            <span class="agent-status">connected</span>
          </div>
        {/each}
      {:else}
        <div class="empty-small">No agents detected. Run Claude Code, OpenCode, or Codex CLI first.</div>
      {/if}
    </div>
  </div>

  <div class="card">
    <div class="card-title">Data Sources</div>
    <div class="card-body">
      <div class="source-row"><span class="source-name">Claude Code</span><span class="source-path">~/.claude/projects/*.jsonl</span></div>
      <div class="source-row"><span class="source-name">OpenCode</span><span class="source-path">~/.opencode/opencode.db</span></div>
      <div class="source-row"><span class="source-name">Codex CLI</span><span class="source-path">~/.codex/logs/*.jsonl</span></div>
    </div>
  </div>

  <div class="card">
    <div class="card-title">About</div>
    <div class="card-body">
      <p>vibe-dashboard v0.1 — Desktop UI for monitoring AI coding agents.</p>
      <p>Built with Wails + Svelte + Go.</p>
      <p class="links">
        <a href="https://github.com/beknuramantay2-del/vibe-dashboard" target="_blank">GitHub</a>
      </p>
    </div>
  </div>
</div>

<style>
  .config { max-width: 600px; }
  .header { margin-bottom: 20px; }
  .header h2 { font-size: 20px; color: var(--text); }
  .card { background: var(--bg2); border-radius: 10px; border: 1px solid var(--border); margin-bottom: 16px; overflow: hidden; }
  .card-title { padding: 12px 16px; font-size: 13px; font-weight: 600; color: var(--text); border-bottom: 1px solid var(--border); }
  .card-body { padding: 16px; }
  .theme-toggle { display: flex; justify-content: space-between; align-items: center; }
  .toggle-btn { padding: 8px 16px; border-radius: 8px; border: 1px solid var(--border); background: var(--bg3); color: var(--text); font-size: 13px; cursor: pointer; }
  .toggle-btn:hover { background: var(--accent); color: #fff; border-color: var(--accent); }
  .agent-card { display: flex; align-items: center; gap: 10px; padding: 10px; border-radius: 8px; background: var(--bg3); margin-bottom: 6px; }
  .dot { width: 10px; height: 10px; border-radius: 50%; }
  .dot.claude { background: #a78bfa; }
  .dot.opencode { background: #34d399; }
  .dot.codex { background: #60a5fa; }
  .agent-name { flex: 1; font-size: 14px; color: var(--text); font-weight: 500; }
  .agent-status { font-size: 11px; color: var(--green); background: rgba(0, 184, 148, 0.1); padding: 2px 8px; border-radius: 4px; }
  .source-row { display: flex; justify-content: space-between; padding: 6px 0; font-size: 13px; border-bottom: 1px solid var(--border); }
  .source-row:last-child { border-bottom: none; }
  .source-name { color: var(--text); font-weight: 500; }
  .source-path { color: var(--text2); font-family: monospace; font-size: 12px; }
  .empty-small { padding: 20px; text-align: center; color: var(--text2); font-size: 13px; }
  p { font-size: 13px; color: var(--text2); line-height: 1.6; margin-bottom: 8px; }
  .links { margin-top: 12px; }
  .links a { color: var(--accent); text-decoration: none; font-weight: 500; }
</style>
