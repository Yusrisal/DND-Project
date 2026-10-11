import { useRef, useState } from 'react'
import { api } from '../api/client'
import type { Character, Token, User } from '../api/types'

export type TokenFormValues = {
  name: string
  image_url: string
  color: string
  size: number
  owner_user_id: number | null
  character_id: number | null
  hp_current: number
  hp_max: number
  armor_class: number
}

type Props = {
  isOpen: boolean
  /** Pass a token to edit it; omit to create a new one. */
  token?: Token | null
  players: User[]
  characters: Character[]
  defaultColor: string
  onClose: () => void
  onSubmit: (values: TokenFormValues) => void
}

const SIZES: { label: string; size: number }[] = [
  { label: 'Medium', size: 1 },
  { label: 'Large', size: 2 },
  { label: 'Huge', size: 3 },
  { label: 'Gargantuan', size: 4 },
]

const field =
  'w-full bg-[#08090d] border-2 border-[#2b3145] p-2 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-[#66fcf1]'
const labelStyle = 'text-[10px] font-bold text-slate-300 block uppercase tracking-wider mb-1'

export function TokenFormModal({
  isOpen,
  token,
  players,
  characters,
  defaultColor,
  onClose,
  onSubmit,
}: Props) {
  const [values, setValues] = useState<TokenFormValues>(
    token
      ? {
          name: token.name,
          image_url: token.image_url,
          color: token.color,
          size: token.size,
          owner_user_id: token.owner_user_id,
          character_id: token.character_id,
          hp_current: token.hp_current,
          hp_max: token.hp_max,
          armor_class: token.armor_class,
        }
      : {
          name: '',
          image_url: '',
          color: defaultColor,
          size: 1,
          owner_user_id: null,
          character_id: null,
          hp_current: 0,
          hp_max: 0,
          armor_class: 10,
        },
  )
  const [uploading, setUploading] = useState(false)
  const [uploadError, setUploadError] = useState('')
  const fileInput = useRef<HTMLInputElement>(null)

  if (!isOpen) return null

  const set = <K extends keyof TokenFormValues>(key: K, value: TokenFormValues[K]) =>
    setValues((current) => ({ ...current, [key]: value }))

  // Linking a sheet is the common case, so it fills in the blanks: the name and
  // the owner almost always follow from it, and the sheet takes over the vitals.
  const linkCharacter = (id: number | null) => {
    const sheet = characters.find((c) => c.id === id)
    setValues((current) => ({
      ...current,
      character_id: id,
      name: sheet && !current.name.trim() ? sheet.name : current.name,
      owner_user_id: sheet?.user_id ?? current.owner_user_id,
    }))
  }

  const handleFile = async (file: File | undefined) => {
    if (!file) return
    setUploading(true)
    setUploadError('')
    try {
      const { url } = await api.uploadImage(file)
      set('image_url', url)
    } catch (error) {
      setUploadError(error instanceof Error ? error.message : 'Upload failed')
    } finally {
      setUploading(false)
      if (fileInput.current) fileInput.current.value = ''
    }
  }

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    const name = values.name.trim()
    if (!name) return
    onSubmit({
      ...values,
      name,
      image_url: values.image_url.trim(),
      size: Math.max(1, values.size || 1),
    })
    onClose()
  }

  const linked = values.character_id !== null

  return (
    <div
      className="fixed inset-0 bg-black/75 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-115 max-h-[90vh] overflow-y-auto bg-[#12141c] border-2 border-[#2b3145] p-5 shadow-[0_15px_40px_rgba(0,0,0,0.8)] space-y-4"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="flex items-center justify-between pb-2 border-b border-[#232838]">
          <h3 className="text-xs font-bold uppercase text-[#66fcf1] m-0 tracking-wider">
            ♟ {token ? 'Edit Token' : 'New Token'}
          </h3>
          <button
            type="button"
            onClick={onClose}
            className="text-slate-400 hover:text-white bg-transparent border-none text-sm"
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3">
          <div>
            <label className={labelStyle}>Character Sheet</label>
            <select
              value={values.character_id ?? ''}
              onChange={(event) =>
                linkCharacter(event.target.value ? Number(event.target.value) : null)
              }
              className={field}
            >
              <option value="">None — a monster or a prop</option>
              {characters.map((character) => (
                <option key={character.id} value={character.id}>
                  {character.name} (level {character.level} {character.class || 'adventurer'})
                </option>
              ))}
            </select>
            <p className="text-[10px] text-[#747d8c] mt-1 mb-0">
              {linked
                ? 'Damage and healing go to the sheet, and the health bar follows it.'
                : 'This token keeps its own hit points.'}
            </p>
          </div>

          <div>
            <label className={labelStyle}>Controlled By</label>
            <select
              value={values.owner_user_id ?? ''}
              onChange={(event) =>
                set('owner_user_id', event.target.value ? Number(event.target.value) : null)
              }
              className={field}
            >
              <option value="">Nobody — only the DM may move it</option>
              {players.map((player) => (
                <option key={player.id} value={player.id}>
                  {player.display_name} ({player.role})
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className={labelStyle}>Token Name</label>
            <input
              autoFocus
              value={values.name}
              onChange={(event) => set('name', event.target.value)}
              placeholder="e.g. Goblin Archer #1"
              required
              className={field}
            />
          </div>

          <div>
            <label className={labelStyle}>Portrait</label>
            <div className="flex gap-2">
              <input
                value={values.image_url}
                onChange={(event) => set('image_url', event.target.value)}
                placeholder="https://… or upload a file"
                className={field}
              />
              <button
                type="button"
                onClick={() => fileInput.current?.click()}
                disabled={uploading}
                className="shrink-0 px-3 bg-[#1e2333] border-2 border-[#32394d] text-slate-300 hover:border-[#66fcf1] hover:text-white text-[10px] font-bold uppercase disabled:opacity-50"
              >
                {uploading ? 'Uploading…' : 'Upload'}
              </button>
            </div>
            <input
              ref={fileInput}
              type="file"
              accept="image/png,image/jpeg,image/gif,image/webp"
              hidden
              onChange={(event) => void handleFile(event.target.files?.[0])}
            />
            {uploadError && <p className="text-[10px] text-[#ff6b81] mt-1 mb-0">{uploadError}</p>}
          </div>

          <div className="grid grid-cols-2 gap-2">
            <div>
              <label className={labelStyle}>Colour</label>
              <input
                type="color"
                value={values.color}
                onChange={(event) => set('color', event.target.value)}
                className="w-full h-9 bg-[#08090d] border-2 border-[#2b3145] p-1"
              />
            </div>
            <div>
              <label className={labelStyle}>Size</label>
              <select
                value={values.size}
                onChange={(event) => set('size', Number(event.target.value))}
                className={field}
              >
                {SIZES.map((option) => (
                  <option key={option.size} value={option.size}>
                    {option.label} ({option.size}×{option.size})
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* A linked token reads its vitals off the sheet, so editing them here
              would be a lie. */}
          {!linked && (
            <div className="grid grid-cols-3 gap-2">
              <div>
                <label className={labelStyle}>HP Now</label>
                <input
                  type="number"
                  min={0}
                  value={values.hp_current}
                  onChange={(event) => set('hp_current', parseInt(event.target.value, 10) || 0)}
                  className={field}
                />
              </div>
              <div>
                <label className={labelStyle}>HP Max</label>
                <input
                  type="number"
                  min={0}
                  value={values.hp_max}
                  onChange={(event) => {
                    const max = parseInt(event.target.value, 10) || 0
                    setValues((current) => ({
                      ...current,
                      hp_max: max,
                      // A fresh token starts at full health, which is almost
                      // always what you want and never worth two fields.
                      hp_current: token ? current.hp_current : max,
                    }))
                  }}
                  className={field}
                />
              </div>
              <div>
                <label className={labelStyle}>Armour Class</label>
                <input
                  type="number"
                  min={0}
                  value={values.armor_class}
                  onChange={(event) => set('armor_class', parseInt(event.target.value, 10) || 0)}
                  className={field}
                />
              </div>
            </div>
          )}

          <div className="flex gap-2 pt-1">
            <button
              type="submit"
              className="flex-1 py-2 bg-[#1f5156] border-2 border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black font-bold text-xs uppercase shadow-[2px_2px_0px_#000] transition-colors"
            >
              {token ? 'Save Token' : 'Create Token'}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="px-3 py-2 bg-[#1e2333] border-2 border-[#32394d] text-slate-300 hover:text-white text-xs uppercase transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
