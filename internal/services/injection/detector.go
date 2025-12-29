package injection

import (
	"regexp"
	"strings"

	"github.com/epps11/goguard/internal/models"
)

// Detector handles prompt injection detection
type Detector struct {
	patterns         []*regexp.Regexp
	keywordPatterns  []string
	enabled          bool
	blockOnDetection bool
}

// NewDetector creates a new injection detector
func NewDetector(customPatterns []string, enabled, blockOnDetection bool) *Detector {
	d := &Detector{
		enabled:          enabled,
		blockOnDetection: blockOnDetection,
	}

	// Default injection patterns
	defaultPatterns := []string{
		// Direct instruction override attempts
		`(?i)ignore\s+(all\s+)?(previous|prior|above)\s+(instructions?|prompts?|rules?)`,
		`(?i)disregard\s+(all\s+)?(previous|prior|above)\s+(instructions?|prompts?|rules?)`,
		`(?i)forget\s+(all\s+)?(previous|prior|above)\s+(instructions?|prompts?|rules?)`,
		`(?i)override\s+(all\s+)?(previous|prior|above)\s+(instructions?|prompts?|rules?)`,

		// Role manipulation
		`(?i)you\s+are\s+now\s+(a|an|the)\s+`,
		`(?i)act\s+as\s+(a|an|if\s+you\s+were)`,
		`(?i)pretend\s+(to\s+be|you\s+are)`,
		`(?i)roleplay\s+as`,
		`(?i)simulate\s+(being|a)`,

		// System prompt extraction
		`(?i)(show|reveal|display|print|output|tell\s+me)\s+(your|the)\s+(system\s+)?(prompt|instructions?)`,
		`(?i)what\s+(are|is)\s+your\s+(system\s+)?(prompt|instructions?)`,
		`(?i)repeat\s+(your|the)\s+(system\s+)?(prompt|instructions?)`,

		// Jailbreak attempts
		`(?i)DAN\s+(mode|prompt)`,
		`(?i)developer\s+mode`,
		`(?i)jailbreak`,
		`(?i)bypass\s+(safety|filter|restriction)`,
		`(?i)disable\s+(safety|filter|restriction)`,
		`(?i)remove\s+(all\s+)?(safety|filter|restriction)`,

		// Code injection markers
		`(?i)<\|im_start\|>`,
		`(?i)<\|im_end\|>`,
		`(?i)\[INST\]`,
		`(?i)\[/INST\]`,
		`(?i)<<SYS>>`,
		`(?i)<</SYS>>`,

		// Data exfiltration attempts
		`(?i)(send|transmit|exfiltrate|leak)\s+(data|information|secrets?)`,
		`(?i)make\s+(a|an)\s+(http|api|web)\s+(request|call)`,

		// Delimiter injection
		`(?i)###\s*(system|instruction|prompt)`,
		`(?i)---\s*(system|instruction|prompt)`,

		// Encoding bypass attempts
		`(?i)base64\s+(decode|encode)`,
		`(?i)hex\s+(decode|encode)`,
		`(?i)rot13`,
	}

	// Compile default patterns
	for _, p := range defaultPatterns {
		if re, err := regexp.Compile(p); err == nil {
			d.patterns = append(d.patterns, re)
		}
	}

	// Compile custom patterns
	for _, p := range customPatterns {
		if re, err := regexp.Compile(p); err == nil {
			d.patterns = append(d.patterns, re)
		}
	}

	// Keyword-based detection (case-insensitive substring matching)
	d.keywordPatterns = []string{
		"ignore previous",
		"ignore all instructions",
		"disregard your instructions",
		"new instructions:",
		"system prompt:",
		"[system]",
		"<system>",
		"</system>",
		"assistant:",
		"human:",
		"user:",
	}

	return d
}

// Analyze checks messages for injection attempts
func (d *Detector) Analyze(messages []models.Message) *models.SecurityReport {
	report := &models.SecurityReport{
		InjectionDetected: false,
		ThreatLevel:       "none",
		Detections:        []models.Detection{},
		Recommendations:   []string{},
	}

	if !d.enabled {
		return report
	}

	for i, msg := range messages {
		// Skip system messages - they're trusted
		if msg.Role == "system" {
			continue
		}

		content := msg.Content
		location := formatLocation(i, msg.Role)

		// Check regex patterns
		for _, pattern := range d.patterns {
			if matches := pattern.FindStringSubmatch(content); len(matches) > 0 {
				detection := models.Detection{
					Type:        categorizePattern(pattern.String()),
					Pattern:     pattern.String(),
					Location:    location,
					Confidence:  0.85,
					Description: "Regex pattern match detected",
				}
				report.Detections = append(report.Detections, detection)
			}
		}

		// Check keyword patterns
		lowerContent := strings.ToLower(content)
		for _, keyword := range d.keywordPatterns {
			if strings.Contains(lowerContent, keyword) {
				detection := models.Detection{
					Type:        "keyword_match",
					Pattern:     keyword,
					Location:    location,
					Confidence:  0.7,
					Description: "Suspicious keyword detected",
				}
				report.Detections = append(report.Detections, detection)
			}
		}

		// Check for suspicious character sequences
		if hasSuspiciousSequences(content) {
			detection := models.Detection{
				Type:        "suspicious_encoding",
				Pattern:     "special_characters",
				Location:    location,
				Confidence:  0.6,
				Description: "Suspicious character sequences detected",
			}
			report.Detections = append(report.Detections, detection)
		}
	}

	// Calculate threat level based on detections
	report.InjectionDetected = len(report.Detections) > 0
	report.ThreatLevel = calculateThreatLevel(report.Detections)

	if report.InjectionDetected {
		report.Recommendations = generateRecommendations(report.Detections)
		if d.blockOnDetection && report.ThreatLevel != "low" {
			report.BlockedReason = "Potential prompt injection detected"
		}
	}

	return report
}

// ShouldBlock returns true if the request should be blocked
func (d *Detector) ShouldBlock(report *models.SecurityReport) bool {
	if !d.blockOnDetection {
		return false
	}
	return report.ThreatLevel == "high" || report.ThreatLevel == "critical"
}

func formatLocation(index int, role string) string {
	return strings.ToLower(role) + "_message_" + string(rune('0'+index))
}

func categorizePattern(pattern string) string {
	lowerPattern := strings.ToLower(pattern)
	switch {
	case strings.Contains(lowerPattern, "ignore") || strings.Contains(lowerPattern, "disregard"):
		return "instruction_override"
	case strings.Contains(lowerPattern, "you are now") || strings.Contains(lowerPattern, "act as"):
		return "role_manipulation"
	case strings.Contains(lowerPattern, "prompt") || strings.Contains(lowerPattern, "instruction"):
		return "prompt_extraction"
	case strings.Contains(lowerPattern, "jailbreak") || strings.Contains(lowerPattern, "bypass"):
		return "jailbreak_attempt"
	case strings.Contains(lowerPattern, "im_start") || strings.Contains(lowerPattern, "inst"):
		return "delimiter_injection"
	case strings.Contains(lowerPattern, "send") || strings.Contains(lowerPattern, "exfiltrate"):
		return "data_exfiltration"
	default:
		return "unknown"
	}
}

func hasSuspiciousSequences(content string) bool {
	suspiciousPatterns := []string{
		"\u200b", // zero-width space
		"\u200c", // zero-width non-joiner
		"\u200d", // zero-width joiner
		"\ufeff", // BOM
		"\u202e", // right-to-left override
	}

	for _, p := range suspiciousPatterns {
		if strings.Contains(content, p) {
			return true
		}
	}
	return false
}

func calculateThreatLevel(detections []models.Detection) string {
	if len(detections) == 0 {
		return "none"
	}

	maxConfidence := 0.0
	criticalTypes := map[string]bool{
		"jailbreak_attempt":   true,
		"data_exfiltration":   true,
		"delimiter_injection": true,
	}

	hasCritical := false
	for _, d := range detections {
		if d.Confidence > maxConfidence {
			maxConfidence = d.Confidence
		}
		if criticalTypes[d.Type] {
			hasCritical = true
		}
	}

	switch {
	case hasCritical || len(detections) >= 3:
		return "critical"
	case maxConfidence >= 0.85:
		return "high"
	case maxConfidence >= 0.7:
		return "medium"
	default:
		return "low"
	}
}

func generateRecommendations(detections []models.Detection) []string {
	recommendations := []string{}
	typesSeen := make(map[string]bool)

	for _, d := range detections {
		if typesSeen[d.Type] {
			continue
		}
		typesSeen[d.Type] = true

		switch d.Type {
		case "instruction_override":
			recommendations = append(recommendations, "Review input for attempts to override system instructions")
		case "role_manipulation":
			recommendations = append(recommendations, "Input attempts to manipulate AI role/persona")
		case "prompt_extraction":
			recommendations = append(recommendations, "Input attempts to extract system prompt")
		case "jailbreak_attempt":
			recommendations = append(recommendations, "Known jailbreak pattern detected - block recommended")
		case "delimiter_injection":
			recommendations = append(recommendations, "Special delimiter tokens detected - potential injection")
		case "data_exfiltration":
			recommendations = append(recommendations, "Potential data exfiltration attempt detected")
		}
	}

	return recommendations
}
