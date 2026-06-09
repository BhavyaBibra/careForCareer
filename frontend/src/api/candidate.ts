import api from './client'

export interface Candidate {
  candidate_id: string
  user_id: string
  years_experience: number
  tier: number
  tier_label: string
  tier_explanation: string
  current_company: string
  current_comp_inr: number
  target_comp_inr: number
  created_at: string
  updated_at: string
}

export const getProfile = () => api.get<Candidate>('/candidate')

export const createProfile = (data: {
  years_experience: number
  current_company: string
  current_comp_inr: number
  target_comp_inr: number
}) => api.post<Candidate>('/candidate', data)

export const updateProfile = (data: {
  years_experience: number
  current_company: string
  current_comp_inr: number
  target_comp_inr: number
}) => api.put<Candidate>('/candidate', data)
