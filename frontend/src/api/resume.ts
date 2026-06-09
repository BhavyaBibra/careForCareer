import api from './client'

export interface Resume {
  resume_id: string
  candidate_id: string
  version: number
  source_type: string
  extraction_status: string
  download_url: string
  storage_key: string
  created_at: string
}

export const uploadResume = (file: File) => {
  const form = new FormData()
  form.append('resume', file)
  return api.post<Resume>('/resumes', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export const getResume = (id: string) => api.get<Resume>(`/resumes/${id}`)
