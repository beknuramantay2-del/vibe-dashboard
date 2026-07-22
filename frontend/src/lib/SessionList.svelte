<script>
  export let sessions = []
  export let onSelect = () => {}

  let search = ''
  let filter = 'all'

  $: filtered = sessions.filter(s => {
    if (filter === 'active' && s.status !== 'active') return false
    if (filter === 'completed' && s.status !== 'completed') return false
    if (search && !s.id.toLowerCase().includes(search.toLowerCase()) && !s.project?.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  function agentColor(agent) {
    if (agent === 'Claude Code') return '#a78bfa'
    if (agent === 'OpenCode') return '#34d399'
    if (agent === 'Codex CLI') return '#60a5fa'
    return '#999'
  }
</script>

<div class="session-list">
  <div class="header">
    <h2>Sessions</h2>
    <span class="count">{filtered.length}</span>
  </div>

  <div class="toolbar">
    <input class="search" bind:value={search} placeholder="Search sessions..." />
    <div class="filters">
      <button class="filter-btn" class:active={filter === 'all'} on:click={() => filter = 'all'}>All</button>
      <button class="filter-btn" class:active={filter === 'active'} on:click={() => filter = 'active'}>Active</button>
      <button class="filter-btn" class:active={filter === 'completed'} on:click={() => filter = 'completed'}>Done</button>
    </div>
  </div>

  <div class="table-header">
    <span>Agent</span>
    <span>ID / Project</span>
    <span>Status</span>
    <span>Cost</span>
    <span>Tokens</span>
    <span>Duration</span>
  </div>

  <div class="rows">
    {#each filtered as s}
      <button class="row" on:click={() => onSelect(s)}>
        <span class="agent-cell">
          <span class="dot" style="background: {agentColor(s.agent)}"></span>
          {s.agent}
        </span>
        <span class="id-cell">
          <span class="id">{s.id.slice(0, 16)}...</span>
          {#if s.project}<span class="project">{s.project}</span>{/if}
        </span>
        <span class="status-cell">
          <span class="status-badge" class:active={s.status === 'active'} class:completed={s.status === 'completed'}>
            {s.status}
          </span>
        </span>
        <span class="cost-cell">${s.cost?.toFixed(2)}</span>
        <span class="tokens-cell">{(s.inputTokens + s.outputTokens).toLocaleString()}</span>
        <span class="duration-cell">{s.duration}</span>
      </button>
    {:else}
      <div class="empty">No sessions found</div>
    {/each}
  </div>
</div>

<style>
  .session-list { display: flex; flex-direction: column; height: 100%; }
  .header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
  .header h2 { font-size: 20px; color: var(--text); }
  .count { font-size: 12px; background: var(--bg3); color: var(--text2); padding: 2px 8px; border-radius: var(--radius-sm); }
  .toolbar { display: flex; gap: 12px; margin-bottom: 16px; align-items: center; }
  .search { flex: 1; padding: 8px 12px; border-radius: var(--radius-md); border: 1px solid var(--border); background: var(--bg2); color: var(--text); font-size: 13px; outline: none; transition: border-color 0.2s var(--ease-out), box-shadow 0.2s var(--ease-out); }
  .search:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15); }
  .filters { display: flex; gap: 4px; }
  .filter-btn { padding: 6px 14px; border-radius: var(--radius-sm); border: 1px solid var(--border); background: var(--bg2); color: var(--text2); font-size: 12px; cursor: pointer; transition: all 0.2s var(--ease-out); }
  .filter-btn:hover { background: var(--bg3); color: var(--text); }
  .filter-btn:active { transform: scale(0.96); }
  .filter-btn.active { background: var(--accent); color: #fff; border-color: var(--accent); }
  .table-header { display: grid; grid-template-columns: 100px 1fr 80px 80px 80px 80px; gap: 12px; padding: 8px 12px; font-size: 11px; color: var(--text2); text-transform: uppercase; letter-spacing: 1px; border-bottom: 1px solid var(--border); }
  .rows { display: flex; flex-direction: column; gap: 2px; margin-top: 4px; overflow-y: auto; flex: 1; }
  .row { display: grid; grid-template-columns: 100px 1fr 80px 80px 80px 80px; gap: 12px; padding: 10px 12px; border: none; border-radius: var(--radius-sm); background: transparent; color: var(--text); font-size: 13px; cursor: pointer; text-align: left; align-items: center; transition: background-color 0.15s var(--ease-out); }
  .row:hover { background: var(--bg3); }
  .agent-cell { display: flex; align-items: center; gap: 6px; }
  .dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
  .id-cell { display: flex; flex-direction: column; }
  .id { font-family: var(--font-mono); font-size: 12px; }
  .project { font-size: 11px; color: var(--text2); }
  .status-badge { padding: 2px 8px; border-radius: var(--radius-sm); font-size: 11px; }
  .status-badge.active { background: rgba(34, 197, 94, 0.15); color: var(--green); }
  .status-badge.completed { background: rgba(100, 100, 100, 0.15); color: var(--text2); }
  .cost-cell { font-family: var(--font-mono); color: var(--accent); }
  .tokens-cell { font-family: var(--font-mono); }
  .duration-cell { color: var(--text2); }
  .empty { padding: 40px; text-align: center; color: var(--text2); }
</style>
