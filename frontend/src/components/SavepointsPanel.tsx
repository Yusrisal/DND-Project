import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Savepoint } from '../api/types'

type Props = {
  campaignId: number | null
  currentMapId: number | null
  onReloadBoard?: () => void
  onError: (error: unknown) => void
}

export function SavepointsPanel({ campaignId, currentMapId, onReloadBoard, onError }: Props) {
  const [savepoints, setSavepoints] = useState<Savepoint[]>([])
  const [search, setSearch] = useState('')
  const [isCreating, setIsCreating] = useState(false)
  const [name, setName] = useState('')
  const [notes, setNotes] = useState('')

  const refresh = useCallback(() => {
    if (!campaignId) return
    api
      .listSavepoints(campaignId, { search, page_size: 50, sort_by: 'created_at', sort_dir: 'desc' })
      .then((page) => setSavepoints(page.items))
      .catch(onError)
  }, [campaignId, search, onError])

  useEffect(() => {
    refresh()
  }, [refresh])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!campaignId) return
    try {
      await api.createSavepoint(campaignId, {
        campaign_id: campaignId,
        map_id: currentMapId,
        name: name.trim() || 'Session Checkpoint',
        notes: notes.trim(),
      })
      setName('')
      setNotes('')
      setIsCreating(false)
      refresh()
      alert('✅ Checkpoint / Savepoint created successfully!')
    } catch (err) {
      onError(err)
    }
  }

  const handleRestore = async (sp: Savepoint) => {
    if (!confirm(`Restore game state to "${sp.name}" (${new Date(sp.created_at).toLocaleString()})?\nThis will revert token positions and HP back to this savepoint.`)) {
      return
    }
    try {
      await api.restoreSavepoint(sp.id)
      onReloadBoard?.()
      alert('💾 Game state restored from savepoint!')
    } catch (err) {
      onError(err)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this savepoint?')) return
    try {
      await api.deleteSavepoint(id)
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  return (
    <section className="savepoints-panel">
      <header className="flex items-center justify-between mb-3">
        <h2>💾 Session Savepoints</h2>
        <button className="btn" onClick={() => setIsCreating(!isCreating)}>
          {isCreating ? 'Cancel' : '+ New Checkpoint'}
        </button>
      </header>

      {isCreating && (
        <form className="card p-3 mb-3 flex flex-col gap-2" onSubmit={handleCreate}>
          <h3 className="text-sm font-semibold m-0">Create Session Snapshot</h3>
          <input
            type="text"
            placeholder="Checkpoint Name (e.g. End of Session 3 / Before Boss Room)"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <textarea
            placeholder="Session notes (what happened, loot found, next plans)..."
            rows={3}
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
          />
          <button type="submit" className="btn-primary">
            Save Checkpoint
          </button>
        </form>
      )}

      <div className="search-row mb-3">
        <input
          type="search"
          placeholder="Search checkpoints..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      <div className="flex flex-col gap-2">
        {savepoints.length === 0 ? (
          <p className="hint text-xs text-muted">No savepoints recorded for this campaign.</p>
        ) : (
          savepoints.map((sp) => (
            <div key={sp.id} className="card p-3 flex flex-col gap-1.5 border border-border">
              <div className="flex items-center justify-between">
                <strong className="text-sm text-text">{sp.name}</strong>
                <span className="text-[10px] text-muted">{new Date(sp.created_at).toLocaleString()}</span>
              </div>
              {sp.notes && <p className="text-xs text-muted m-0 whitespace-pre-wrap">{sp.notes}</p>}
              <div className="flex items-center gap-2 mt-1">
                <button
                  className="btn text-xs hover:text-accent"
                  onClick={() => handleRestore(sp)}
                  title="Restore board and token vitals to this point"
                >
                  🔄 Restore State
                </button>
                <button
                  className="btn text-xs text-danger ml-auto"
                  onClick={() => handleDelete(sp.id)}
                  title="Delete savepoint"
                >
                  Delete
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </section>
  )
}
