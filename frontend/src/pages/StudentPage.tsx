import { useNavigate } from 'react-router-dom'

/**
 * Track C — Student / Fresher / Non-tech background
 * Placeholder — full implementation coming soon.
 */
export default function StudentPage() {
  const navigate = useNavigate()

  return (
    <div className="min-h-screen bg-gray-950 text-white flex flex-col">
      {/* Header */}
      <div className="border-b border-gray-800 px-6 py-4 flex items-center justify-between">
        <button onClick={() => navigate('/')} className="flex items-center gap-2 text-gray-400 hover:text-white transition-colors text-sm">
          ← Back
        </button>
        <span className="text-xs bg-teal-900/50 text-teal-300 border border-teal-800 px-2.5 py-1 rounded-full font-medium">
          🎓 Student Track
        </span>
        <button onClick={() => navigate('/login')} className="text-xs text-gray-500 hover:text-gray-300 transition-colors">
          Sign in →
        </button>
      </div>

      {/* Main */}
      <div className="flex-1 flex items-center justify-center px-4">
        <div className="text-center space-y-6 max-w-md">
          {/* Icon */}
          <div className="w-20 h-20 rounded-2xl bg-teal-900/40 border border-teal-800/60 flex items-center justify-center mx-auto text-4xl">
            🎓
          </div>

          <div className="space-y-2">
            <h1 className="text-2xl font-bold text-white">Student Track — Coming Soon</h1>
            <p className="text-gray-400 text-sm leading-relaxed">
              We're building a dedicated path for students from any background — non-CS branches,
              tier-3 colleges, zero experience. Questionnaire-based positioning, honest timelines,
              and role-specific roadmaps.
            </p>
          </div>

          {/* What's coming */}
          <div className="bg-gray-800/60 border border-gray-700 rounded-xl p-5 text-left space-y-3">
            <p className="text-xs text-gray-400 uppercase tracking-wider font-semibold">What's coming</p>
            {[
              '8-question profile (no resume needed)',
              'Domain-level readiness for your target role',
              'Honest timeline: "6 months if you put in 2hr/day"',
              'Role-specific roadmap with free resources',
              'Student coach who knows your full context',
            ].map((item, i) => (
              <div key={i} className="flex items-start gap-2.5 text-sm text-gray-300">
                <span className="text-teal-500 shrink-0 mt-0.5">✓</span>
                {item}
              </div>
            ))}
          </div>

          {/* Notify me */}
          <div className="space-y-3">
            <p className="text-gray-500 text-xs">
              In the meantime, if you already have some coding experience, the Professional track works for freshers too.
            </p>
            <div className="flex gap-2 justify-center">
              <button
                onClick={() => navigate('/login')}
                className="bg-teal-600 hover:bg-teal-500 text-white text-sm font-semibold px-5 py-2.5 rounded-lg transition-colors"
              >
                Use Professional track
              </button>
              <button
                onClick={() => navigate('/')}
                className="bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm px-5 py-2.5 rounded-lg transition-colors border border-gray-700"
              >
                Go back
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
