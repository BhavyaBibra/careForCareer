package llm

import (
	"fmt"
	"regexp"
	"strings"
)

// Redactor replaces PII tokens before sending text to any LLM provider.
// Replacement is per-request — token map is not stored or logged.
// LLM never sees real names, emails, phone numbers, or company names.
type Redactor struct {
	emailRegex *regexp.Regexp
	phoneRegex *regexp.Regexp
	// Company names are injected per-request (known from candidate profile).
}

var defaultRedactor = &Redactor{
	emailRegex: regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`),
	phoneRegex: regexp.MustCompile(`(\+91[\-\s]?)?[6-9]\d{9}`),
}

// Redact replaces known PII in the input text.
// companyNames: list of company names from the candidate's profile.
func Redact(text string, companyNames []string) string {
	counter := &tokenCounter{}

	// Replace emails
	text = defaultRedactor.emailRegex.ReplaceAllStringFunc(text, func(_ string) string {
		return counter.next("EMAIL")
	})

	// Replace Indian phone numbers
	text = defaultRedactor.phoneRegex.ReplaceAllStringFunc(text, func(_ string) string {
		return counter.next("PHONE")
	})

	// Replace known company names (case-insensitive)
	for _, company := range companyNames {
		if company == "" {
			continue
		}
		placeholder := counter.next("COMPANY")
		text = strings.ReplaceAll(text, company, placeholder)
		// Also replace common capitalisation variants
		text = strings.ReplaceAll(text, strings.ToUpper(company), placeholder)
		text = strings.ReplaceAll(text, strings.ToLower(company), placeholder)
	}

	return text
}

type tokenCounter struct {
	counts map[string]int
}

func (c *tokenCounter) next(prefix string) string {
	if c.counts == nil {
		c.counts = make(map[string]int)
	}
	c.counts[prefix]++
	return fmt.Sprintf("[%s_%d]", prefix, c.counts[prefix])
}
