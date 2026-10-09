import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Handout } from '../api/types'

type Props = {
  campaignId: number | null
  isDM: boolean
  onError: (error: unknown) => void
}

export function HandoutsPanel({ campaignId, isDM, onError }: Props) {
  const [handouts, setHandouts] = useState<Handout[]>([])
  const [search, setSearch] = useState('')
  const [selectedHandout, setSelectedHandout] = useState<Handout | null>(null)
  const [isCreating, setIsCreating] = useState(false)

  // Form state
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [imageUrl, setImageUrl] = useState('')
  const [isRevealed, setIsRevealed] = useState(false)

  const refresh = useCallback(() => {
    if (!campaignId) return
    api
      .listHandouts(campaignId, { search, page_size: 50, sort_by: 'created_at', sort_dir: 'desc' })
      .then((page) => setHandouts(page.items))
      .catch(onError)
  }, [campaignId, search, onError])

  useEffect(() => {
    refresh()
  }, [refresh])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!campaignId || !title.trim()) return
    try {
      await api.createHandout(campaignId, {
        campaign_id: campaignId,
        title: title.trim(),
        content: content.trim(),
        image_url: imageUrl.trim(),
        is_revealed: isRevealed,
      })
      setTitle('')
      setContent('')
      setImageUrl('')
      setIsRevealed(false)
      setIsCreating(false)
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const handleToggleReveal = async (handout: Handout) => {
    try {
      const updated = await api.updateHandout(handout.id, {
        campaign_id: handout.campaign_id,
        title: handout.title,
        content: handout.content,
        image_url: handout.image_url,
        is_revealed: !handout.is_revealed,
      })
      if (selectedHandout?.id === handout.id) {
        setSelectedHandout(updated)
      }
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this handout?')) return
    try {
      await api.deleteHandout(id)
      if (selectedHandout?.id === id) setSelectedHandout(null)
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  return (
    <section className="handouts-panel">
      <header className="handouts-header">
        <h2>📜 Handouts & Lore</h2>
        {isDM && (
          <button className="btn-small" onClick={() => setIsCreating(!isCreating)}>
            {isCreating ? 'Cancel' : '+ New Handout'}
          </button>
        )}
      </header>

      <div className="search-row">
        <input
          type="search"
          placeholder="Search handouts & lore..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {isCreating && isDM && (
        <form className="handout-form card" onSubmit={handleCreate}>
          <h3>New Handout / Clue</h3>
          <input
            type="text"
            placeholder="Title (e.g. Ancient Map of Skull Keep)"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            required
          />
          <input
            type="url"
            placeholder="Image URL (optional)"
            value={imageUrl}
            onChange={(e) => setImageUrl(e.target.value)}
          />
          <textarea
            placeholder="Handout text, letter contents, poem, puzzle, or DM lore..."
            rows={4}
            value={content}
            onChange={(e) => setContent(e.target.value)}
          />
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={isRevealed}
              onChange={(e) => setIsRevealed(e.target.checked)}
            />
            Reveal immediately to all players
          </label>
          <button type="submit" className="primary">
            Save Handout
          </button>
        </form>
      )}

      <div className="handouts-list">
        {handouts.length === 0 ? (
          <p className="hint">No handouts available.</p>
        ) : (
          handouts.map((h) => (
            <div
              key={h.id}
              className={`handout-card ${selectedHandout?.id === h.id ? 'active' : ''}`}
              onClick={() => setSelectedHandout(h)}
            >
              <div className="handout-info">
                <h4>{h.title}</h4>
                <p className="handout-preview">
                  {h.content ? h.content.slice(0, 60) + (h.content.length > 60 ? '...' : '') : '(No text)'}
                </p>
                {isDM && (
                  <span className={`badge ${h.is_revealed ? 'badge-online' : 'badge-offline'}`}>
                    {h.is_revealed ? '👁️ Revealed to Party' : '🔒 Secret / DM only'}
                  </span>
                )}
              </div>
              {isDM && (
                <div className="handout-actions" onClick={(e) => e.stopPropagation()}>
                  <button
                    className="btn-small link"
                    onClick={() => handleToggleReveal(h)}
                    title={h.is_revealed ? 'Hide from players' : 'Reveal to players'}
                  >
                    {h.is_revealed ? 'Hide' : 'Reveal'}
                  </button>
                  <button
                    className="btn-small link text-danger"
                    onClick={() => handleDelete(h.id)}
                    title="Delete"
                  >
                    ×
                  </button>
                </div>
              )}
            </div>
          ))
        )}
      </div>

      {selectedHandout && (
        <div className="handout-modal-backdrop" onClick={() => setSelectedHandout(null)}>
          <div className="handout-modal card" onClick={(e) => e.stopPropagation()}>
            <header className="handout-modal-header">
              <h2>{selectedHandout.title}</h2>
              <button className="btn-close" onClick={() => setSelectedHandout(null)}>
                ×
              </button>
            </header>
            <div className="handout-modal-body">
              {selectedHandout.image_url && (
                <img
                  src={selectedHandout.image_url}
                  alt={selectedHandout.title}
                  className="handout-modal-img"
                />
              )}
              {selectedHandout.content && (
                <div className="handout-parchment">
                  <pre>{selectedHandout.content}</pre>
                </div>
              )}
            </div>
            {isDM && (
              <footer className="handout-modal-footer">
                <button className="secondary" onClick={() => handleToggleReveal(selectedHandout)}>
                  {selectedHandout.is_revealed ? '🔒 Hide from Players' : '👁️ Reveal to Table'}
                </button>
              </footer>
            )}
          </div>
        </div>
      )}
    </section>
  )
}
