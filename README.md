![Go](https://img.shields.io/badge/Go-1.22-blue)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue)
![Redis](https://img.shields.io/badge/Redis-7-red)
![React](https://img.shields.io/badge/React-18-61dafb)
![License](https://img.shields.io/badge/license-MIT-green)

# careForCareer

> AI-powered career intelligence platform for software engineers.
> Analyses your resume against live job postings, scores skill gaps, generates week-by-week prep plans, and coaches you for a specific role — all grounded in your actual resume and the actual JD.

**Built with:** Go · PostgreSQL · Redis · React · Anthropic Claude API · AWS S3

---

## The problem it solves

Most career tools give generic advice. careForCareer is built around one idea: **every answer should be grounded in your specific resume and the specific job you're targeting.**

Upload your resume → search for a job on LinkedIn → see exactly where you stand → get a week-by-week plan → chat with a coach that knows the full context.

---

## Features

| Feature | Description |
|---|---|
| Resume upload + parse | Upload PDF → extracted, stored in S3, parsed by LLM |
| LinkedIn job search | Search by keyword + location via Apify scraper (mock fallback if no API key) |
| Candidate positioning | Match % against a JD, skill gaps, tier fit, company bar, time-to-ready |
| Week-by-week prep plan | Personalised study plan from your gaps + the JD |
| JD-aware coach | SSE streaming coach that knows your resume, the JD, and your positioning |
| Career pivot analyser | PM / TPM / EM / SRE pivot: transferable skills, timeline, entry paths, day-in-the-life |
| Student track | For non-CS / tier-3 college students targeting specific roles _(placeholder — coming soon)_ |
| Guest access | Pivot and student flows work without an account |

---

## Architecture

```
careForCareer/
├── cmd/
│   ├── api/              # HTTP server entrypoint
│   └── worker/           # Background job worker (Asynq)
├── config/               # Viper config + RSA key loading
├── internal/
│   ├── domain/           # Pure domain types — zero external imports
│   │   ├── auth/  candidate/  coach/  company/
│   │   ├── gap/   jd/         readiness/  resume/
│   │   └── roadmap/  skill/
│   ├── application/      # Use cases (auth service, coach service)
│   ├── infrastructure/
│   │   ├── llm/          # Anthropic Claude provider + Redis caching layer
│   │   ├── postgres/     # Repository implementations (11 tables)
│   │   ├── redis/        # Cache + daily message counter
│   │   └── s3/           # Resume storage (MinIO locally, S3 in prod)
│   └── interfaces/
│       └── http/
│           ├── handlers/ # One handler per feature (auth, jobs, positioning, prep, pivot, student)
│           ├── middleware/  # JWT auth, request ID
│           └── router.go
├── migrations/           # 11 SQL migration files (golang-migrate)
├── seed/                 # Dev seed data
├── frontend/             # React 18 + Vite + TypeScript + Tailwind CSS
│   └── src/
│       ├── pages/        # Landing · Login · Onboarding · Dashboard · Pivot · Student
│       ├── components/   # PositioningPanel · PrepPlanPanel
│       └── api/          # Typed API clients per feature
└── deployments/          # Dockerfile + docker-compose
```

### Key design decisions

- **Hexagonal (clean) architecture** — domain layer has zero infrastructure imports; swap Postgres for anything without touching business logic
- **RS256 JWT** — asymmetric signing; access token lives in memory (XSS safe), refresh token in sessionStorage
- **Redis-cached LLM inference** — cache key = hash of (system prompt + user prompt + model version); configurable TTL eliminates redundant API calls
- **SSE over WebSockets** — coach streaming is server→client only; SSE works over plain HTTP/1.1 and auto-reconnects
- **Structured JSON from LLM** — all Claude responses return typed JSON validated against Go structs; no freeform parsing

---

## Tech stack

| Layer | Tech |
|---|---|
| Backend language | Go 1.22 |
| HTTP framework | Gin |
| Database | PostgreSQL 16 |
| Cache + background jobs | Redis 7, Asynq |
| File storage | AWS S3 / MinIO (local) |
| Auth | JWT RS256 |
| LLM | Anthropic Claude (claude-sonnet) |
| Job scraping | Apify LinkedIn Jobs actor |
| Frontend | React 18, Vite, TypeScript, Tailwind CSS |
| Infra | Docker, docker-compose |

---

## API routes

```
# Auth (public)
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout

# Profile (authenticated)
GET    /api/v1/candidate
POST   /api/v1/candidate
PUT    /api/v1/candidate

# Resume (authenticated)
POST   /api/v1/resumes
GET    /api/v1/resumes/:id

# Assessments (authenticated)
GET    /api/v1/assessments/:id
GET    /api/v1/assessments/:id/readiness

# Jobs + Positioning (authenticated)
GET    /api/v1/jobs/search              # keyword + location → LinkedIn jobs
POST   /api/v1/jobs/position            # resume + JD → match %, gaps, action plan
POST   /api/v1/jobs/prep-plan           # gaps + JD → week-by-week study plan

# Coach (authenticated)
POST   /api/v1/coach/sessions
GET    /api/v1/coach/sessions/:id
POST   /api/v1/coach/sessions/:id/messages
GET    /api/v1/coach/sessions/:id/stream       # SSE streaming

# JD-aware coach (authenticated)
POST   /api/v1/coach/jd-sessions
GET    /api/v1/coach/jd-sessions/:id/stream    # SSE streaming

# Career pivot (authenticated — needs profile for YOE/tier context)
POST   /api/v1/pivot/analyse

# Student track (open — no auth required)
POST   /api/v1/student/assess
```

---

## Local setup

**Prerequisites:** Go 1.22+, Node 20+, Docker, `golang-migrate` CLI

```bash
# 1. Clone
git clone https://github.com/zekst/careForCareer
cd careForCareer

# 2. Start PostgreSQL, Redis, MinIO
make docker-up

# 3. Copy env and fill in values
cp .env.example .env
# Required: ANTHROPIC_API_KEY, DATABASE_URL, REDIS_ADDR, S3 keys
# Optional: APIFY_API_TOKEN (falls back to mock jobs if not set)

# 4. Generate RSA keys for JWT
make keys

# 5. Run database migrations
make migrate-up

# 6. Seed dev data (optional)
make seed

# 7. Start API server
make run-api          # → http://localhost:8080

# 8. Start frontend (separate terminal)
cd frontend
npm install
npm run dev           # → http://localhost:5173
```

---

## Status

### Done ✅

- [x] Auth — register, login, JWT RS256, refresh token rotation, logout
- [x] Candidate profile — YOE, current company, skills, tier inference
- [x] Resume upload — PDF → S3, LLM parse + readiness score
- [x] Job search — LinkedIn scraper via Apify with realistic mock fallback
- [x] Candidate positioning — match %, skill gaps, tier fit, company bar, action plan, time-to-ready
- [x] Prep plan generator — week-by-week study plan from JD + positioning gaps
- [x] JD-aware coach — SSE streaming, full positioning context injected, session snapshots in PostgreSQL
- [x] Career pivot analyser — transferable skills, skills to learn, timeline, entry paths, day-in-the-life
- [x] Career pivot frontend — role picker, 4-tab result (overview / skills / paths / plan)
- [x] Landing page — 3-track selector (Professional / Career Pivot / Student)
- [x] Guest access — student and pivot routes accessible without account
- [x] Student track placeholder page

### In progress 🔧

- [ ] End-to-end build verification on host machine
- [ ] Student track full implementation — 8-question profile, domain readiness bars, phase timeline, role-specific roadmap

### Pending 📋

- [ ] Hosting + deployment (Railway / Fly.io / EC2 — TBD)
- [ ] Naukri + Wellfound job sources (currently only LinkedIn via Apify)
- [ ] Rate limiting on open routes (student/assess)
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] End-to-end and integration tests
- [ ] Custom domain + SSL

---

## License

MIT
