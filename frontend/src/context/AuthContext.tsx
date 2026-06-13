import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import axios from 'axios'
import { setAccessToken, clearAccessToken, apiBase } from '../api/client'

interface AuthState {
  userId: string | null
  refreshToken: string | null
  isAuthenticated: boolean
  ready: boolean  // true once the initial token restore attempt has completed
  signIn: (userId: string, accessToken: string, refreshToken: string) => void
  signOut: () => void
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [userId, setUserId] = useState<string | null>(null)
  const [refreshToken, setRefreshToken] = useState<string | null>(
    sessionStorage.getItem('refresh_token')
  )
  // ready=false until we've tried to restore a session from sessionStorage
  const [ready, setReady] = useState(false)

  // On mount: if sessionStorage has a refresh token but memory has no access
  // token (e.g. after a page reload), silently exchange it for a new access token.
  useEffect(() => {
    const stored = sessionStorage.getItem('refresh_token')
    if (!stored) {
      setReady(true)
      return
    }
    const base = apiBase ? `${apiBase}/api/v1` : '/api/v1'
    axios.post(`${base}/auth/refresh`, { refresh_token: stored })
      .then(r => {
        const { access_token, refresh_token: newRT, user_id } = r.data
        setUserId(user_id ?? null)
        setAccessToken(access_token)
        setRefreshToken(newRT)
        sessionStorage.setItem('refresh_token', newRT)
      })
      .catch(() => {
        // Refresh token expired or revoked — clear everything
        clearAccessToken()
        setRefreshToken(null)
        sessionStorage.removeItem('refresh_token')
      })
      .finally(() => setReady(true))
  }, [])

  const signIn = (uid: string, accessToken: string, rt: string) => {
    setUserId(uid)
    setAccessToken(accessToken)
    setRefreshToken(rt)
    sessionStorage.setItem('refresh_token', rt)
  }

  const signOut = () => {
    setUserId(null)
    clearAccessToken()
    setRefreshToken(null)
    sessionStorage.removeItem('refresh_token')
  }

  return (
    <AuthContext.Provider value={{
      userId,
      refreshToken,
      isAuthenticated: !!refreshToken,
      ready,
      signIn,
      signOut,
    }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
