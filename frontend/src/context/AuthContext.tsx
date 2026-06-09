import { createContext, useContext, useState, type ReactNode } from 'react'
import { setAccessToken, clearAccessToken } from '../api/client'

interface AuthState {
  userId: string | null
  refreshToken: string | null
  isAuthenticated: boolean
  signIn: (userId: string, accessToken: string, refreshToken: string) => void
  signOut: () => void
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [userId, setUserId] = useState<string | null>(null)
  const [refreshToken, setRefreshToken] = useState<string | null>(
    sessionStorage.getItem('refresh_token')
  )

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
