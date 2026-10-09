import { useCallback, useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Invite, Role } from '../api/types'

type Props = {
  onError: (error: unknown) => void
}

/** InvitePanel is DM-only: it mints the codes new players register with. */
export function InvitePanel({ onError }: Props) {
  const [invites, setInvites] = useState<Invite[]>([])
  const [copied, setCopied] = useState<string | null>(null)

  const refresh = useCallback(() => {
    api
      .listInvites()
      .then((page) => setInvites(page.items))
      .catch(onError)
  }, [onError])

  useEffect(refresh, [refresh])

  const create = async (role: Role) => {
    try {
      await api.createInvite(role)
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const revoke = async (code: string) => {
    try {
      await api.deleteInvite(code)
      refresh()
    } catch (err) {
      onError(err)
    }
  }

  const copy = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code)
      setCopied(code)
      setTimeout(() => setCopied(null), 1500)
    } catch {
      // Clipboard access can be refused; the code is on screen either way.
    }
  }

  const open = invites.filter((invite) => invite.used_by === null)

  return (
    <section>
      <header>
        <h2>Invites</h2>
        <span className="actions">
          <button onClick={() => create('player')}>+ player</button>
          <button onClick={() => create('dm')}>+ dm</button>
        </span>
      </header>

      {open.length === 0 && <p className="hint">No unused codes.</p>}

      <ul>
        {open.map((invite) => (
          <li key={invite.code} className="invite-row">
            <code>{invite.code}</code>
            <span className={invite.role === 'dm' ? 'badge dm' : 'badge'}>{invite.role}</span>
            <button className="link" onClick={() => copy(invite.code)}>
              {copied === invite.code ? 'copied' : 'copy'}
            </button>
            <button className="link" onClick={() => revoke(invite.code)}>
              revoke
            </button>
          </li>
        ))}
      </ul>
    </section>
  )
}
