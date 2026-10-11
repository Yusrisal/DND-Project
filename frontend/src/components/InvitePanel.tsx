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
    <section className="space-y-3 pt-3 border-t-2 border-[#2b3145]">
      <header className="flex items-center justify-between">
        <h2 className="text-[10px] font-bold uppercase tracking-wider text-[#e0e2ec]">Invites</h2>
        <div className="flex gap-1.5">
          <button
            onClick={() => create('player')}
            className="px-2 py-1 text-[9px] font-bold bg-[#1f5156] border-2 border-[#66fcf1] text-[#66fcf1] hover:bg-[#45a29e] hover:text-black uppercase shadow-[1px_1px_0px_#000]"
          >
            + Player
          </button>
          <button
            onClick={() => create('dm')}
            className="px-2 py-1 text-[9px] font-bold bg-[#3b123d] border-2 border-[#e056fd] text-[#e056fd] hover:bg-[#be2edd] hover:text-white uppercase shadow-[1px_1px_0px_#000]"
          >
            + DM
          </button>
        </div>
      </header>

      {open.length === 0 && <p className="text-[10px] text-[#57606f] italic py-1">No unused invite codes.</p>}

      <div className="space-y-1.5">
        {open.map((invite) => (
          <div
            key={invite.code}
            className="flex items-center justify-between p-2 bg-[#12141d] border-2 border-[#2b3145] text-xs"
          >
            <div className="flex items-center gap-2">
              <code className="text-[#66fcf1] font-mono font-bold text-[11px] bg-black px-1.5 py-0.5 border border-[#232838]">
                {invite.code}
              </code>
              <span
                className={`text-[9px] px-1 py-0.5 font-bold uppercase border ${
                  invite.role === 'dm'
                    ? 'bg-[#3b123d] text-[#e056fd] border-[#e056fd]'
                    : 'bg-[#18233c] text-[#686de0] border-[#686de0]'
                }`}
              >
                {invite.role}
              </span>
            </div>

            <div className="flex gap-1.5">
              <button
                onClick={() => copy(invite.code)}
                className="px-2 py-0.5 bg-[#1e272e] border border-[#66fcf1] text-[#66fcf1] hover:bg-[#66fcf1] hover:text-black text-[9px] font-bold uppercase"
              >
                {copied === invite.code ? 'COPIED!' : 'COPY'}
              </button>
              <button
                onClick={() => revoke(invite.code)}
                className="px-2 py-0.5 bg-[#2d1214] border border-[#ff4757] text-[#ff4757] hover:bg-[#ff4757] hover:text-white text-[9px] font-bold uppercase"
              >
                REVOKE
              </button>
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
