import { useState } from 'react'
import type { GameMap } from '../api/types'

export type MapFormValues = {
  name: string
  image_url: string
  grid_size: number
  width: number
  height: number
}

type Props = {
  isOpen: boolean
  /** Pass a map to edit it; omit to create a new one. */
  map?: GameMap | null
  onClose: () => void
  onSubmit: (values: MapFormValues) => void
}

const DEFAULTS: MapFormValues = {
  name: '',
  image_url: '',
  grid_size: 50,
  width: 1920,
  height: 1080,
}

const PRESETS: { label: string; width: number; height: number }[] = [
  { label: 'Small 20x15', width: 1000, height: 750 },
  { label: 'Medium 38x21', width: 1920, height: 1080 },
  { label: 'Large 51x28', width: 2560, height: 1440 },
]

const field =
  'w-full bg-[#08090d] border-2 border-[#2b3145] p-2 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-[#66fcf1]'
const labelStyle = 'text-[10px] font-bold text-slate-300 block uppercase tracking-wider mb-1'

export function MapFormModal({ isOpen, map, onClose, onSubmit }: Props) {
  const [values, setValues] = useState<MapFormValues>(
    map
      ? {
          name: map.name,
          image_url: map.image_url,
          grid_size: map.grid_size,
          width: map.width,
          height: map.height,
        }
      : DEFAULTS,
  )
  // A broken URL should say so rather than silently render an empty board.
  const [imageBroken, setImageBroken] = useState(false)

  if (!isOpen) return null

  const set = <K extends keyof MapFormValues>(key: K, value: MapFormValues[K]) =>
    setValues((current) => ({ ...current, [key]: value }))

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    const name = values.name.trim()
    if (!name) return
    onSubmit({
      name,
      image_url: values.image_url.trim(),
      grid_size: Math.max(10, values.grid_size || 50),
      width: Math.max(100, values.width || 1920),
      height: Math.max(100, values.height || 1080),
    })
    onClose()
  }

  const squaresAcross = Math.floor((values.width || 0) / (values.grid_size || 50))
  const squaresDown = Math.floor((values.height || 0) / (values.grid_size || 50))

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
            🗺️ {map ? 'Edit Map' : 'New Map'}
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
            <label className={labelStyle}>Map Name</label>
            <input
              autoFocus
              value={values.name}
              onChange={(event) => set('name', event.target.value)}
              placeholder="e.g. Tavern Cellar"
              required
              className={field}
            />
          </div>

          <div>
            <label className={labelStyle}>Background Image URL</label>
            <input
              value={values.image_url}
              onChange={(event) => {
                setImageBroken(false)
                set('image_url', event.target.value)
              }}
              placeholder="https://… (leave blank for a plain grid)"
              className={field}
            />
          </div>

          {/* The preview is the point: the DM sees the board before saving. */}
          <div className="h-37.5 bg-[#08090d] border-2 border-[#2b3145] flex items-center justify-center overflow-hidden relative">
            {values.image_url.trim() && !imageBroken ? (
              <img
                src={values.image_url.trim()}
                alt="Map preview"
                className="w-full h-full object-contain"
                onError={() => setImageBroken(true)}
              />
            ) : (
              <div
                className="w-full h-full"
                style={{
                  backgroundImage:
                    'linear-gradient(#ffffff14 1px, transparent 1px), linear-gradient(90deg, #ffffff14 1px, transparent 1px)',
                  backgroundSize: '16px 16px',
                }}
              />
            )}
            {imageBroken && (
              <span className="absolute text-[10px] font-bold text-[#ff6b81] uppercase bg-black/70 px-2 py-1">
                Image failed to load
              </span>
            )}
          </div>

          <div className="grid grid-cols-3 gap-2">
            <div>
              <label className={labelStyle}>Grid px</label>
              <input
                type="number"
                min={10}
                step={5}
                value={values.grid_size}
                onChange={(event) => set('grid_size', parseInt(event.target.value, 10) || 0)}
                className={field}
              />
            </div>
            <div>
              <label className={labelStyle}>Width px</label>
              <input
                type="number"
                min={100}
                step={10}
                value={values.width}
                onChange={(event) => set('width', parseInt(event.target.value, 10) || 0)}
                className={field}
              />
            </div>
            <div>
              <label className={labelStyle}>Height px</label>
              <input
                type="number"
                min={100}
                step={10}
                value={values.height}
                onChange={(event) => set('height', parseInt(event.target.value, 10) || 0)}
                className={field}
              />
            </div>
          </div>

          <p className="text-[10px] text-[#747d8c] m-0">
            Board is {squaresAcross} × {squaresDown} squares ({squaresAcross * 5} ft ×{' '}
            {squaresDown * 5} ft at 5 ft per square).
          </p>

          <div className="flex flex-wrap gap-1.5">
            {PRESETS.map((preset) => (
              <button
                key={preset.label}
                type="button"
                onClick={() => {
                  set('width', preset.width)
                  set('height', preset.height)
                }}
                className="px-2 py-1 text-[9px] font-bold uppercase bg-[#1e2333] border-2 border-[#32394d] text-slate-300 hover:border-[#66fcf1] hover:text-white"
              >
                {preset.label}
              </button>
            ))}
          </div>

          <div className="flex gap-2 pt-1">
            <button
              type="submit"
              className="flex-1 py-2 bg-[#1f5156] border-2 border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black font-bold text-xs uppercase shadow-[2px_2px_0px_#000] transition-colors"
            >
              {map ? 'Save Map' : 'Create Map'}
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
