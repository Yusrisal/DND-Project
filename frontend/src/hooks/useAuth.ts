import { useCallback, useEffect, useState } from 'react'
import { api, ApiError } from '../api/client'
import type { User } from '../api/types'

type AuthState = {
  user: User | null
  /** True until the first /auth/me answers, so the app does not flash the login form. */
  loading: boolean
}

/**
 * useAuth owns the session. The cookie itself is HttpOnly and invisible to this
 * code; /auth/me is how the app finds out whether it has one and what role it
 * carries.
 */
export function useAuth() {
  const [{ user, loading }, setState] = useState<AuthState>({ user: null, loading: true })

  useEffect(() => {
    api
      .me()
      .then((me) => setState({ user: me, loading: false }))
      .catch(() => setState({ user: null, loading: false }))
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    setState({ user: await api.login({ email, password }), loading: false })
  }, [])

  const register = useCallback(
    async (input: { code: string; email: string; display_name: string; password: string }) => {
      setState({ user: await api.register(input), loading: false })
    },
    [],
  )

  const logout = useCallback(async () => {
    await api.logout()
    setState({ user: null, loading: false })
  }, [])

  /** Called when any request comes back 401, so an expired session drops to the login form. */
  const onUnauthorized = useCallback((error: unknown) => {
    if (error instanceof ApiError && error.isUnauthorized) {
      setState({ user: null, loading: false })
      return true
    }
    return false
  }, [])

  return { user, loading, login, register, logout, onUnauthorized }
}
