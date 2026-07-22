<script>
  export let sessions = []
  export let onSelect = () => {}

  let search = ''
  let filter = 'all'
  let sortBy = 'time'
  let sortDir = 'desc'

  $: filtered = sessions
    .filter(s => {
      if (filter === 'active' && s.status !== 'active') return false
      if (filter === 'completed' && s.status !== 'completed') return false
      if (search) {
        const q = search.toLowerCase()
        const matchId = s.id?.toLowerCase().includes(q)
        const matchProject = s.project?.toLowerCase().includes(q)
        const matchAgent = s.agent?.toLowerCase().includes(q)
        if (!matchId && !matchProject && !matchAgent) return false
      }
      return true
    })
    .sort((a, b) => {
      let cmp = 0
      switch (sortBy) {
        case 'cost': cmp = (a.cost || 0) - (b.cost || 0); break
        case 'tokens': cmp = ((a.inputTokens || 0) + (a.outputTokens || 0)) - ((b.inputTokens || 0) + (b.outputTokens || 0)); break
        case 'agent': cmp = (a.agent || '').localeCompare(b.agent || ''); break
        case 'status': cmp = (a.status || '').localeCompare(b.status || ''); break
        case 'cache': cmp = (a.cacheHitRate || 0) - (b.cacheHitRate || 0); break
        default: cmp = (a.startTime || '').localeCompare(b.startTime || ''); break
      }
      return sortDir === 'desc' ? -cmp : cmp
    })

  function agentColor(agent) {
    if (agent === 'Claude Code') return '#a78bfa'
    if (agent === 'OpenCode') return '#34d399'
    if (agent === 'Codex CLI') return '#60a5fa'
    return '#999'
  }

  function toggleSort(col) {
    if (sortBy === col) {
      sortDir = sortDir === 'desc' ? 'asc' : 'desc'
    } else {
      sortBy = col
      sortDir = 'desc'
    }
  }

  function sortIndicator(col) {
    if (sortBy !== col) return ''
    return sortDir === 'desc' ? ' ↓' : ' ↑'
  }

  function cacheColor(rate) {
    if (rate >= 80) return 'var(--green)'
    if (rate >= 50) return 'var(--yellow)'
    return 'var(--red)'
  }

  $: activeCount = sessions.filter(s => s.status === 'active').length
</script>

<div class="session-list">
  <div class="header">
    <h2>Sessions</h2>
    <span class="count">{filtered.length}</span>
    {#if activeCount > 0}
      <span class="active-badge">{activeCount} active</span>
    {/if}
  </div>

  <div class="toolbar">
    <input
      class="search"
      bind:value={search}
      placeholder="Search sessions, projects, agents..."
    />
    <div class="filters">
      <button class="filter-btn" class:active={filter === 'all'} on:click={() => filter = 'all'}>All</button>
      <button class="filter-btn" class:active={filter === 'active'} on:click={() => filter = 'active'}>Active</button>
      <button class="filter-btn" class:active={filter === 'completed'} on:click={() => filter = 'completed'}>Done</button>
    </div>
  </div>

  <div class="table-header">
    <button class="th" on:click={() => toggleSort('agent')}>Agent{sortIndicator('agent')}</button>
    <button class="th th-wide" on:click={() => toggleSort('time')}>ID / Project{sortIndicator('time')}</button>
    <button class="th" on:click={() => toggleSort('status')}>Status{sortIndicator('status')}</button>
    <button class="th" on:click={() => toggleSort('cost')}>Cost{sortIndicator('cost')}</button>
    <button class="th" on:click={() => toggleSort('tokens')}>Tokens{sortIndicator('tokens')}</button>
    <button class="th" on:click={() => toggleSort('cache')}>Cache{sortIndicator('cache')}</button>
    <button class="th" on:click={() => toggleSort('time')}>Time{sortIndicator('time')}</button>
  </div>

  <div class="rows">
    {#each filtered as s (s.id)}
      <button class="row" on:click={() => onSelect(s)}>
        <span class="agent-cell">
          <span class="dot" style="background: {agentColor(s.agent)}"></span>
          <span class="agent-label">{s.agent}</span>
        </span>
        <span class="id-cell">
          <span class="id">{s.id.slice(0, 16)}…</span>
          {#if s.project}<span class="project">{s.project}</span>{/if}
        </span>
        <span class="status-cell">
          <span class="status-badge" class:active={s.status === 'active'} class:completed={s.status === 'completed'}>
            {#if s.status === 'active'}<span class="pulse"></span>{/if}
            {s.status}
          </span>
        </span>
        <span class="cost-cell">${s.cost?.toFixed(3)}</span>
        <span class="tokens-cell">{((s.inputTokens || 0) + (s.outputTokens || 0)).toLocaleString()}</span>
        <span class="cache-cell" style="color: {cacheColor(s.cacheHitRate || 0)}">{(s.cacheHitRate || 0).toFixed(0)}%</span>
        <span class="duration-cell">{s.duration || s.startTime}</span>
      </button>
    {:else}
      <div class="empty">
        {#if search}
          No sessions matching "{search}"
        {:else}
          No sessions found. Start an AI coding agent to see data.
        {/if}
      </div>
    {/each}
  </div>
</div>

<style>
  .session-list { display: flex; flex-direction: column; height: 100%; }
  .header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
  .header h2 { font-size: 20px; color: var(--text); }
  .count {
    font-size: 12px; background: var(--bg3); color: var(--text2);
    padding: 2px 8px; border-radius: var(--radius-sm);
  }
  .active-badge {
    font-size: 11px; background: var(--green-bg); color: var(--green);
    padding: 2px 8px; border-radius: var(--radius-sm); font-weight: 600;
  }
  .toolbar { display: flex; gap: 12px; margin-bottom: 16px; align-items: center; flex-wrap: wrap; }
  .search {
    flex: 1; min-width: 200px; padding: 8px 12px; border-radius: var(--radius-md);
    border: 1px solid var(--border); background: var(--bg2); color: var(--text);
    font-size: 13px; outline: none; font-family: var(--font-sans);
    transition: border-color 0.2s var(--ease-out), box-shadow 0.2s var(--ease-out);
  }
  .search:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15); }
  .search::placeholder { color: var(--text3); }
  .filters { display: flex; gap: 4px; }
  .filter-btn {
    padding: 6px 14px; border-radius: var(--radius-sm); border: 1px solid var(--border);
    background: var(--bg2); color: var(--text2); font-size: 12px; cursor: pointer;
    transition: all 0.2s var(--ease-out); font-family: var(--font-sans);
  }
  .filter-btn:hover { background: var(--bg3); color: var(--text); }
  .filter-btn:active { transform: scale(0.96); }
  .filter-btn.active { background: var(--accent); color: #fff; border-color: var(--accent); }

  .table-header {
    display: grid;
    grid-template-columns: 100px 1fr 80px 70px 90px 60px 80px;
    gap: 8px; padding: 0 12px;
    border-bottom: 1px solid var(--border);
  }
  .th {
    font-size: 11px; color: var(--text3); text-transform: uppercase; letter-spacing: 0.5px;
    background: none; border: none; cursor: pointer; padding: 8px 0; text-align: left;
    transition: color 0.15s var(--ease-out); font-family: var(--font-sans);
    white-space: nowrap;
  }
  .th:hover { color: var(--text); }
  .th-wide { min-width: 0; }

  .rows { display: flex; flex-direction: column; gap: 1px; margin-top: 4px; overflow-y: auto; flex: 1; }
  .row {
    display: grid;
    grid-template-columns: 100px 1fr 80px 70px 90px 60px 80px;
    gap: 8px; padding: 10px 12px; border: none;
    border-radius: var(--radius-sm); background: transparent;
    color: var(--text); font-size: 13px; cursor: pointer;
    text-align: left; align-items: center;
    transition: background-color 0.15s var(--ease-out);
    font-family: var(--font-sans);
    width: 100%;
  }
  .row:hover { background: var(--bg3); }
  .agent-cell { display: flex; align-items: center; gap: 6px; overflow: hidden; }
  .agent-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 12px; }
  .dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
  .id-cell { display: flex; flex-direction: column; min-width: 0; }
  .id { font-family: var(--font-mono); font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .project { font-size: 11px; color: var(--text2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .status-badge {
    display: inline-flex; align-items: center; gap: 4px;
    padding: 2px 8px; border-radius: var(--radius-sm); font-size: 11px;
  }
  .status-badge.active { background: var(--green-bg); color: var(--green); }
  .status-badge.completed { background: rgba(100, 100, 100, 0.15); color: var(--text2); }
  .pulse {
    width: 6px; height: 6px; border-radius: 50%; background: var(--green);
    animation: pulse 2s ease-in-out infinite;
  }
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }
  .cost-cell { font-family: var(--font-mono); color: var(--accent); font-size: 12px; }
  .tokens-cell { font-family: var(--font-mono); font-size: 12px; }
  .cache-cell { font-family: var(--font-mono); font-size: 12px; font-weight: 600; }
  .duration-cell { color: var(--text2); font-size: 12px; }
  .empty { padding: 40px; text-align: center; color: var(--text2); font-size: 14px; }
</style>
