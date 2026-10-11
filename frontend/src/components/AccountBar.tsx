import type { User } from '../api/types'

type Props = {
  user: User
  onLogout: () => void
}

export function AccountBar({ user, onLogout }: Props) {
  return (
    <div className="flex items-center justify-between px-3 py-2 bg-[#08090d] border-b-2 border-[#2b3145] text-xs">
      <div className="flex items-center gap-2">
        <div className="w-6 h-6 bg-[#1f5156] border-2 border-[#66fcf1] flex items-center justify-center text-[9px] font-bold text-[#66fcf1] shadow-[1px_1px_0px_#000]">
          {user.display_name.slice(0, 2).toUpperCase()}
        </div>
        <span className="font-bold text-[#e0e2ec] tracking-wider text-[11px]">{user.display_name}</span>
        <span className={`px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-widest border ${
          user.role === 'dm'
            ? 'bg-[#3b123d] text-[#e056fd] border-[#e056fd]'
            : 'bg-[#18233c] text-[#686de0] border-[#686de0]'
        }`}>
          [{user.role}]
        </span>
      </div>
      <button
        onClick={onLogout}
        className="text-[10px] text-[#ff7979] hover:text-[#eb4d4b] uppercase font-bold tracking-wider bg-transparent border-none cursor-pointer"
      >
        [EXIT]
      </button>
    </div>
  )
}
