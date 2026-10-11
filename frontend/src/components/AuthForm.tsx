import { useState } from 'react'
import { ApiError } from '../api/client'

type Props = {
  onLogin: (email: string, password: string) => Promise<void>
  onRegister: (input: {
    code: string
    email: string
    display_name: string
    password: string
  }) => Promise<void>
}

type Mode = 'login' | 'register'

export function AuthForm({ onLogin, onRegister }: Props) {
  const [mode, setMode] = useState<Mode>('login')
  const [code, setCode] = useState('')
  const [email, setEmail] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  const submit = async (event: React.FormEvent) => {
    event.preventDefault()
    setError(null)
    setBusy(true)
    try {
      if (mode === 'login') {
        await onLogin(email, password)
      } else {
        await onRegister({ code, email, display_name: displayName, password })
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Something went wrong')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="min-h-screen w-full flex items-center justify-center bg-[#0d0f14] bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-[#1e2330] via-[#0d0f14] to-[#08090c] p-4 sm:p-6 text-[#e7e8ea]">
      <form
        className="w-full max-w-[420px] bg-[#16181f]/90 backdrop-blur-md border border-[#2b313e] shadow-[0_20px_50px_rgba(0,0,0,0.7),0_0_0_1px_rgba(108,140,255,0.08)] rounded-xl p-6 sm:p-8 flex flex-col gap-4"
        onSubmit={submit}
      >
        <div className="text-center space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-white bg-gradient-to-r from-white via-slate-200 to-indigo-300 bg-clip-text text-transparent">
            DnD Table
          </h1>
          <p className="text-xs text-slate-400">Virtual Tabletop & Campaign Manager</p>
        </div>

        <div className="flex bg-[#0f1116] p-1 rounded-lg border border-[#232733] gap-1">
          <button
            type="button"
            className={`flex-1 py-1.5 px-3 text-xs font-semibold rounded-md transition-all ${
              mode === 'login'
                ? 'bg-[#252a37] text-white shadow-sm shadow-black/50'
                : 'text-slate-400 hover:text-slate-200 hover:bg-[#1a1d26]'
            }`}
            onClick={() => setMode('login')}
          >
            Sign In
          </button>
          <button
            type="button"
            className={`flex-1 py-1.5 px-3 text-xs font-semibold rounded-md transition-all ${
              mode === 'register'
                ? 'bg-[#252a37] text-white shadow-sm shadow-black/50'
                : 'text-slate-400 hover:text-slate-200 hover:bg-[#1a1d26]'
            }`}
            onClick={() => setMode('register')}
          >
            Use an Invite
          </button>
        </div>

        {mode === 'register' && (
          <div className="space-y-3">
            <label className="flex flex-col gap-1.5 text-xs font-medium text-slate-300">
              Invite code
              <input
                value={code}
                onChange={(event) => setCode(event.target.value.trim())}
                autoComplete="off"
                placeholder="Leave blank for initial DM account"
                className="w-full bg-[#0d0f14] border border-[#282d38] rounded-md px-3 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors"
              />
            </label>
            <label className="flex flex-col gap-1.5 text-xs font-medium text-slate-300">
              Display name
              <input
                value={displayName}
                onChange={(event) => setDisplayName(event.target.value)}
                autoComplete="nickname"
                required
                className="w-full bg-[#0d0f14] border border-[#282d38] rounded-md px-3 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors"
              />
            </label>
          </div>
        )}

        <div className="space-y-3">
          <label className="flex flex-col gap-1.5 text-xs font-medium text-slate-300">
            Email
            <input
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              autoComplete="username"
              required
              className="w-full bg-[#0d0f14] border border-[#282d38] rounded-md px-3 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors"
            />
          </label>

          <label className="flex flex-col gap-1.5 text-xs font-medium text-slate-300">
            Password
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              minLength={mode === 'register' ? 10 : undefined}
              required
              className="w-full bg-[#0d0f14] border border-[#282d38] rounded-md px-3 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-colors"
            />
          </label>
        </div>

        {mode === 'register' && (
          <p className="text-[11px] text-slate-400">Password must be at least 10 characters.</p>
        )}

        {error && (
          <div className="bg-red-500/10 border border-red-500/30 text-red-300 text-xs rounded-md px-3 py-2">
            {error}
          </div>
        )}

        <button
          type="submit"
          className="mt-2 w-full py-2.5 px-4 bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white font-medium text-sm rounded-md shadow-md shadow-indigo-600/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          disabled={busy}
        >
          {busy ? 'Working…' : mode === 'login' ? 'Sign in' : 'Create account'}
        </button>
      </form>
    </div>
  )
}
