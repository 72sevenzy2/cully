import { ref } from 'vue'

export const agents = [
  { id: 'codex', label: 'Codex' },
  { id: 'claude', label: 'Claude Code' },
  { id: 'cursor', label: 'Cursor' }
]

// Shared so every command block on a page follows the selected agent.
export const selectedAgent = ref('codex')
