package assessment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"careergps/internal/domain/candidate"
	"careergps/internal/domain/gap"
	"careergps/internal/domain/jd"
	"careergps/internal/domain/readiness"
	"careergps/internal/domain/skill"
)

// Service orchestrates gap analysis and readiness scoring.
// All computation is deterministic — no LLM calls here.
type Service struct {
	candidateSkillRepo skill.CandidateSkillRepository
	jdSkillRepo        jd.JDSkillRepository
	gapRepo            gap.Repository
	readinessRepo      readiness.Repository
	weightTable        map[candidate.ExperienceTier]readiness.TierWeights
	engineVersion      string
}

func NewService(
	csRepo skill.CandidateSkillRepository,
	jdSkillRepo jd.JDSkillRepository,
	gapRepo gap.Repository,
	readinessRepo readiness.Repository,
	weightTable map[candidate.ExperienceTier]readiness.TierWeights,
	engineVersion string,
) *Service {
	return &Service{
		candidateSkillRepo: csRepo,
		jdSkillRepo:        jdSkillRepo,
		gapRepo:            gapRepo,
		readinessRepo:      readinessRepo,
		weightTable:        weightTable,
		engineVersion:      engineVersion,
	}
}

// RunGapAnalysis computes per-skill gaps and persists the result.
func (s *Service) RunGapAnalysis(ctx context.Context, cand *candidate.Candidate, jdID uuid.UUID, resumeID uuid.UUID) (*gap.GapAnalysis, error) {
	candidateSkills, err := s.candidateSkillRepo.ListByCandidateAndResume(ctx, cand.ID, resumeID)
	if err != nil {
		return nil, fmt.Errorf("assessment: get candidate skills: %w", err)
	}

	jdSkills, err := s.jdSkillRepo.ListByJDID(ctx, jdID)
	if err != nil {
		return nil, fmt.Errorf("assessment: get jd skills: %w", err)
	}

	// Build candidate skill map: skillID → score
	candidateMap := make(map[uuid.UUID]int, len(candidateSkills))
	for _, cs := range candidateSkills {
		candidateMap[cs.SkillID] = cs.Score
	}

	var gaps []gap.SkillGap
	var totalWeight float64
	var weightedGap float64
	var coveredSkills int

	for _, js := range jdSkills {
		candidateScore := candidateMap[js.SkillID] // 0 if absent
		if candidateScore > 0 {
			coveredSkills++
		}

		g := gap.ComputeGap(candidateScore, js.MinRequiredScore)
		tierMult := tierMultiplier(cand.InferredTier, js)

		weightedPriority := float64(g) * js.Weight * tierMult

		gaps = append(gaps, gap.SkillGap{
			SkillID:          js.SkillID,
			SkillName:        js.SkillName,
			CandidateScore:   candidateScore,
			JDRequiredScore:  js.MinRequiredScore,
			Gap:              g,
			IsRequired:       js.IsRequired,
			WeightedPriority: weightedPriority,
		})

		if js.IsRequired {
			totalWeight += js.Weight
			weightedGap += float64(g) * js.Weight
		}
	}

	aggregateGap := 0.0
	if totalWeight > 0 {
		aggregateGap = weightedGap / totalWeight
	}

	confidence := 0.0
	requiredCount := countRequired(jdSkills)
	if requiredCount > 0 {
		confidence = float64(coveredSkills) / float64(requiredCount)
	}

	ga := &gap.GapAnalysis{
		ID:           uuid.New(),
		CandidateID:  cand.ID,
		JDID:         jdID,
		Gaps:         gaps,
		AggregateGap: aggregateGap,
		Confidence:   confidence,
		CreatedAt:    time.Now().UTC(),
	}
	// FitLevel is set based on readiness score — placeholder for now, updated after scoring
	ga.FitLevel = gap.FitFromScore(100 - aggregateGap*10)

	if err := s.gapRepo.Create(ctx, ga); err != nil {
		return nil, fmt.Errorf("assessment: persist gap analysis: %w", err)
	}
	return ga, nil
}

// RunReadiness computes the composite readiness score and persists it.
func (s *Service) RunReadiness(ctx context.Context, cand *candidate.Candidate, jdID uuid.UUID, ga *gap.GapAnalysis) (*readiness.ReadinessAssessment, error) {
	weights, ok := s.weightTable[cand.InferredTier]
	if !ok {
		weights = readiness.DefaultWeightTable[cand.InferredTier]
	}

	components := s.computeComponents(cand, ga)

	compositeScore := readiness.Score(components, weights)

	// Build input snapshot for auditability
	snapshot, _ := json.Marshal(map[string]interface{}{
		"candidate_id":  cand.ID,
		"jd_id":         jdID,
		"tier":          cand.InferredTier,
		"gap_analysis":  ga.ID,
		"components":    components,
		"engine_version": s.engineVersion,
	})

	ra := &readiness.ReadinessAssessment{
		ID:            uuid.New(),
		CandidateID:   cand.ID,
		JDID:          jdID,
		GapAnalysisID: ga.ID,
		Tier:          cand.InferredTier,
		EngineVersion: s.engineVersion,
		CompositeScore: compositeScore,
		Components:    components,
		WeightsUsed:   weights,
		InputSnapshot: string(snapshot),
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.readinessRepo.Create(ctx, ra); err != nil {
		return nil, fmt.Errorf("assessment: persist readiness: %w", err)
	}
	return ra, nil
}

// computeComponents derives the [0,100] score for each component from gap analysis.
// All computation is deterministic.
func (s *Service) computeComponents(cand *candidate.Candidate, ga *gap.GapAnalysis) readiness.ComponentScores {
	var skillMatchSum, dsaSum, sdSum, archSum, domainSum float64
	var skillMatchW, dsaW, sdW, archW, domainW float64

	for _, g := range ga.Gaps {
		// Compute skill score as fraction of requirement met, scaled to [0,100]
		score := 0.0
		if g.JDRequiredScore > 0 {
			score = float64(g.CandidateScore) / float64(g.JDRequiredScore) * 100
			if score > 100 {
				score = 100
			}
		}

		switch skillCategory(g.SkillName) {
		case "dsa":
			dsaSum += score * g.WeightedPriority
			dsaW += g.WeightedPriority
		case "system_design":
			sdSum += score * g.WeightedPriority
			sdW += g.WeightedPriority
		case "architecture":
			archSum += score * g.WeightedPriority
			archW += g.WeightedPriority
		case "domain":
			domainSum += score * g.WeightedPriority
			domainW += g.WeightedPriority
		default:
			skillMatchSum += score * g.WeightedPriority
			skillMatchW += g.WeightedPriority
		}
	}

	avg := func(sum, w float64) float64 {
		if w == 0 {
			return 50 // neutral when no signal
		}
		return sum / w
	}

	// Experience match: simple ratio of YOE to typical tier expectation
	expMatch := experienceMatchScore(cand.YearsExperience, cand.InferredTier)

	return readiness.ComponentScores{
		SkillMatch:      avg(skillMatchSum, skillMatchW),
		DSASignal:       avg(dsaSum, dsaW),
		SystemDesign:    avg(sdSum, sdW),
		ArchDepth:       avg(archSum, archW),
		DomainRelevance: avg(domainSum, domainW),
		ExperienceMatch: expMatch,
	}
}

func experienceMatchScore(yoe int, tier candidate.ExperienceTier) float64 {
	// Score how well the YOE matches the centre of the tier band
	midpoints := map[candidate.ExperienceTier]float64{
		candidate.TierFreshGrad: 0,
		candidate.TierJunior:    2,
		candidate.TierMidLevel:  4.5,
		candidate.TierSenior:    8,
		candidate.TierStaff:     12,
	}
	mid := midpoints[tier]
	if mid == 0 {
		return 70
	}
	diff := float64(yoe) - mid
	if diff < 0 {
		diff = -diff
	}
	score := 100 - diff*10
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// skillCategory is a lightweight heuristic — will be replaced by skill registry lookup.
func skillCategory(skillName string) string {
	dsaKeywords := []string{"dsa", "algorithms", "data structures", "leetcode", "competitive"}
	sdKeywords := []string{"system design", "distributed", "scalability", "availability"}
	archKeywords := []string{"architecture", "microservices", "platform", "infrastructure"}
	domainKeywords := []string{"payments", "fintech", "ecommerce", "machine learning", "ml"}

	lower := toLower(skillName)
	for _, k := range dsaKeywords {
		if contains(lower, k) {
			return "dsa"
		}
	}
	for _, k := range sdKeywords {
		if contains(lower, k) {
			return "system_design"
		}
	}
	for _, k := range archKeywords {
		if contains(lower, k) {
			return "architecture"
		}
	}
	for _, k := range domainKeywords {
		if contains(lower, k) {
			return "domain"
		}
	}
	return "backend"
}

func tierMultiplier(tier candidate.ExperienceTier, js *jd.JDSkill) float64 {
	// Higher tiers weight architecture and system design gaps more heavily
	if tier >= candidate.TierSenior {
		return 1.3
	}
	if tier <= candidate.TierJunior {
		return 0.8
	}
	return 1.0
}

func countRequired(skills []*jd.JDSkill) int {
	c := 0
	for _, s := range skills {
		if s.IsRequired {
			c++
		}
	}
	return c
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, sub string) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
