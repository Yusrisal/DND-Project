import { useState } from 'react'

type Props = {
  isOpen: boolean
  title: string
  label?: string
  placeholder?: string
  defaultValue?: string
  submitText?: string
  onClose: () => void
  onSubmit: (value: string) => void
}

export function InputModal({
  isOpen,
  title,
  label,
  placeholder,
  defaultValue = '',
  submitText = 'Confirm',
  onClose,
  onSubmit,
}: Props) {
  const [val, setVal] = useState(defaultValue)

  if (!isOpen) return null

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!val.trim()) return
    onSubmit(val.trim())
    setVal('')
    onClose()
  }

  return (
    <div
      className="fixed inset-0 bg-black/75 backdrop-blur-xs flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-[380px] bg-[#12141c] border-2 border-[#2b3145] p-5 shadow-[0_15px_40px_rgba(0,0,0,0.8)] space-y-4"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between pb-2 border-b border-[#232838]">
          <h3 className="text-xs font-bold uppercase text-[#66fcf1] m-0 tracking-wider">
            ⚔️ {title}
          </h3>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-white bg-transparent border-none text-sm cursor-pointer"
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3">
          {label && (
            <label className="text-[11px] font-bold text-slate-300 block uppercase">
              {label}
            </label>
          )}
          <input
            autoFocus
            type="text"
            placeholder={placeholder}
            value={val}
            onChange={(e) => setVal(e.target.value)}
            required
            className="w-full bg-[#08090d] border-2 border-[#2b3145] p-2 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-[#66fcf1]"
          />

          <div className="flex gap-2 pt-1">
            <button
              type="submit"
              className="flex-1 py-2 bg-[#1f5156] border-2 border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black font-bold text-xs uppercase shadow-[2px_2px_0px_#000] transition-colors"
            >
              {submitText}
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
