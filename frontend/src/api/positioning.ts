import api from './client'

export interface SkillSignal {
  skill: string
  level: 'strong' | 'partial' | 'missing'
  comment: string
}

export interface ActionItem {
  priority: 'critical' | 'high' | 'medium'
  title: string
  detail: string
  resource?: string
}

export interface PositioningResult {
  overall_match: number
  tier_fit: 'below' | 'match' | 'above'
  tier_fit_label: string
  company_bar: 'high' | 'medium' | 'accessible'
  company_bar_label: string
  summary: string
  skill_matches: SkillSignal[]
  skill_gaps: SkillSignal[]
  action_plan: ActionItem[]
  interview_focus: string[]
  time_to_ready: string
  confidence: 'high' | 'medium' | 'low'
}

export interface PositionRequest {
  job_title: string
  company?: string
  location?: string
  job_description: string
  declared_skills?: string[]
}

export function analysePosition(req: PositionRequest) {
  return api.post<PositioningResult>('/jobs/position', req)
}
