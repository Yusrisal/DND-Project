import type { User } from '../api/types'

type Props = {
  user: User
  onLogout: () => void
}

export function AccountBar({ user, onLogout }: Props) {
  return (
    <div className="account">
      <div>
        <strong>{user.display_name}</strong>
        <span className={user.role === 'dm' ? 'badge dm' : 'badge'}>{user.role}</span>
      </div>
      <button className="link" onClick={onLogout}>
        sign out
      </button>
    </div>
  )
}
