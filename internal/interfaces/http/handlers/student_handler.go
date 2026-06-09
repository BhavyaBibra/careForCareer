package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"careergps/internal/infrastructure/llm"
)

// StudentHandler handles Track C — student / fresher / non-tech-background assessments.
// This route is intentionally unauthenticated so students can try without registering.
type StudentHandler struct {
	llm llm.LLMProvider
}

func NewStudentHandler(llmProvider llm.LLMProvider) *StudentHandler {
	return &StudentHandler{llm: llmProvider}
}

// ── Request ────────────────────────────────────────────────────────────────────

type studentAssessRequest struct {
	// Q1 — academic status
	AcademicStatus string `json:"academic_status" binding:"required"`
	// e.g. "1st_year" | "2nd_year" | "3rd_year" | "final_year" | "recent_grad" | "self_learning"

	// Q2 — branch / background
	Branch string `json:"branch" binding:"required"`
	// e.g. "cs_it" | "ece_eee" | "mechanical_civil" | "bca_mca" | "non_engineering" | "diploma"

	// Q3 — target role (the north star)
	TargetRole string `json:"target_role" binding:"required"`
	// e.g. "blockchain" | "backend_sde" | "frontend_fullstack" | "data_science_ml" | "mobile" | "devops_cloud" | "cybersecurity" | "game_dev" | "other"
	TargetRoleCustom string `json:"target_role_custom"` // if target_role == "other"

	// Q4 — programming comfort
	ProgrammingLevel string `json:"programming_level" binding:"required"`
	// "none" | "basics" | "comfortable" | "solid"

	// Q5 — what they've built (comma-separated or array)
	ProjectsBuilt []string `json:"projects_built"` // e.g. ["nothing", "personal_website", "college_project", "working_project", "open_source", "real_users"]

	// Q6 — DSA level
	DSALevel string `json:"dsa_level" binding:"required"`
	// "never_tried" | "stuck_easy" | "easy" | "medium" | "hard"

	// Q7 — hours per day available
	HoursPerDay string `json:"hours_per_day" binding:"required"`
	// "less_than_1" | "1_to_2" | "2_to_4" | "4_plus"

	// Q8 — first milestone
	FirstMilestone string `json:"first_milestone" binding:"required"`
	// "campus_internship" | "offcampus_internship" | "first_job" | "freelance" | "startup" | "not_sure"
}

// ── Response ───────────────────────────────────────────────────────────────────

type StudentAssessmentResult struct {
	TargetRole        string              `json:"target_role"`
	HonestyNote       string              `json:"honesty_note"`        // 1-2 sentences: is this role realistic from this background?
	DomainReadiness   []DomainReadiness   `json:"domain_readiness"`    // per-domain score bars
	OverallReadiness  int                 `json:"overall_readiness"`   // 0-100 "hireable readiness" (not a match%, more a journey %)
	TimelinePhases    []StudentPhase      `json:"timeline_phases"`     // phase-based realistic timeline
	EarliestHireable  string              `json:"earliest_hireable"`   // e.g. "6 months"
	ComfortableTarget string              `json:"comfortable_target"`  // e.g. "10 months"
	Roadmap           StudentRoadmap      `json:"roadmap"`
	ChancesAssessment ChancesAssessment   `json:"chances_assessment"`
	QuickStartToday   []string            `json:"quick_start_today"`   // 3-4 things to do right now (free resources)
	HonestCaveats     []string            `json:"honest_caveats"`
}

type DomainReadiness struct {
	Domain      string `json:"domain"`       // e.g. "Programming Foundation", "DSA", "Blockchain Fundamentals"
	Score       int    `json:"score"`        // 0-100
	Status      string `json:"status"`       // "strong" | "partial" | "missing"
	Comment     string `json:"comment"`      // one specific sentence
	WeeksNeeded int    `json:"weeks_needed"` // rough weeks to get this domain to job-ready
}

type StudentPhase struct {
	Phase       string   `json:"phase"`        // e.g. "Foundation", "Core Skills", "Portfolio", "Apply"
	Weeks       string   `json:"weeks"`        // e.g. "1-8"
	Goal        string   `json:"goal"`
	Milestones  []string `json:"milestones"`
}

type StudentRoadmap struct {
	LearnResources  []RoadmapResource `json:"learn_resources"`
	ProjectsToBuild []string          `json:"projects_to_build"`
	Community       []string          `json:"community"`
	FirstJobTargets []string          `json:"first_job_targets"`
	PortfolioAdvice string            `json:"portfolio_advice"`
}

type RoadmapResource struct {
	Name        string `json:"name"`
	Type        string `json:"type"`   // "course" | "book" | "platform" | "channel"
	URL         string `json:"url"`
	Free        bool   `json:"free"`
	Description string `json:"description"`
}

type ChancesAssessment struct {
	InternshipChance string `json:"internship_chance"` // e.g. "High — Web3 startups actively hire self-taught devs"
	FirstJobChance   string `json:"first_job_chance"`
	TopBlocker       string `json:"top_blocker"`        // single biggest thing holding them back
	WildCard         string `json:"wild_card"`          // something that could fast-track them
}

// ── Handler ────────────────────────────────────────────────────────────────────

// Assess godoc
// POST /api/v1/student/assess  (no auth required)
func (h *StudentHandler) Assess(c *gin.Context) {
	var req studentAssessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorEnvelope("VALIDATION_ERROR", err.Error()))
		return
	}

	targetRole := req.TargetRole
	if targetRole == "other" && req.TargetRoleCustom != "" {
		targetRole = req.TargetRoleCustom
	}

	projectsStr := "nothing yet"
	if len(req.ProjectsBuilt) > 0 {
		projectsStr = strings.Join(req.ProjectsBuilt, ", ")
	}

	hoursLabel := map[string]string{
		"less_than_1": "less than 1 hour/day",
		"1_to_2":      "1–2 hours/day",
		"2_to_4":      "2–4 hours/day",
		"4_plus":      "4+ hours/day (full-time focus)",
	}

	systemPrompt := `You are a career mentor who specialises in helping students from ANY background — non-CS branches, tier-3 colleges, no guidance, no experience — break into specific tech roles.

You are HONEST but ENCOURAGING. You never say "just follow this generic roadmap." You give SPECIFIC advice for the exact role they want, calibrated to their current level.

Return ONLY valid JSON matching this exact schema — no markdown, no extra text:
{
  "target_role": "<clean role name, e.g. 'Blockchain Developer'>",
  "honesty_note": "<1-2 honest sentences: is this role realistic from their background? Mention the biggest challenge AND that it's achievable with consistency>",
  "domain_readiness": [
    {
      "domain": "<domain name specific to target role, e.g. 'Solidity / Smart Contracts' for blockchain>",
      "score": <0-100>,
      "status": "<strong|partial|missing>",
      "comment": "<one specific sentence about their current level in this domain>",
      "weeks_needed": <integer — weeks to get this domain to job-ready from current level>
    }
  ],
  "overall_readiness": <0-100, journey % — 0=complete beginner, 100=ready to apply. A student with zero background starts ~5-15>,
  "timeline_phases": [
    {
      "phase": "<phase name>",
      "weeks": "<e.g. '1-8'>",
      "goal": "<what they achieve in this phase>",
      "milestones": ["<milestone1>", "<milestone2>"]
    }
  ],
  "earliest_hireable": "<e.g. '6 months' — optimistic but realistic>",
  "comfortable_target": "<e.g. '10-12 months' — the timeline where they'd feel confident applying>",
  "roadmap": {
    "learn_resources": [
      {
        "name": "<resource name>",
        "type": "<course|book|platform|channel>",
        "url": "<actual URL — must be real>",
        "free": <true|false>,
        "description": "<one sentence on what this teaches and why it's the best for this role>"
      }
    ],
    "projects_to_build": ["<project1 — specific and achievable>", "<project2>", "<project3>"],
    "community": ["<community1 with link or platform>", "<community2>"],
    "first_job_targets": ["<type of company or specific target>"],
    "portfolio_advice": "<specific advice on what their GitHub/portfolio should look like for this role>"
  },
  "chances_assessment": {
    "internship_chance": "<honest assessment of internship chance after following this plan — e.g. 'High in Web3 startups after 6 months'>",
    "first_job_chance": "<honest assessment of first full-time role>",
    "top_blocker": "<single biggest thing that could stop them — be honest>",
    "wild_card": "<one thing that could fast-track them dramatically>"
  },
  "quick_start_today": ["<thing to do TODAY — must be free and specific>", "<thing2>", "<thing3>"],
  "honest_caveats": ["<honest caveat 1 — things most guides don't tell you>", "<caveat2>"]
}

Rules:
- domain_readiness: tailor domains 100% to the target role. Blockchain: Solidity, EVM, Web3.js/ethers.js, DeFi concepts, cryptography basics. NOT generic CS domains.
- For non-CS students: always include "Programming Foundation" as first domain even if score is 0.
- overall_readiness: reflect reality. Complete beginner in any field starts 5-15. Someone with some relevant projects might start 20-35. Nobody starts above 40 without significant prior work.
- timeline_phases: 3-4 phases. Keep phase weeks honest based on hours/day — someone with 1hr/day needs 2x the weeks of 4+hrs/day.
- roadmap.learn_resources: list 4-6 resources. Include at least 3 FREE resources. Name real, well-known resources — CryptoZombies, Patrick Collins' Solidity course, freeCodeCamp, The Odin Project, etc. Real URLs only.
- projects_to_build: 3 projects in order of difficulty. Start simple, end impressive. Role-specific.
- honest_caveats: say things guides don't — e.g. "Blockchain jobs in India are mostly remote-first or in Web3 startups, not traditional product companies" or "Most blockchain learning content assumes you can code — you'll need to build programming basics first"
- quick_start_today: must be doable TODAY, free, takes 30-60 minutes. E.g. "Create a free account on CryptoZombies and complete Level 1" not "start learning programming"
- hours/day directly affects weeks_needed — use this: 1hr/day = 3x multiplier, 2hrs/day = 1.5x, 4+hrs = 1x base`

	userPrompt := fmt.Sprintf(`STUDENT PROFILE:
- Academic status: %s
- Branch / background: %s
- Target role: %s
- Programming comfort: %s
- Projects built: %s
- DSA level: %s
- Available time: %s
- First milestone goal: %s

Assess this student's current positioning for their target role. Give them an honest, specific, encouraging assessment with a realistic timeline and exact roadmap.`,
		req.AcademicStatus,
		req.Branch,
		targetRole,
		req.ProgrammingLevel,
		projectsStr,
		req.DSALevel,
		orDefault(hoursLabel[req.HoursPerDay], req.HoursPerDay),
		req.FirstMilestone,
	)

	resp, err := h.llm.Generate(c.Request.Context(), llm.LLMRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    3000,
		Temperature:  0.3,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorEnvelope("LLM_ERROR", "Assessment failed — please try again"))
		return
	}

	raw := strings.TrimSpace(resp.Content)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var result StudentAssessmentResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		c.JSON(http.StatusInternalServerError, errorEnvelope("PARSE_ERROR", "Failed to parse assessment"))
		return
	}

	c.JSON(http.StatusOK, result)
}
