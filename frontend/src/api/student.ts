import axios from 'axios'

// Student track uses a separate axios instance — no auth token needed
const baseURL = import.meta.env.VITE_API_URL
  ? `${import.meta.env.VITE_API_URL}/api/v1`
  : '/api/v1'

const guestApi = axios.create({ baseURL })

export interface StudentAssessRequest {
  academic_status: string
  branch: string
  target_role: string
  target_role_custom?: string
  programming_level: string
  projects_built: string[]
  dsa_level: string
  hours_per_day: string
  first_milestone: string
}

export interface DomainReadiness {
  domain: string
  score: number
  status: 'strong' | 'partial' | 'missing'
  comment: string
  weeks_needed: number
}

export interface StudentPhase {
  phase: string
  weeks: string
  goal: string
  milestones: string[]
}

export interface RoadmapResource {
  name: string
  type: 'course' | 'book' | 'platform' | 'channel'
  url: string
  free: boolean
  description: string
}

export interface StudentRoadmap {
  learn_resources: RoadmapResource[]
  projects_to_build: string[]
  community: string[]
  first_job_targets: string[]
  portfolio_advice: string
}

export interface ChancesAssessment {
  internship_chance: string
  first_job_chance: string
  top_blocker: string
  wild_card: string
}

export interface StudentAssessmentResult {
  target_role: string
  honesty_note: string
  domain_readiness: DomainReadiness[]
  overall_readiness: number
  timeline_phases: StudentPhase[]
  earliest_hireable: string
  comfortable_target: string
  roadmap: StudentRoadmap
  chances_assessment: ChancesAssessment
  quick_start_today: string[]
  honest_caveats: string[]
}

export function assessStudent(req: StudentAssessRequest) {
  return guestApi.post<StudentAssessmentResult>('/student/assess', req)
}
