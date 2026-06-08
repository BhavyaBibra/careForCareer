package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"careergps/internal/domain/readiness"
)

type AssessmentHandler struct {
	readinessRepo readiness.Repository
}

func NewAssessmentHandler(readinessRepo readiness.Repository) *AssessmentHandler {
	return &AssessmentHandler{readinessRepo: readinessRepo}
}

func (h *AssessmentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorEnvelope("INVALID_UUID", "Invalid assessment ID"))
		return
	}
	ra, err := h.readinessRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorEnvelope("NOT_FOUND", "Assessment not found"))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"assessment_id":   ra.ID,
		"composite_score": ra.CompositeScore,
		"engine_version":  ra.EngineVersion,
		"tier":            ra.Tier,
		"created_at":      ra.CreatedAt,
	})
}

func (h *AssessmentHandler) GetReadiness(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errorEnvelope("INVALID_UUID", "Invalid assessment ID"))
		return
	}
	ra, err := h.readinessRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, errorEnvelope("NOT_FOUND", "Assessment not found"))
		return
	}

	contributions := ra.ComponentContribution()

	c.JSON(http.StatusOK, gin.H{
		"composite_score": ra.CompositeScore,
		"breakdown": gin.H{
			"skill_match": gin.H{
				"score":        ra.Components.SkillMatch,
				"weight":       ra.WeightsUsed.SkillMatch,
				"contribution": contributions["skill_match"],
			},
			"dsa_signal": gin.H{
				"score":        ra.Components.DSASignal,
				"weight":       ra.WeightsUsed.DSASignal,
				"contribution": contributions["dsa_signal"],
			},
			"system_design": gin.H{
				"score":        ra.Components.SystemDesign,
				"weight":       ra.WeightsUsed.SystemDesign,
				"contribution": contributions["system_design"],
			},
			"arch_depth": gin.H{
				"score":        ra.Components.ArchDepth,
				"weight":       ra.WeightsUsed.ArchDepth,
				"contribution": contributions["arch_depth"],
			},
			"domain_relevance": gin.H{
				"score":        ra.Components.DomainRelevance,
				"weight":       ra.WeightsUsed.DomainRelevance,
				"contribution": contributions["domain_relevance"],
			},
			"experience_match": gin.H{
				"score":        ra.Components.ExperienceMatch,
				"weight":       ra.WeightsUsed.ExperienceMatch,
				"contribution": contributions["experience_match"],
			},
		},
		"engine_version": ra.EngineVersion,
		"tier":           ra.Tier,
		"weights_used":   ra.WeightsUsed,
	})
}
