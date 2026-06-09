import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

const TRACKS = [
  {
    id: 'professional',
    icon: '💼',
    title: 'Professional',
    subtitle: 'I have a job or recent experience',
    description: 'Upload your resume, find jobs on LinkedIn, see your positioning, and get a personalised prep plan for any role.',
    bullets: ['Resume analysis', 'LinkedIn job search', 'Match % against JDs', 'Week-by-week prep', 'Role-specific coach'],
    cta: 'Get started',
    ctaPath: '/login',
    accent: 'indigo',
    badge: 'Most popular',
  },
  {
    id: 'pivot',
    icon: '🔄',
    title: 'Career Pivot',
    subtitle: 'Engineer switching to PM / TPM / EM / SRE',
    description: 'Honest assessment of your pivot: what transfers, what\'s genuinely new, realistic timelines, and the best entry paths.',
    bullets: ['Transferable skills analysis', 'Pivot difficulty rating', 'Entry paths (APM, internal transfer…)', 'Phase-by-phase prep plan', 'Day-in-the-life reality check'],
    cta: 'Analyse my pivot',
    ctaPath: '/pivot',
    accent: 'purple',
    badge: null,
  },
  {
    id: 'student',
    icon: '🎓',
    title: 'Student / Fresher',
    subtitle: 'No experience, any college, any branch',
    description: 'Designed for students who decided on a tech role but have no mentor, no placement cell, and no clear path forward.',
    bullets: ['No resume needed', 'Any branch — CS, Mech, ECE…', 'Role-specific roadmap', 'Free resource recommendations', 'Honest timeline'],
    cta: 'See my path',
    ctaPath: '/student',
    accent: 'teal',
    badge: 'Coming soon',
  },
]

const accentStyles: Record<string, {
  bg: string; border: string; icon: string; badge: string; cta: string; bullet: string
}> = {
  indigo: {
    bg: 'bg-indigo-900/20',
    border: 'border-indigo-800/60',
    icon: 'bg-indigo-900/50 border-indigo-800',
    badge: 'bg-indigo-600 text-white',
    cta: 'bg-indigo-600 hover:bg-indigo-500',
    bullet: 'text-indigo-400',
  },
  purple: {
    bg: 'bg-purple-900/20',
    border: 'border-purple-800/60',
    icon: 'bg-purple-900/50 border-purple-800',
    badge: 'bg-purple-600 text-white',
    cta: 'bg-purple-600 hover:bg-purple-500',
    bullet: 'text-purple-400',
  },
  teal: {
    bg: 'bg-teal-900/20',
    border: 'border-teal-800/60',
    icon: 'bg-teal-900/50 border-teal-800',
    badge: 'bg-teal-700 text-teal-100',
    cta: 'bg-teal-600 hover:bg-teal-500',
    bullet: 'text-teal-400',
  },
}

export default function LandingPage() {
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth()

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      {/* Nav */}
      <div className="border-b border-gray-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-xl">🧭</span>
          <span className="font-bold text-white">CareerGPS</span>
          <span className="text-xs text-gray-600 ml-1">for Indian engineers</span>
        </div>
        <div className="flex items-center gap-3">
          {isAuthenticated ? (
            <button onClick={() => navigate('/dashboard')}
              className="text-sm bg-gray-800 hover:bg-gray-700 text-gray-200 px-4 py-2 rounded-lg border border-gray-700 transition-colors">
              Dashboard →
            </button>
          ) : (
            <>
              <button onClick={() => navigate('/login')}
                className="text-sm text-gray-400 hover:text-white transition-colors">
                Sign in
              </button>
              <button onClick={() => navigate('/register')}
                className="text-sm bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-lg transition-colors">
                Create account
              </button>
            </>
          )}
        </div>
      </div>

      {/* Hero */}
      <div className="text-center px-4 py-16 max-w-2xl mx-auto space-y-4">
        <h1 className="text-4xl font-bold text-white leading-tight">
          Your career, honestly assessed.
        </h1>
        <p className="text-gray-400 text-lg leading-relaxed">
          Not generic advice. Not fake encouragement.
          Real positioning, real gaps, real timelines — built for Indian engineers.
        </p>
      </div>

      {/* Track cards */}
      <div className="max-w-5xl mx-auto px-4 pb-20">
        <p className="text-center text-xs text-gray-500 uppercase tracking-wider mb-8">Choose your track</p>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          {TRACKS.map(track => {
            const s = accentStyles[track.accent]
            return (
              <div
                key={track.id}
                className={`relative rounded-2xl border p-6 flex flex-col gap-4 transition-all hover:scale-[1.01] ${s.bg} ${s.border}`}
              >
                {track.badge && (
                  <span className={`absolute top-4 right-4 text-xs px-2.5 py-1 rounded-full font-medium ${s.badge}`}>
                    {track.badge}
                  </span>
                )}

                <div className={`w-12 h-12 rounded-xl border flex items-center justify-center text-2xl ${s.icon}`}>
                  {track.icon}
                </div>

                <div>
                  <h2 className="text-lg font-bold text-white">{track.title}</h2>
                  <p className="text-xs text-gray-400 mt-0.5">{track.subtitle}</p>
                </div>

                <p className="text-gray-300 text-sm leading-relaxed">{track.description}</p>

                <ul className="space-y-1.5 flex-1">
                  {track.bullets.map(b => (
                    <li key={b} className="flex items-center gap-2 text-xs text-gray-400">
                      <span className={`shrink-0 ${s.bullet}`}>✓</span>
                      {b}
                    </li>
                  ))}
                </ul>

                <button
                  onClick={() => navigate(track.ctaPath)}
                  className={`w-full text-white font-semibold text-sm py-3 rounded-xl transition-colors ${s.cta}`}
                >
                  {track.cta} →
                </button>
              </div>
            )
          })}
        </div>
      </div>

      {/* Footer */}
      <div className="border-t border-gray-800 px-6 py-6 text-center text-xs text-gray-600">
        CareerGPS · Built for engineers who want honest, specific career guidance · No ads, no fluff
      </div>
    </div>
  )
}
