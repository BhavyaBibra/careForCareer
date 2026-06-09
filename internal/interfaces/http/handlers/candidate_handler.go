package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"careergps/internal/domain/candidate"
	"careergps/internal/infrastructure/postgres"
	"careergps/internal/interfaces/http/middleware"
	"careergps/pkg/apperrors"
)

type CandidateHandler struct {
	candidateRepo *postgres.CandidateRepo
}

func NewCandidateHandler(candidateRepo *postgres.CandidateRepo) *CandidateHandler {
	return &CandidateHandler{candidateRepo: candidateRepo}
}

type upsertCandidateRequest struct {
	YearsExperience int    `json:"years_experience" binding:"min=0,max=50"`
	CurrentCompany  string `json:"current_company"`
	CurrentCompINR  int64  `json:"current_comp_inr"`
	TargetCompINR   int64  `json:"target_comp_inr"`
}

// GetProfile returns the candidate profile for the authenticated user.
func (h *CandidateHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorEnvelope("UNAUTHORIZED", "Not authenticated"))
		return
	}

	cand, err := h.candidateRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, errorEnvelope("NOT_FOUND", "Candidate profile not found. Create one first."))
			return
		}
		c.JSON(http.StatusInternalServerError, errorEnvelope("INTERNAL_ERROR", "Failed to fetch profile"))
		return
	}

	c.JSON(http.StatusOK, candidateResponse(cand))
}

// CreateProfile creates a new candidate profile for the authenticated user.
func (h *CandidateHandler) CreateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorEnvelope("UNAUTHORIZED", "Not authenticated"))
		return
	}

	var req upsertCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorEnvelope("VALIDATION_ERROR", err.Error()))
		return
	}

	// Check if profile already exists
	existing, err := h.candidateRepo.GetByUserID(c.Request.Context(), userID)
	if err == nil && existing != nil {
		c.JSON(http.StatusConflict, errorEnvelope("ALREADY_EXISTS", "Profile already exists. Use PUT to update."))
		return
	}

	cand, err := candidate.New(
		userID,
		req.YearsExperience,
		req.CurrentCompany,
		candidate.CompensationINR(req.CurrentCompINR),
		candidate.CompensationINR(req.TargetCompINR),
		nil,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorEnvelope("VALIDATION_ERROR", err.Error()))
		return
	}

	if err := h.candidateRepo.Create(c.Request.Context(), cand); err != nil {
		c.JSON(http.StatusInternalServerError, errorEnvelope("INTERNAL_ERROR", "Failed to create profile"))
		return
	}

	c.JSON(http.StatusCreated, candidateResponse(cand))
}

// UpdateProfile updates the candidate profile for the authenticated user.
func (h *CandidateHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, errorEnvelope("UNAUTHORIZED", "Not authenticated"))
		return
	}

	var req upsertCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorEnvelope("VALIDATION_ERROR", err.Error()))
		return
	}

	cand, err := h.candidateRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, errorEnvelope("NOT_FOUND", "Profile not found. Create one first via POST /candidate"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorEnvelope("INTERNAL_ERROR", "Failed to fetch profile"))
		return
	}

	if err := cand.UpdateProfile(
		req.YearsExperience,
		req.CurrentCompany,
		candidate.CompensationINR(req.CurrentCompINR),
		candidate.CompensationINR(req.TargetCompINR),
		nil,
	); err != nil {
		c.JSON(http.StatusBadRequest, errorEnvelope("VALIDATION_ERROR", err.Error()))
		return
	}

	if err := h.candidateRepo.Update(c.Request.Context(), cand); err != nil {
		c.JSON(http.StatusInternalServerError, errorEnvelope("INTERNAL_ERROR", "Failed to update profile"))
		return
	}

	c.JSON(http.StatusOK, candidateResponse(cand))
}

func candidateResponse(cand *candidate.Candidate) gin.H {
	return gin.H{
		"candidate_id":     cand.ID,
		"user_id":          cand.UserID,
		"years_experience": cand.YearsExperience,
		"tier":             cand.InferredTier,
		"tier_label":       cand.InferredTier.TierLabel(),
		"tier_explanation": cand.TierExplanation,
		"current_company":  cand.CurrentCompany,
		"current_comp_inr": int64(cand.CurrentComp),
		"target_comp_inr":  int64(cand.TargetComp),
		"created_at":       cand.CreatedAt,
		"updated_at":       cand.UpdatedAt,
	}
}
