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
    <div className="auth">
      <form className="auth-card" onSubmit={submit}>
        <h1>DnD Table</h1>

        <div className="tabs">
          <button
            type="button"
            className={mode === 'login' ? 'tab active' : 'tab'}
            onClick={() => setMode('login')}
          >
            Sign in
          </button>
          <button
            type="button"
            className={mode === 'register' ? 'tab active' : 'tab'}
            onClick={() => setMode('register')}
          >
            Use an invite
          </button>
        </div>

        {mode === 'register' && (
          <>
            <label>
              Invite code
              <input
                value={code}
                onChange={(event) => setCode(event.target.value.trim())}
                autoComplete="off"
                placeholder="Leave empty for the very first account"
              />
            </label>
            <label>
              Display name
              <input
                value={displayName}
                onChange={(event) => setDisplayName(event.target.value)}
                autoComplete="nickname"
                required
              />
            </label>
          </>
        )}

        <label>
          Email
          <input
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            autoComplete="username"
            required
          />
        </label>

        <label>
          Password
          <input
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
            minLength={mode === 'register' ? 10 : undefined}
            required
          />
        </label>

        {mode === 'register' && <p className="hint">At least 10 characters.</p>}
        {error && <p className="error">{error}</p>}

        <button type="submit" className="primary" disabled={busy}>
          {busy ? 'Working…' : mode === 'login' ? 'Sign in' : 'Create account'}
        </button>
      </form>
    </div>
  )
}
