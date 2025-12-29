package pii

import (
	"regexp"
	"strings"

	"github.com/epps11/goguard/internal/models"
)

// Masker handles PII detection and masking
type Masker struct {
	patterns       map[string]*regexp.Regexp
	enabled        bool
	maskChar       string
	preserveDomain bool
	enabledTypes   map[string]bool
}

// NewMasker creates a new PII masker
func NewMasker(piiTypes []string, maskChar string, preserveDomain, enabled bool) *Masker {
	m := &Masker{
		patterns:       make(map[string]*regexp.Regexp),
		enabled:        enabled,
		maskChar:       maskChar,
		preserveDomain: preserveDomain,
		enabledTypes:   make(map[string]bool),
	}

	// Enable specified PII types
	for _, t := range piiTypes {
		m.enabledTypes[t] = true
	}

	// Define PII patterns
	piiPatterns := map[string]string{
		// Email addresses
		"email": `[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`,

		// Phone numbers (various formats)
		"phone": `(?:\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}`,

		// Social Security Numbers
		"ssn": `\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b`,

		// Credit card numbers (major providers)
		"credit_card": `\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b`,

		// IP addresses (IPv4)
		"ip_address": `\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`,

		// IPv6 addresses
		"ipv6_address": `\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`,

		// Dates of birth (various formats)
		"date_of_birth": `\b(?:0?[1-9]|1[0-2])[/\-](?:0?[1-9]|[12][0-9]|3[01])[/\-](?:19|20)\d{2}\b`,

		// US Passport numbers
		"passport": `\b[A-Z]{1,2}[0-9]{6,9}\b`,

		// Driver's license (generic pattern)
		"drivers_license": `\b[A-Z]{1,2}[0-9]{5,8}\b`,

		// Bank account numbers (generic)
		"bank_account": `\b[0-9]{8,17}\b`,

		// Routing numbers
		"routing_number": `\b[0-9]{9}\b`,

		// AWS access keys
		"aws_key": `\bAKIA[0-9A-Z]{16}\b`,

		// AWS secret keys
		"aws_secret": `\b[A-Za-z0-9/+=]{40}\b`,

		// API keys (generic pattern)
		"api_key": `\b[a-zA-Z0-9_\-]{32,64}\b`,

		// Names (basic pattern - first last)
		"name": `\b[A-Z][a-z]+\s+[A-Z][a-z]+\b`,

		// Street addresses
		"address": `\b\d{1,5}\s+[A-Za-z]+\s+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Way|Circle|Cir)\b`,

		// ZIP codes
		"zip_code": `\b\d{5}(?:-\d{4})?\b`,

		// Medical record numbers (generic)
		"medical_record": `\bMRN[:\s]?[0-9]{6,10}\b`,

		// Health insurance IDs
		"health_insurance_id": `\b[A-Z]{3}[0-9]{9}\b`,
	}

	// Compile enabled patterns
	for name, pattern := range piiPatterns {
		if m.enabledTypes[name] || len(piiTypes) == 0 {
			if re, err := regexp.Compile(pattern); err == nil {
				m.patterns[name] = re
			}
		}
	}

	return m
}

// Mask processes messages and masks detected PII
func (m *Masker) Mask(messages []models.Message) ([]models.Message, *models.PIIReport) {
	report := &models.PIIReport{
		PIIDetected: false,
		PIICount:    0,
		PIITypes:    []models.PIIMatch{},
		MaskedCount: 0,
	}

	if !m.enabled {
		return messages, report
	}

	maskedMessages := make([]models.Message, len(messages))

	for i, msg := range messages {
		maskedContent, matches := m.maskContent(msg.Content, formatLocation(i, msg.Role))
		maskedMessages[i] = models.Message{
			Role:    msg.Role,
			Content: maskedContent,
		}
		report.PIITypes = append(report.PIITypes, matches...)
	}

	report.PIICount = len(report.PIITypes)
	report.PIIDetected = report.PIICount > 0
	report.MaskedCount = report.PIICount

	return maskedMessages, report
}

// maskContent masks PII in a single content string
func (m *Masker) maskContent(content, location string) (string, []models.PIIMatch) {
	matches := []models.PIIMatch{}
	result := content

	for piiType, pattern := range m.patterns {
		allMatches := pattern.FindAllStringIndex(result, -1)

		// Process matches in reverse order to maintain positions
		for i := len(allMatches) - 1; i >= 0; i-- {
			match := allMatches[i]
			start, end := match[0], match[1]
			originalValue := result[start:end]

			// Skip if it looks like a false positive
			if m.isFalsePositive(piiType, originalValue) {
				continue
			}

			maskedValue := m.generateMask(piiType, originalValue)

			piiMatch := models.PIIMatch{
				Type:          piiType,
				OriginalValue: originalValue,
				MaskedValue:   maskedValue,
				Location:      location,
				StartPosition: start,
				EndPosition:   end,
			}
			matches = append(matches, piiMatch)

			// Replace in result
			result = result[:start] + maskedValue + result[end:]
		}
	}

	return result, matches
}

// generateMask creates a masked version of the PII
func (m *Masker) generateMask(piiType, original string) string {
	maskChar := m.maskChar
	if maskChar == "" {
		maskChar = "*"
	}

	switch piiType {
	case "email":
		if m.preserveDomain {
			parts := strings.Split(original, "@")
			if len(parts) == 2 {
				masked := strings.Repeat(maskChar, len(parts[0]))
				return masked + "@" + parts[1]
			}
		}
		return strings.Repeat(maskChar, len(original))

	case "phone":
		// Keep last 4 digits visible
		if len(original) >= 4 {
			return strings.Repeat(maskChar, len(original)-4) + original[len(original)-4:]
		}
		return strings.Repeat(maskChar, len(original))

	case "ssn":
		// Show only last 4 digits
		cleaned := strings.ReplaceAll(strings.ReplaceAll(original, "-", ""), " ", "")
		if len(cleaned) >= 4 {
			return "***-**-" + cleaned[len(cleaned)-4:]
		}
		return strings.Repeat(maskChar, len(original))

	case "credit_card":
		// Show only last 4 digits
		cleaned := strings.ReplaceAll(original, " ", "")
		if len(cleaned) >= 4 {
			return strings.Repeat(maskChar, len(cleaned)-4) + cleaned[len(cleaned)-4:]
		}
		return strings.Repeat(maskChar, len(original))

	case "ip_address", "ipv6_address":
		// Mask middle octets
		return "[MASKED_IP]"

	case "aws_key", "aws_secret", "api_key":
		// Show only first 4 characters
		if len(original) > 4 {
			return original[:4] + strings.Repeat(maskChar, len(original)-4)
		}
		return strings.Repeat(maskChar, len(original))

	default:
		// Default: mask entirely
		return "[MASKED_" + strings.ToUpper(piiType) + "]"
	}
}

// isFalsePositive checks for common false positives
func (m *Masker) isFalsePositive(piiType, value string) bool {
	switch piiType {
	case "phone":
		// Skip if it's likely a version number or ID
		if strings.HasPrefix(value, "v") || strings.HasPrefix(value, "V") {
			return true
		}
	case "ssn":
		// Skip common non-SSN patterns
		if value == "000-00-0000" || value == "123-45-6789" {
			return true
		}
	case "bank_account", "routing_number":
		// Skip if too short or all same digits
		if len(strings.ReplaceAll(value, " ", "")) < 8 {
			return true
		}
		allSame := true
		for i := 1; i < len(value); i++ {
			if value[i] != value[0] {
				allSame = false
				break
			}
		}
		if allSame {
			return true
		}
	case "name":
		// Skip common words that match name pattern
		commonWords := []string{"Hello World", "Lorem Ipsum", "Foo Bar", "Test User"}
		for _, w := range commonWords {
			if strings.EqualFold(value, w) {
				return true
			}
		}
	case "zip_code":
		// Skip if it's likely not a ZIP (e.g., year)
		if len(value) == 4 && (strings.HasPrefix(value, "19") || strings.HasPrefix(value, "20")) {
			return true
		}
	}
	return false
}

func formatLocation(index int, role string) string {
	return strings.ToLower(role) + "_message_" + string(rune('0'+index))
}

// Analyze detects PII without masking (for reporting only)
func (m *Masker) Analyze(messages []models.Message) *models.PIIReport {
	report := &models.PIIReport{
		PIIDetected: false,
		PIICount:    0,
		PIITypes:    []models.PIIMatch{},
	}

	if !m.enabled {
		return report
	}

	for i, msg := range messages {
		_, matches := m.maskContent(msg.Content, formatLocation(i, msg.Role))
		report.PIITypes = append(report.PIITypes, matches...)
	}

	report.PIICount = len(report.PIITypes)
	report.PIIDetected = report.PIICount > 0

	return report
}
