<script>
  let before = ''
  let after = ''

  $: lines = computeDiff(before, after)

  function computeDiff(b, a) {
    if (!b && !a) return [{ type: 'empty', text: 'Enter before/after text to compare' }]
    const beforeLines = b.split('\n')
    const afterLines = a.split('\n')
    const result = []
    const maxLen = Math.max(beforeLines.length, afterLines.length)
    for (let i = 0; i < maxLen; i++) {
      const bl = beforeLines[i] ?? ''
      const al = afterLines[i] ?? ''
      if (bl === al) {
        result.push({ type: 'equal', text: `  ${bl}` })
      } else {
        if (bl !== undefined) result.push({ type: 'remove', text: `- ${bl}` })
        if (al !== undefined) result.push({ type: 'add', text: `+ ${al}` })
      }
    }
    return result
  }

  function handlePaste(evt, target) {
    const text = evt.clipboardData.getData('text')
    if (target === 'before') before = text
    else after = text
    evt.preventDefault()
  }
</script>

<div class="diff-viewer">
  <div class="header">
    <h2>Diff Viewer</h2>
  </div>

  <div class="inputs">
    <div class="input-group">
      <label for="diff-before">Before</label>
      <textarea id="diff-before" bind:value={before} on:paste={(e) => handlePaste(e, 'before')} placeholder="Paste original content..." class="editor"></textarea>
    </div>
    <div class="input-group">
      <label for="diff-after">After</label>
      <textarea id="diff-after" bind:value={after} on:paste={(e) => handlePaste(e, 'after')} placeholder="Paste new content..." class="editor"></textarea>
    </div>
  </div>

  <div class="output">
    <div class="output-header">
      <span>Diff</span>
      <span class="stats">
        <span class="add-count">+{lines.filter(l => l.type === 'add').length}</span>
        <span class="del-count">-{lines.filter(l => l.type === 'remove').length}</span>
      </span>
    </div>
    <div class="diff-content">
      {#each lines as line}
        <div class="diff-line" class:add={line.type === 'add'} class:remove={line.type === 'remove'} class:equal={line.type === 'equal'} class:empty={line.type === 'empty'}>
          <code>{line.text}</code>
        </div>
      {/each}
    </div>
  </div>
</div>

<style>
  .diff-viewer { height: 100%; display: flex; flex-direction: column; }
  .header { margin-bottom: 16px; }
  .header h2 { font-size: 20px; color: var(--text); }
  .inputs { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-bottom: 16px; }
  .input-group label { display: block; font-size: 11px; color: var(--text2); text-transform: uppercase; letter-spacing: 1px; margin-bottom: 6px; }
  .editor { width: 100%; height: 180px; padding: 12px; border-radius: var(--radius-md); border: 1px solid var(--border); background: var(--bg2); color: var(--text); font-family: var(--font-mono); font-size: 12px; resize: vertical; outline: none; transition: border-color 0.2s var(--ease-out), box-shadow 0.2s var(--ease-out); }
  .editor:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15); }
  .output { flex: 1; display: flex; flex-direction: column; }
  .output-header { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: var(--bg2); border-radius: var(--radius-md) var(--radius-md) 0 0; border: 1px solid var(--border); font-size: 12px; color: var(--text2); }
  .stats { display: flex; gap: 12px; }
  .add-count { color: var(--green); font-weight: 600; }
  .del-count { color: var(--red); font-weight: 600; }
  .diff-content { flex: 1; overflow-y: auto; background: var(--bg2); border: 1px solid var(--border); border-top: none; border-radius: 0 0 var(--radius-md) var(--radius-md); padding: 8px 0; font-family: var(--font-mono); font-size: 12px; }
  .diff-line { padding: 2px 12px; white-space: pre; }
  .diff-line code { font-family: inherit; }
  .diff-line.add { background: rgba(34, 197, 94, 0.1); }
  .diff-line.add code { color: var(--green); }
  .diff-line.remove { background: rgba(239, 68, 68, 0.1); }
  .diff-line.remove code { color: var(--red); }
  .diff-line.equal code { color: var(--text); }
  .diff-line.empty { padding: 20px; text-align: center; }
  .diff-line.empty code { color: var(--text2); }
</style>
