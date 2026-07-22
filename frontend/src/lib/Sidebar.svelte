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
    { id: 'snapshots', label: 'Snapshots', icon: '⏱' },
    { id: 'config', label: 'Config', icon: '⚙' },
  ]

  function agentColor(agent) {
    if (agent === 'Claude Code') return '#a78bfa'
    if (agent === 'OpenCode') return '#34d399'
    if (agent === 'Codex CLI') return '#60a5fa'
    return '#999'
  }
</script>

<aside class="sidebar">
  <div class="logo">
    <span class="logo-icon">◈</span>
    <span class="logo-text">vibe</span>
  </div>

  <nav class="nav">
    {#each tabs as tab}
      <button
        class="nav-item"
        class:active={selectedTab === tab.id}
        on:click={() => onTabChange(tab.id)}
        title={tab.label}
      >
        <span class="nav-icon">{tab.icon}</span>
        <span class="nav-label">{tab.label}</span>
      </button>
    {/each}
  </nav>

  <div class="agents">
    <div class="section-label">Agents</div>
    {#if agents.length}
      {#each agents as agent}
        <div class="agent-row">
          <span class="agent-dot" style="background: {agentColor(agent)}"></span>
          <span class="agent-name">{agent}</span>
        </div>
      {/each}
    {:else}
      <div class="no-agents">No agents detected</div>
    {/if}
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
    min-width: 200px;
    background: var(--bg2);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    padding: 16px;
    flex-shrink: 0;
    transition: background-color var(--transition-speed) var(--ease-out),
                border-color var(--transition-speed) var(--ease-out);
    overflow-y: auto;
  }
  .logo { display: flex; align-items: center; gap: 8px; margin-bottom: 24px; }
  .logo-icon { font-size: 24px; color: var(--accent); }
  .logo-text { font-size: 18px; font-weight: 700; color: var(--text); letter-spacing: -0.5px; }
  .nav { display: flex; flex-direction: column; gap: 2px; margin-bottom: 24px; }
  .nav-item {
    display: flex; align-items: center; gap: 10px;
    padding: 10px 12px; border: none; border-radius: var(--radius-md);
    background: transparent; color: var(--text2); font-size: 14px;
    cursor: pointer; transition: all 0.2s var(--ease-out);
    font-family: var(--font-sans);
  }
  .nav-item:hover { background: var(--bg3); color: var(--text); }
  .nav-item:active { transform: scale(0.97); }
  .nav-item.active { background: var(--accent); color: #fff; }
  .nav-icon { font-size: 16px; width: 20px; text-align: center; }
  .nav-label { white-space: nowrap; }
  .section-label {
    font-size: 11px; text-transform: uppercase; letter-spacing: 1px;
    color: var(--text3); margin-bottom: 8px;
  }
  .agents { margin-bottom: 24px; }
  .agent-row {
    display: flex; align-items: center; gap: 8px;
    padding: 6px 0; font-size: 13px; color: var(--text);
  }
  .agent-dot {
    width: 8px; height: 8px; border-radius: 50%;
    transition: transform 0.2s var(--ease-out);
    flex-shrink: 0;
  }
  .agent-row:hover .agent-dot { transform: scale(1.3); }
  .agent-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .no-agents { font-size: 12px; color: var(--text3); padding: 4px 0; }
  .cost-badge {
    background: var(--bg3); border-radius: var(--radius-lg);
    padding: 12px; margin-bottom: auto; text-align: center;
    transition: background-color var(--transition-speed) var(--ease-out);
  }
  .cost-label { font-size: 11px; color: var(--text2); text-transform: uppercase; letter-spacing: 1px; }
  .cost-value { font-size: 20px; font-weight: 700; color: var(--accent); margin-top: 4px; }
  .sidebar-footer { display: flex; gap: 8px; margin-top: 16px; }
  .icon-btn {
    width: 36px; height: 36px; border-radius: var(--radius-md);
    border: 1px solid var(--border); background: var(--bg2); color: var(--text2);
    font-size: 16px; cursor: pointer;
    display: flex; align-items: center; justify-content: center;
    transition: all 0.2s var(--ease-out);
  }
  .icon-btn:hover { background: var(--bg3); color: var(--text); border-color: var(--accent); }
  .icon-btn:active { transform: scale(0.93); }
</style>
