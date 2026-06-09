import api from './client'

export interface PivotRequest {
  target_role: string
  target_company?: string
  company_type?: string
  job_description?: string
  current_role?: string
  current_skills?: string[]
}

export interface PivotSkill {
  skill: string
  relevance: 'high' | 'medium' | 'low'
  comment: string
}

export interface PivotTimeline {
  optimistic: string
  realistic: string
  what_helps: string
  what_hurts: string
}

export interface EntryPath {
  name: string
  difficulty: 'high' | 'medium' | 'low'
  description: string
  examples: string[]
  best_for: string
}

export interface PivotWeek {
  phase: string
  weeks: string
  topics: string[]
  goal: string
}

export interface PivotResult {
  target_role: string
  pivot_difficulty: 'hard' | 'moderate' | 'easy'
  difficulty_label: string
  overall_fit: number
  summary: string
  transferable_skills: PivotSkill[]
  skills_to_learn: PivotSkill[]
  timeline: PivotTimeline
  entry_paths: EntryPath[]
  prep_plan_outline: PivotWeek[]
  day_in_the_life: string
  honest_caveats: string[]
  quick_wins: string[]
}

export function analysePivot(req: PivotRequest) {
  return api.post<PivotResult>('/pivot/analyse', req)
}
