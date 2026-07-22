<script>
  import { onMount, onDestroy } from 'svelte'
  import Sidebar from './lib/Sidebar.svelte'
  import SessionList from './lib/SessionList.svelte'
  import SessionDetail from './lib/SessionDetail.svelte'
  import DiffViewer from './lib/DiffViewer.svelte'
  import ConfigPanel from './lib/ConfigPanel.svelte'
  import { GetSessions, GetConnectedAgents, GetAggregatedCost } from '../wailsjs/go/main/App'
  import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'

  let sessions = []
  let agents = []
  let selectedTab = 'sessions'
  let selectedSession = null
  let totalCost = 0
  let theme = 'dark'

  onMount(() => {
    loadData()
    EventsOn('sessions-updated', (data) => {
      sessions = data
    })
  })

  onDestroy(() => {
    EventsOff('sessions-updated')
  })

  async function loadData() {
    sessions = await GetSessions()
    agents = await GetConnectedAgents()
    totalCost = await GetAggregatedCost()
  }

  function selectSession(s) {
    selectedSession = s
  }

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark'
    document.documentElement.setAttribute('data-theme', theme)
  }
</script>

<div class="app" class:dark={theme === 'dark'} class:light={theme === 'light'}>
  <Sidebar {agents} {totalCost} {selectedTab} onTabChange={(t) => selectedTab = t} onRefresh={loadData} {theme} onToggleTheme={toggleTheme} />
  <main class="main-content">
    {#if selectedTab === 'sessions'}
      <SessionList {sessions} onSelect={selectSession} />
    {:else if selectedTab === 'detail' && selectedSession}
      <SessionDetail session={selectedSession} {agents} />
    {:else if selectedTab === 'diff'}
      <DiffViewer />
    {:else if selectedTab === 'config'}
      <ConfigPanel {agents} {theme} onToggleTheme={toggleTheme} />
    {/if}
  </main>
</div>

<style>
  .app {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
  }
  .main-content {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }
  :global(*) { margin: 0; padding: 0; box-sizing: border-box; }
  :global(body) { margin: 0; overflow: hidden; }
  :global(::-webkit-scrollbar) { width: 6px; }
  :global(::-webkit-scrollbar-track) { background: transparent; }
  :global(::-webkit-scrollbar-thumb) { background: #555; border-radius: 3px; }

  :global(html[data-theme='dark']) { --bg: #0f0f14; --bg2: #1a1a24; --bg3: #252535; --text: #e0e0e0; --text2: #999; --accent: #6c5ce7; --border: #2a2a3a; --green: #00b894; --red: #e17055; --yellow: #fdcb6e; }
  :global(html[data-theme='light']) { --bg: #ffffff; --bg2: #f5f5f8; --bg3: #e8e8ee; --text: #1a1a2e; --text2: #666; --accent: #6c5ce7; --border: #d0d0da; --green: #00b894; --red: #e17055; --yellow: #fdcb6e; }
</style>
