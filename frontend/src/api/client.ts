import axios from 'axios'

// apiBase is the scheme+host of the backend (e.g. https://careforcareer.onrender.com).
// Empty string in local dev — relative URLs hit the Vite proxy.
export const apiBase = import.meta.env.VITE_API_URL ?? ''

const baseURL = apiBase ? `${apiBase}/api/v1` : '/api/v1'

const api = axios.create({ baseURL })

// In-memory token storage (never localStorage — XSS safe)
let _accessToken = ''
export const setAccessToken = (t: string) => { _accessToken = t }
export const getAccessToken = () => _accessToken
export const clearAccessToken = () => { _accessToken = '' }

// Attach access token from memory on every request
api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// On 401: try once to refresh, then replay the original request.
// If refresh fails, clear session so the user is prompted to log in again.
let _refreshing: Promise<string> | null = null

api.interceptors.response.use(
  res => res,
  async (error) => {
    const original = error.config
    if (error.response?.status !== 401 || original._retried) {
      return Promise.reject(error)
    }
    original._retried = true

    const rt = sessionStorage.getItem('refresh_token')
    if (!rt) return Promise.reject(error)

    if (!_refreshing) {
      _refreshing = axios.post(`${baseURL}/auth/refresh`, { refresh_token: rt })
        .then(r => {
          const { access_token, refresh_token: newRT } = r.data
          setAccessToken(access_token)
          sessionStorage.setItem('refresh_token', newRT)
          return access_token
        })
        .catch(e => {
          clearAccessToken()
          sessionStorage.removeItem('refresh_token')
          throw e
        })
        .finally(() => { _refreshing = null })
    }

    try {
      const newToken = await _refreshing
      original.headers.Authorization = `Bearer ${newToken}`
      return api(original)
    } catch {
      return Promise.reject(error)
    }
  }
)

export default api
