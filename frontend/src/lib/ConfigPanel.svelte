<script>
  export let agents = []
  export let theme = 'dark'
  export let onToggleTheme = () => {}

  function agentColor(agent) {
    if (agent === 'Claude Code') return '#a78bfa'
    if (agent === 'OpenCode') return '#34d399'
    if (agent === 'Codex CLI') return '#60a5fa'
    return '#999'
  }
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
          {theme === 'dark' ? '☀ Switch to Light' : '☾ Switch to Dark'}
        </button>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-title">Connected Agents ({agents.length})</div>
    <div class="card-body">
      {#if agents.length}
        {#each agents as agent}
          <div class="agent-card">
            <span class="dot" style="background: {agentColor(agent)}"></span>
            <span class="agent-name">{agent}</span>
            <span class="agent-status">connected</span>
          </div>
        {/each}
      {:else}
        <div class="empty-small">
          No agents detected. Start Claude Code, OpenCode, or Codex CLI to connect automatically.
        </div>
      {/if}
    </div>
  </div>

  <div class="card">
    <div class="card-title">Data Sources</div>
    <div class="card-body">
      <div class="source-row">
        <span class="source-name">Claude Code</span>
        <span class="source-path">~/.claude/projects/**/*.jsonl</span>
      </div>
      <div class="source-row">
        <span class="source-name">OpenCode</span>
        <span class="source-path">~/.opencode/opencode.db</span>
      </div>
      <div class="source-row">
        <span class="source-name">Codex CLI</span>
        <span class="source-path">~/.codex/logs/**/*.jsonl</span>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-title">Storage</div>
    <div class="card-body">
      <div class="source-row">
        <span class="source-name">Database</span>
        <span class="source-path">~/.vibe-dashboard/vibe.db</span>
      </div>
      <div class="source-row">
        <span class="source-name">Snapshots</span>
        <span class="source-path">~/.vibe-dashboard/snapshots/</span>
      </div>
      <div class="source-row">
        <span class="source-name">Logs</span>
        <span class="source-path">~/.vibe-dashboard/vibe-desktop.log</span>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-title">About</div>
    <div class="card-body">
      <p><strong>vibe-dashboard</strong> v0.2.1</p>
      <p>Desktop UI for monitoring AI coding agents — sessions, tokens, costs, cache hit rates, file diffs, and rollback.</p>
      <p>Built with Wails + Svelte + Go.</p>
      <p class="links">
        <a href="https://github.com/beknuramantay2-del/vibe-dashboard" target="_blank">GitHub Repository</a>
      </p>
    </div>
  </div>
</div>

<style>
  .config { max-width: 600px; }
  .header { margin-bottom: 20px; }
  .header h2 { font-size: 20px; color: var(--text); }
  .card {
    background: var(--bg2); border-radius: var(--radius-lg);
    border: 1px solid var(--border); margin-bottom: 16px; overflow: hidden;
    transition: background-color var(--transition-speed) var(--ease-out),
                border-color var(--transition-speed) var(--ease-out);
  }
  .card-title {
    padding: 12px 16px; font-size: 13px; font-weight: 600; color: var(--text);
    border-bottom: 1px solid var(--border);
  }
  .card-body { padding: 16px; }
  .theme-toggle { display: flex; justify-content: space-between; align-items: center; }
  .toggle-btn {
    padding: 8px 16px; border-radius: var(--radius-md); border: 1px solid var(--border);
    background: var(--bg3); color: var(--text); font-size: 13px; cursor: pointer;
    transition: all 0.2s var(--ease-out); font-family: var(--font-sans);
  }
  .toggle-btn:hover { background: var(--accent); color: #fff; border-color: var(--accent); }
  .toggle-btn:active { transform: scale(0.96); }
  .agent-card {
    display: flex; align-items: center; gap: 10px; padding: 10px;
    border-radius: var(--radius-md); background: var(--bg3); margin-bottom: 6px;
    transition: background-color 0.2s var(--ease-out);
  }
  .agent-card:hover { background: var(--bg-hover); }
  .dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
  .agent-name { flex: 1; font-size: 14px; color: var(--text); font-weight: 500; }
  .agent-status {
    font-size: 11px; color: var(--green); background: var(--green-bg);
    padding: 2px 8px; border-radius: 4px;
  }
  .source-row {
    display: flex; justify-content: space-between; padding: 8px 0; font-size: 13px;
    border-bottom: 1px solid var(--border);
  }
  .source-row:last-child { border-bottom: none; }
  .source-name { color: var(--text); font-weight: 500; }
  .source-path { color: var(--text2); font-family: var(--font-mono); font-size: 12px; }
  .empty-small { padding: 20px; text-align: center; color: var(--text3); font-size: 13px; }
  p { font-size: 13px; color: var(--text2); line-height: 1.6; margin-bottom: 6px; }
  p strong { color: var(--text); }
  .links { margin-top: 12px; }
  .links a {
    color: var(--accent); text-decoration: none; font-weight: 500;
    transition: opacity 0.2s var(--ease-out);
  }
  .links a:hover { opacity: 0.8; text-decoration: underline; }
</style>
