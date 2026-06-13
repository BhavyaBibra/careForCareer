import api from './client'

export interface Job {
  id: string
  title: string
  company: string
  location: string
  apply_url: string
  description: string
  posted_at?: string
  source: string
}

export interface JobSearchResponse {
  jobs: Job[]
  total: number
  query: string
  location: string
  based_on?: string
}

export function searchJobs(q: string, location: string, limit = 20) {
  return api.get<JobSearchResponse>('/jobs/search', {
    params: { q, location, limit },
  })
}

/** Returns jobs personalised to the authenticated candidate's tier and profile. */
export function getSuggestedJobs() {
  return api.get<JobSearchResponse>('/jobs/suggested')
}
