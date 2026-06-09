import axios from 'axios'

const api = axios.create({ baseURL: '/api/v1' })

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
