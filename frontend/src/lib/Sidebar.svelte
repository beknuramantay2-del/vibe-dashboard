<script>
  export let agents = []
  export let totalCost = 0
  export let selectedTab = 'sessions'
  export let theme = 'dark'
  export let onTabChange = () => {}
  export let onRefresh = () => {}
  export let onToggleTheme = () => {}

  const tabs = [
    { id: 'sessions', label: 'Sessions', icon: '◉' },
    { id: 'detail', label: 'Detail', icon: '◎' },
    { id: 'diff', label: 'Diff', icon: '≡' },
    { id: 'config', label: 'Config', icon: '⚙' },
  ]
</script>

<aside class="sidebar">
  <div class="logo">
    <span class="logo-icon">◈</span>
    <span class="logo-text">vibe</span>
  </div>

  <nav class="nav">
    {#each tabs as tab}
      <button class="nav-item" class:active={selectedTab === tab.id} on:click={() => onTabChange(tab.id)}>
        <span class="nav-icon">{tab.icon}</span>
        <span>{tab.label}</span>
      </button>
    {/each}
  </nav>

  <div class="agents">
    <div class="section-label">Agents</div>
    {#each agents as agent}
      <div class="agent-row">
        <span class="agent-dot" class:claude={agent === 'Claude Code'} class:opencode={agent === 'OpenCode'} class:codex={agent === 'Codex CLI'}></span>
        <span>{agent}</span>
      </div>
    {/each}
  </div>

  <div class="cost-badge">
    <div class="cost-label">Total Cost</div>
    <div class="cost-value">${totalCost.toFixed(2)}</div>
  </div>

  <div class="sidebar-footer">
    <button class="icon-btn" on:click={onRefresh} title="Refresh">⟳</button>
    <button class="icon-btn" on:click={onToggleTheme} title="Toggle theme">
      {theme === 'dark' ? '☀' : '☾'}
    </button>
  </div>
</aside>

<style>
  .sidebar {
    width: 200px;
    background: var(--bg2);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    padding: 16px;
    flex-shrink: 0;
    transition: background-color 0.3s var(--ease-out), border-color 0.3s var(--ease-out);
  }
  .logo { display: flex; align-items: center; gap: 8px; margin-bottom: 24px; }
  .logo-icon { font-size: 24px; color: var(--accent); transition: color 0.3s var(--ease-out); }
  .logo-text { font-size: 18px; font-weight: 700; color: var(--text); letter-spacing: -0.5px; }
  .nav { display: flex; flex-direction: column; gap: 2px; margin-bottom: 24px; }
  .nav-item { display: flex; align-items: center; gap: 10px; padding: 10px 12px; border: none; border-radius: var(--radius-md); background: transparent; color: var(--text2); font-size: 14px; cursor: pointer; transition: all 0.2s var(--ease-out); }
  .nav-item:hover { background: var(--bg3); color: var(--text); }
  .nav-item:active { transform: scale(0.97); }
  .nav-item.active { background: var(--accent); color: #fff; }
  .nav-icon { font-size: 16px; }
  .section-label { font-size: 11px; text-transform: uppercase; letter-spacing: 1px; color: var(--text2); margin-bottom: 8px; }
  .agents { margin-bottom: 24px; }
  .agent-row { display: flex; align-items: center; gap: 8px; padding: 6px 0; font-size: 13px; color: var(--text); }
  .agent-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--text2); transition: transform 0.2s var(--ease-out); }
  .agent-row:hover .agent-dot { transform: scale(1.3); }
  .agent-dot.claude { background: #a78bfa; }
  .agent-dot.opencode { background: #34d399; }
  .agent-dot.codex { background: #60a5fa; }
  .cost-badge { background: var(--bg3); border-radius: var(--radius-lg); padding: 12px; margin-bottom: auto; text-align: center; transition: background-color 0.3s var(--ease-out); }
  .cost-label { font-size: 11px; color: var(--text2); text-transform: uppercase; letter-spacing: 1px; }
  .cost-value { font-size: 20px; font-weight: 700; color: var(--accent); margin-top: 4px; transition: color 0.3s var(--ease-out); }
  .sidebar-footer { display: flex; gap: 8px; margin-top: 16px; }
  .icon-btn { width: 36px; height: 36px; border-radius: var(--radius-md); border: 1px solid var(--border); background: var(--bg2); color: var(--text2); font-size: 16px; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s var(--ease-out); }
  .icon-btn:hover { background: var(--bg3); color: var(--text); border-color: var(--accent); }
  .icon-btn:active { transform: scale(0.93); }
</style>
