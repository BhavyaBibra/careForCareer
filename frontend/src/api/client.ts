import axios from 'axios'

// apiBase is the scheme+host of the backend (e.g. https://careforcareer.onrender.com).
// Empty string in local dev — relative URLs hit the Vite proxy.
export const apiBase = import.meta.env.VITE_API_URL ?? ''

const baseURL = apiBase ? `${apiBase}/api/v1` : '/api/v1'

const api = axios.create({ baseURL })

// Attach access token from memory on every request
api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// In-memory token storage (never localStorage — XSS safe)
let _accessToken = ''
export const setAccessToken = (t: string) => { _accessToken = t }
export const getAccessToken = () => _accessToken
export const clearAccessToken = () => { _accessToken = '' }

export default api
