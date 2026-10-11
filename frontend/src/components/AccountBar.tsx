import type { User } from '../api/types'

type Props = {
  user: User
  onLogout: () => void
}

export function AccountBar({ user, onLogout }: Props) {
  return (
    <div className="flex items-center justify-between px-3 py-2.5 bg-[#12141a] border-b border-[#232834] text-xs">
      <div className="flex items-center gap-2">
        <div className="w-6 h-6 rounded-full bg-indigo-900/60 border border-indigo-500/40 flex items-center justify-center text-[10px] font-bold text-indigo-200">
          {user.display_name.slice(0, 2).toUpperCase()}
        </div>
        <span className="font-semibold text-slate-200">{user.display_name}</span>
        <span className={`px-1.5 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider ${
          user.role === 'dm'
            ? 'bg-purple-900/50 text-purple-300 border border-purple-700/50'
            : 'bg-slate-800 text-slate-300 border border-slate-700'
        }`}>
          {user.role}
        </span>
      </div>
      <button
        onClick={onLogout}
        className="text-[11px] text-slate-400 hover:text-red-400 transition-colors bg-transparent border-none cursor-pointer"
      >
        Sign out
      </button>
    </div>
  )
}
