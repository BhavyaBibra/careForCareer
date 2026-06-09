package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Job represents a normalized job listing returned to the frontend.
type Job struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	ApplyURL    string `json:"apply_url"`
	Description string `json:"description"`
	PostedAt    string `json:"posted_at,omitempty"`
	Source      string `json:"source"`
}

// JobsHandler handles job search requests.
type JobsHandler struct {
	apifyToken string
	httpClient *http.Client
}

func NewJobsHandler() *JobsHandler {
	return &JobsHandler{
		apifyToken: os.Getenv("APIFY_API_TOKEN"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Search godoc
// GET /api/v1/jobs/search?q=backend+engineer&location=bangalore&limit=20
func (h *JobsHandler) Search(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	location := strings.TrimSpace(c.Query("location"))
	limitStr := c.DefaultQuery("limit", "20")

	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"message": "query param 'q' is required"}})
		return
	}

	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// Build search query combining keywords and location
	searchQuery := q
	if location != "" {
		searchQuery = fmt.Sprintf("%s %s", q, location)
	}

	var jobs []Job
	var err error

	if h.apifyToken != "" {
		jobs, err = h.fetchFromApify(c.Request.Context(), searchQuery, location, limit)
	}

	// Fallback to mock data if Apify fails or token not set
	if err != nil || h.apifyToken == "" {
		jobs = h.mockJobs(q, location, limit)
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":  jobs,
		"total": len(jobs),
		"query": q,
		"location": location,
	})
}

// fetchFromApify calls the Apify LinkedIn Jobs Scraper actor.
func (h *JobsHandler) fetchFromApify(ctx context.Context, query, location string, limit int) ([]Job, error) {
	// Use apify/linkedin-jobs-scraper actor
	actorID := "hKByXkMQaC5Qt9UMG" // Apify linkedin-jobs-scraper

	payload := map[string]interface{}{
		"queries":  []string{query},
		"location": location,
		"count":    limit,
	}

	payloadBytes, _ := json.Marshal(payload)

	runURL := fmt.Sprintf("https://api.apify.com/v2/acts/%s/run-sync-get-dataset-items?token=%s&timeout=25",
		actorID, h.apifyToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, runURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("apify returned %d", resp.StatusCode)
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	jobs := make([]Job, 0, len(raw))
	for i, item := range raw {
		if i >= limit {
			break
		}
		job := Job{
			ID:     getString(item, "id", fmt.Sprintf("linkedin-%d", i)),
			Title:  getString(item, "title", getString(item, "position", "")),
			Company: getString(item, "company", getString(item, "companyName", "")),
			Location: getString(item, "location", ""),
			ApplyURL: getString(item, "url", getString(item, "applyUrl", getString(item, "jobUrl", ""))),
			Description: getString(item, "description", getString(item, "descriptionText", "")),
			PostedAt: getString(item, "publishedAt", ""),
			Source:  "linkedin",
		}
		if job.Title != "" {
			jobs = append(jobs, job)
		}
	}
	return jobs, nil
}

// mockJobs returns realistic-looking placeholder jobs for when Apify is unavailable.
func (h *JobsHandler) mockJobs(query, location string, limit int) []Job {
	loc := location
	if loc == "" {
		loc = "India"
	}

	templates := []Job{
		{
			ID:       "mock-1",
			Title:    fmt.Sprintf("Senior %s Engineer", titleCase(query)),
			Company:  "Amazon",
			Location: loc,
			ApplyURL: "https://www.linkedin.com/jobs/search/?keywords=" + url.QueryEscape(query),
			Description: fmt.Sprintf("We are looking for a Senior %s Engineer to join our growing team. You will design and implement scalable distributed systems, mentor junior engineers, and drive technical decisions.\n\nRequirements:\n- 5+ years of experience\n- Strong system design skills\n- Proficiency in Go, Java, or Python\n- Experience with AWS, microservices", titleCase(query)),
			Source:   "mock",
		},
		{
			ID:       "mock-2",
			Title:    fmt.Sprintf("%s Engineer II", titleCase(query)),
			Company:  "Google",
			Location: loc,
			ApplyURL: "https://www.linkedin.com/jobs/search/?keywords=" + url.QueryEscape(query),
			Description: fmt.Sprintf("%s Engineer II at Google. Join our team building products used by billions. You'll work on large-scale infrastructure and develop innovative solutions.\n\nRequirements:\n- 3+ years experience\n- Strong CS fundamentals\n- Coding proficiency in C++, Java, Go, or Python\n- Problem solving and system design skills", titleCase(query)),
			Source:   "mock",
		},
		{
			ID:       "mock-3",
			Title:    fmt.Sprintf("Staff %s Engineer", titleCase(query)),
			Company:  "Microsoft",
			Location: loc,
			ApplyURL: "https://www.linkedin.com/jobs/search/?keywords=" + url.QueryEscape(query),
			Description: fmt.Sprintf("Staff %s Engineer — Azure. Drive the technical strategy for our cloud platform. Own end-to-end delivery of complex systems.\n\nRequirements:\n- 8+ years of experience\n- Track record of leading large engineering projects\n- Deep expertise in distributed systems\n- Excellent communication skills", titleCase(query)),
			Source:   "mock",
		},
		{
			ID:       "mock-4",
			Title:    fmt.Sprintf("%s Developer (Backend)", titleCase(query)),
			Company:  "Flipkart",
			Location: loc,
			ApplyURL: "https://www.linkedin.com/jobs/search/?keywords=" + url.QueryEscape(query),
			Description: fmt.Sprintf("Backend %s Developer at Flipkart. Build high-performance APIs and data pipelines for India's largest e-commerce platform.\n\nRequired skills:\n- Go / Java / Python\n- Kafka, Redis, MySQL\n- REST API design\n- 2+ years experience", titleCase(query)),
			Source:   "mock",
		},
		{
			ID:       "mock-5",
			Title:    fmt.Sprintf("Principal %s Engineer", titleCase(query)),
			Company:  "Uber",
			Location: loc,
			ApplyURL: "https://www.linkedin.com/jobs/search/?keywords=" + url.QueryEscape(query),
			Description: fmt.Sprintf("Principal %s Engineer at Uber. Define the architecture for our real-time platform serving millions of rides daily.\n\nRequirements:\n- 10+ years of engineering experience\n- Deep expertise in distributed systems and databases\n- Proven technical leadership\n- Strong system design skills", titleCase(query)),
			Source:   "mock",
		},
		{
			ID:       "mock-6",
			Title:    fmt.Sprintf("SDE-2 (%s)", titleCase(query)),
			Company:  "Swiggy",
			Location: loc,
			ApplyURL: "https://www.linkedin.com/jobs/search/?keywords=" + url.QueryEscape(query),
			Description: fmt.Sprintf("Software Development Engineer-2 at Swiggy. Work on our food delivery platform scaling to millions of orders.\n\nSkills required:\n- 3-5 years experience\n- Java/Go/Python\n- Microservices, Docker, Kubernetes\n- SQL and NoSQL databases\n- Strong problem-solving skills"),
			Source:   "mock",
		},
		{
			ID:       "mock-7",
			Title:    fmt.Sprintf("Senior %s Engineer", titleCase(query)),
			Company:  "Meesho",
			Location: loc,
			ApplyURL: "https://www.linkedin.com/jobs/search/?keywords=" + url.QueryEscape(query),
			Description: fmt.Sprintf("Senior %s Engineer at Meesho. Help build the commerce infrastructure for 150M+ Indian entrepreneurs.\n\nWhat we're looking for:\n- 4-7 years experience\n- Strong backend development skills\n- Experience with high-traffic systems\n- Passion for impact at scale", titleCase(query)),
			Source:   "mock",
		},
		{
			ID:       "mock-8",
			Title:    fmt.Sprintf("%s Engineer - Payments", titleCase(query)),
			Company:  "Razorpay",
			Location: loc,
			ApplyURL: "https://www.linkedin.com/jobs/search/?keywords=" + url.QueryEscape(query),
			Description: fmt.Sprintf("%s Engineer on the Payments team at Razorpay. Build reliable, secure payment infrastructure processing billions of transactions.\n\nRequirements:\n- 3+ years backend experience\n- Strong understanding of distributed systems\n- Knowledge of payment protocols a plus\n- Go, Java, or Python", titleCase(query)),
			Source:   "mock",
		},
	}

	if limit > len(templates) {
		limit = len(templates)
	}
	return templates[:limit]
}

func getString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func titleCase(s string) string {
	if s == "" {
		return "Software"
	}
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
