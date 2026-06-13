import { useState } from 'react'
import { analysePosition, type PositioningResult, type ActionItem, type SkillSignal } from '../api/positioning'
import type { Job } from '../api/jobs'
import PrepPlanPanel from './PrepPlanPanel'

interface Props {
  job: Job
  /** Optional: declared skills from candidate profile */
  declaredSkills?: string[]
  /** If true, show a compact version (for onboarding step) */
  compact?: boolean
  /** Candidate YOE for prep plan calibration */
  yoe?: number
}

export default function PositioningPanel({ job, declaredSkills, compact, yoe = 0 }: Props) {
  const [result, setResult] = useState<PositioningResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const run = async () => {
    if (!job.description) {
      setError('No job description available for this listing.')
      return
    }
    setLoading(true)
    setError('')
    try {
      const res = await analysePosition({
        job_title: job.title,
        company: job.company,
        location: job.location,
        job_description: job.description,
        declared_skills: declaredSkills,
      })
      const data = res.data as any
      if (data?.parse_error) {
        setError('Analysis returned unexpected format — please try again.')
        return
      }
      setResult(res.data)
    } catch (err: any) {
      setError(err.response?.data?.error?.message ?? 'Analysis failed. Try again.')
    } finally {
      setLoading(false)
    }
  }

  if (!result) {
    return (
      <div className="bg-gray-800/60 border border-gray-700 rounded-xl p-4 space-y-3">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-sm font-medium text-white">See your positioning</p>
            <p className="text-xs text-gray-400 mt-0.5">
              How strong are you for <span className="text-indigo-300">{job.title}</span> at{' '}
              <span className="text-indigo-300">{job.company || 'this company'}</span>?
            </p>
          </div>
          <button
            onClick={run}
            disabled={loading || !job.description}
            className="shrink-0 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white text-xs font-semibold px-4 py-2 rounded-lg transition-colors"
          >
            {loading ? (
              <span className="flex items-center gap-1.5">
                <span className="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                Analysing…
              </span>
            ) : (
              'Analyse fit →'
            )}
          </button>
        </div>
        {!job.description && (
          <p className="text-xs text-yellow-500/80">No JD available for this listing — click "Analyse this role" to use it in the dashboard instead.</p>
        )}
        {error && <p className="text-xs text-red-400">{error}</p>}
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* ── Score header ──────────────────────────────────────────────── */}
      <div className="bg-gray-800 rounded-xl p-5 border border-gray-700">
        <div className="flex items-center justify-between mb-4">
          <div>
            <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">Overall match</p>
            <div className="flex items-baseline gap-2">
              <span className={`text-4xl font-bold ${matchColor(result.overall_match)}`}>
                {result.overall_match}%
              </span>
              <span className="text-gray-500 text-sm">fit score</span>
            </div>
          </div>
          <div className="text-right space-y-1.5">
            <TierFitBadge fit={result.tier_fit} />
            <CompanyBarBadge bar={result.company_bar} />
          </div>
        </div>

        {/* Score bar */}
        <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all duration-700 ${matchBarColor(result.overall_match)}`}
            style={{ width: `${result.overall_match}%` }}
          />
        </div>

        <p className="text-gray-300 text-sm mt-3 leading-relaxed">{result.summary}</p>
      </div>

      {!compact && (
        <>
          {/* ── Tier + company context ───────────────────────────────────── */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <InfoCard label="Tier fit" value={result.tier_fit_label} icon="🎯" />
            <InfoCard label="Company bar" value={result.company_bar_label} icon="🏢" />
          </div>

          {/* ── Skills ──────────────────────────────────────────────────── */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {(result.skill_matches?.length ?? 0) > 0 && (
              <SkillBlock title="✅ Your strengths" skills={result.skill_matches} />
            )}
            {(result.skill_gaps?.length ?? 0) > 0 && (
              <SkillBlock title="⚠️ Gaps to close" skills={result.skill_gaps} />
            )}
          </div>

          {/* ── Action plan ─────────────────────────────────────────────── */}
          {(result.action_plan?.length ?? 0) > 0 && (
            <div className="bg-gray-800 rounded-xl p-5 border border-gray-700 space-y-3">
              <h3 className="text-sm font-semibold text-white">🗺️ Your prep plan</h3>
              <div className="space-y-3">
                {result.action_plan.map((item, i) => (
                  <ActionCard key={i} item={item} index={i + 1} />
                ))}
              </div>
            </div>
          )}

          {/* ── Interview focus + time ───────────────────────────────────── */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {(result.interview_focus?.length ?? 0) > 0 && (
              <div className="bg-gray-800 rounded-xl p-4 border border-gray-700">
                <p className="text-xs text-gray-400 uppercase tracking-wider mb-3">🎤 Expect in interview</p>
                <div className="flex flex-wrap gap-2">
                  {result.interview_focus.map(topic => (
                    <span key={topic} className="bg-indigo-900/50 text-indigo-300 text-xs px-2.5 py-1 rounded-full border border-indigo-800">
                      {topic}
                    </span>
                  ))}
                </div>
              </div>
            )}
            <div className="bg-gray-800 rounded-xl p-4 border border-gray-700 space-y-3">
              <div>
                <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">⏱ Time to interview-ready</p>
                <p className="text-white font-semibold text-sm">{result.time_to_ready}</p>
              </div>
              <div>
                <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">Analysis confidence</p>
                <ConfidenceBadge confidence={result.confidence} />
              </div>
            </div>
          </div>
        </>
      )}

      {/* Prep plan + JD coach — only shown in full mode */}
      {!compact && (
        <div className="space-y-2">
          <div className="border-t border-gray-700 pt-4">
            <p className="text-sm font-semibold text-white mb-3">🗓️ Prep for this role</p>
            <PrepPlanPanel job={job} result={result} yoe={yoe} />
          </div>
        </div>
      )}

      {/* Re-analyse button */}
      <button
        onClick={() => { setResult(null); setError('') }}
        className="text-xs text-gray-500 hover:text-gray-300 transition-colors"
      >
        ↩ Re-analyse
      </button>
    </div>
  )
}

// ─── Sub-components ───────────────────────────────────────────────────────────

function TierFitBadge({ fit }: { fit: string }) {
  const map: Record<string, string> = {
    below: 'bg-yellow-900/50 text-yellow-300 border-yellow-800',
    match: 'bg-emerald-900/50 text-emerald-300 border-emerald-800',
    above: 'bg-blue-900/50 text-blue-300 border-blue-800',
  }
  const label: Record<string, string> = {
    below: '📈 Stretch role',
    match: '✓ Tier match',
    above: '⬆️ Over-qualified',
  }
  return (
    <span className={`text-xs px-2.5 py-1 rounded-full border font-medium ${map[fit] ?? ''}`}>
      {label[fit] ?? fit}
    </span>
  )
}

function CompanyBarBadge({ bar }: { bar: string }) {
  const map: Record<string, string> = {
    high: 'bg-red-900/40 text-red-300 border-red-800',
    medium: 'bg-orange-900/40 text-orange-300 border-orange-800',
    accessible: 'bg-emerald-900/40 text-emerald-300 border-emerald-800',
  }
  const label: Record<string, string> = {
    high: '🔴 High bar (FAANG)',
    medium: '🟡 Medium bar',
    accessible: '🟢 Accessible',
  }
  return (
    <span className={`text-xs px-2.5 py-1 rounded-full border font-medium ${map[bar] ?? ''}`}>
      {label[bar] ?? bar}
    </span>
  )
}

function ConfidenceBadge({ confidence }: { confidence: string }) {
  const map: Record<string, string> = {
    high: 'text-emerald-400',
    medium: 'text-yellow-400',
    low: 'text-gray-500',
  }
  return (
    <span className={`text-sm font-medium ${map[confidence] ?? ''}`}>
      {confidence.charAt(0).toUpperCase() + confidence.slice(1)}
    </span>
  )
}

function InfoCard({ label, value, icon }: { label: string; value: string; icon: string }) {
  return (
    <div className="bg-gray-800 rounded-xl p-4 border border-gray-700">
      <p className="text-xs text-gray-400 uppercase tracking-wider mb-1">{icon} {label}</p>
      <p className="text-gray-200 text-sm leading-snug">{value}</p>
    </div>
  )
}

function SkillBlock({ title, skills }: { title: string; skills: SkillSignal[] }) {
  return (
    <div className="bg-gray-800 rounded-xl p-4 border border-gray-700 space-y-2">
      <p className="text-xs font-semibold text-gray-300">{title}</p>
      {skills.map((s, i) => (
        <div key={i} className="flex items-start gap-2">
          <span className={`mt-0.5 text-xs font-bold shrink-0 ${levelColor(s.level)}`}>
            {levelIcon(s.level)}
          </span>
          <div>
            <p className="text-white text-xs font-medium">{s.skill}</p>
            <p className="text-gray-500 text-xs mt-0.5">{s.comment}</p>
          </div>
        </div>
      ))}
    </div>
  )
}

function ActionCard({ item, index }: { item: ActionItem; index: number }) {
  const priorityStyle: Record<string, string> = {
    critical: 'bg-red-900/30 border-red-800/50',
    high: 'bg-orange-900/20 border-orange-800/40',
    medium: 'bg-gray-700/40 border-gray-600/40',
  }
  const priorityLabel: Record<string, string> = {
    critical: '🔴 Critical',
    high: '🟠 High',
    medium: '🔵 Medium',
  }
  return (
    <div className={`rounded-lg px-4 py-3 border ${priorityStyle[item.priority] ?? 'border-gray-700'}`}>
      <div className="flex items-start gap-3">
        <span className="text-gray-500 text-xs font-mono mt-0.5 shrink-0">{index}.</span>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <p className="text-white text-sm font-medium">{item.title}</p>
            <span className="text-xs text-gray-500">{priorityLabel[item.priority]}</span>
          </div>
          <p className="text-gray-400 text-xs mt-1 leading-relaxed">{item.detail}</p>
          {item.resource && (
            <p className="text-indigo-400 text-xs mt-1">📚 {item.resource}</p>
          )}
        </div>
      </div>
    </div>
  )
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function matchColor(score: number) {
  if (score >= 75) return 'text-emerald-400'
  if (score >= 50) return 'text-yellow-400'
  return 'text-red-400'
}

function matchBarColor(score: number) {
  if (score >= 75) return 'bg-emerald-500'
  if (score >= 50) return 'bg-yellow-500'
  return 'bg-red-500'
}

function levelColor(level: string) {
  if (level === 'strong') return 'text-emerald-400'
  if (level === 'partial') return 'text-yellow-400'
  return 'text-red-400'
}

function levelIcon(level: string) {
  if (level === 'strong') return '✓'
  if (level === 'partial') return '~'
  return '✗'
}
