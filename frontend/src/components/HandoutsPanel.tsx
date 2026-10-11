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
    <section className="space-y-3 text-xs">
      <header className="flex items-center justify-between pb-1 border-b border-[#2b3145]">
        <h2 className="text-[10px] font-bold uppercase tracking-wider text-[#e0e2ec]">📜 Handouts & Lore</h2>
        {isDM && (
          <button
            className={`px-2 py-0.5 text-[9px] font-bold uppercase border-2 transition-all ${
              isCreating
                ? 'bg-[#2d1214] border-[#ff4757] text-[#ff4757]'
                : 'bg-[#1f5156] border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black'
            }`}
            onClick={() => setIsCreating(!isCreating)}
          >
            {isCreating ? 'CANCEL' : '+ NEW HANDOUT'}
          </button>
        )}
      </header>

      <input
        type="search"
        placeholder="SEARCH HANDOUTS & LORE..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="w-full bg-[#090a0f] border-2 border-[#2b3145] p-1.5 text-[10px] text-[#e0e2ec] placeholder-[#57606f] focus:outline-none focus:border-[#66fcf1]"
      />

      {isCreating && isDM && (
        <form onSubmit={handleCreate} className="p-3 bg-[#11131c] border border-indigo-500/50 rounded-lg space-y-3 shadow-lg">
          <div className="flex items-center justify-between pb-1 border-b border-[#232838]">
            <h3 className="text-xs font-bold text-indigo-300 m-0 uppercase tracking-wide">New Handout / Clue</h3>
          </div>

          <div className="space-y-1">
            <label className="text-[11px] font-semibold text-slate-300 block">Title *</label>
            <input
              type="text"
              placeholder="e.g. Ancient Map of Skull Keep"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
            />
          </div>

          <div className="space-y-1">
            <label className="text-[11px] font-semibold text-slate-300 block">Image URL (Optional)</label>
            <input
              type="url"
              placeholder="https://..."
              value={imageUrl}
              onChange={(e) => setImageUrl(e.target.value)}
            />
          </div>

          <div className="space-y-1">
            <label className="text-[11px] font-semibold text-slate-300 block">Content / Note</label>
            <textarea
              placeholder="Handout text, letter contents, poem, puzzle, or DM lore..."
              rows={4}
              value={content}
              onChange={(e) => setContent(e.target.value)}
            />
          </div>

          <label className="flex items-center gap-2 text-xs text-slate-300 cursor-pointer pt-1">
            <input
              type="checkbox"
              checked={isRevealed}
              onChange={(e) => setIsRevealed(e.target.checked)}
            />
            <span>Reveal immediately to all players</span>
          </label>

          <div className="flex gap-2 pt-1">
            <button
              type="submit"
              className="flex-1 py-2 bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs rounded transition-colors"
            >
              Save Handout
            </button>
            <button
              type="button"
              onClick={() => setIsCreating(false)}
              className="px-3 py-2 bg-[#1e2333] hover:bg-[#2b3145] text-slate-300 text-xs rounded transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      <div className="space-y-1.5">
        {handouts.length === 0 ? (
          <p className="text-[10px] text-[#57606f] italic text-center py-2">NO HANDOUTS AVAILABLE.</p>
        ) : (
          handouts.map((h) => (
            <div
              key={h.id}
              className={`p-2 bg-[#12141d] border-2 transition-all cursor-pointer ${
                selectedHandout?.id === h.id ? 'border-[#66fcf1]' : 'border-[#2b3145] hover:border-[#45a29e]'
              }`}
              onClick={() => setSelectedHandout(h)}
            >
              <div className="space-y-1">
                <div className="flex items-center justify-between">
                  <h4 className="m-0 font-bold text-xs text-[#e0e2ec] uppercase">{h.title}</h4>
                  {isDM && (
                    <span className={`text-[9px] px-1 py-0.2 border ${
                      h.is_revealed
                        ? 'bg-[#0e271a] border-[#2ed573] text-[#2ed573]'
                        : 'bg-[#29170e] border-[#ffa502] text-[#ffa502]'
                    }`}>
                      {h.is_revealed ? 'REVEALED' : 'SECRET'}
                    </span>
                  )}
                </div>
                <p className="m-0 text-[11px] text-slate-400 line-clamp-2">
                  {h.content || '(No text)'}
                </p>
              </div>

              {isDM && (
                <div className="flex items-center gap-2 pt-1.5 mt-1.5 border-t border-[#1e2333]" onClick={(e) => e.stopPropagation()}>
                  <button
                    className="text-[9px] text-slate-400 hover:text-[#66fcf1] uppercase bg-transparent border-none cursor-pointer"
                    onClick={() => handleToggleReveal(h)}
                  >
                    [{h.is_revealed ? 'HIDE' : 'REVEAL'}]
                  </button>
                  <button
                    className="text-[9px] text-[#ff4757] hover:underline uppercase ml-auto bg-transparent border-none cursor-pointer"
                    onClick={() => handleDelete(h.id)}
                  >
                    [DEL]
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
