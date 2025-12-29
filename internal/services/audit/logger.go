package audit

import (
	"context"
	"sync"
	"time"

	"github.com/epps11/goguard/internal/models"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Logger handles audit logging
type Logger struct {
	logs    []models.AuditLog
	alerts  []models.Alert
	mu      sync.RWMutex
	maxLogs int
}

// NewLogger creates a new audit logger
func NewLogger(maxLogs int) *Logger {
	if maxLogs <= 0 {
		maxLogs = 10000
	}
	return &Logger{
		logs:    make([]models.AuditLog, 0),
		alerts:  make([]models.Alert, 0),
		maxLogs: maxLogs,
	}
}

// Log creates a new audit log entry
func (l *Logger) Log(ctx context.Context, entry *models.AuditLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	l.logs = append(l.logs, *entry)

	// Trim old logs if exceeding max
	if len(l.logs) > l.maxLogs {
		l.logs = l.logs[len(l.logs)-l.maxLogs:]
	}

	log.Debug().
		Str("audit_id", entry.ID).
		Str("event_type", string(entry.EventType)).
		Str("action", entry.Action).
		Str("status", string(entry.Status)).
		Msg("Audit log created")

	return nil
}

// Query retrieves audit logs based on query parameters
func (l *Logger) Query(ctx context.Context, query *models.AuditQuery) ([]models.AuditLog, int, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var filtered []models.AuditLog

	for _, entry := range l.logs {
		if l.matchesQuery(&entry, query) {
			filtered = append(filtered, entry)
		}
	}

	total := len(filtered)

	// Apply pagination
	offset := query.Offset
	if offset > len(filtered) {
		offset = len(filtered)
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	// Reverse for newest first
	result := make([]models.AuditLog, 0, end-offset)
	for i := len(filtered) - 1 - offset; i >= len(filtered)-end && i >= 0; i-- {
		result = append(result, filtered[i])
	}

	return result, total, nil
}

func (l *Logger) matchesQuery(entry *models.AuditLog, query *models.AuditQuery) bool {
	if query.StartTime != nil && entry.Timestamp.Before(*query.StartTime) {
		return false
	}
	if query.EndTime != nil && entry.Timestamp.After(*query.EndTime) {
		return false
	}
	if query.UserID != "" && entry.UserID != query.UserID {
		return false
	}
	if query.ResourceType != "" && entry.ResourceType != query.ResourceType {
		return false
	}
	if query.Status != "" && entry.Status != query.Status {
		return false
	}
	if len(query.EventTypes) > 0 {
		found := false
		for _, et := range query.EventTypes {
			if entry.EventType == et {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// GetStats returns aggregated statistics
func (l *Logger) GetStats(ctx context.Context, period string) (*models.AuditStats, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var startTime time.Time
	now := time.Now()

	switch period {
	case "24h":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = now.Add(-30 * 24 * time.Hour)
	default:
		startTime = now.Add(-24 * time.Hour)
		period = "24h"
	}

	stats := &models.AuditStats{
		Period:         period,
		RequestsByHour: make(map[string]int64),
		EventsByType:   make(map[string]int64),
		TopUsers:       []models.UserStats{},
		TopModels:      []models.ModelStats{},
	}

	userStats := make(map[string]*models.UserStats)
	modelStats := make(map[string]*models.ModelStats)

	for _, entry := range l.logs {
		if entry.Timestamp.Before(startTime) {
			continue
		}

		stats.TotalRequests++
		stats.EventsByType[string(entry.EventType)]++

		switch entry.Status {
		case models.AuditStatusBlocked:
			stats.BlockedRequests++
		case models.AuditStatusSuccess:
			stats.AllowedRequests++
		case models.AuditStatusWarning:
			stats.WarningRequests++
		}

		// Track by hour
		hour := entry.Timestamp.Format("2006-01-02T15")
		stats.RequestsByHour[hour]++

		// Track user stats
		if entry.UserID != "" {
			if _, exists := userStats[entry.UserID]; !exists {
				userStats[entry.UserID] = &models.UserStats{
					UserID:    entry.UserID,
					UserEmail: entry.UserEmail,
				}
			}
			userStats[entry.UserID].RequestCount++

			if entry.Details != nil {
				if tokens, ok := entry.Details["total_tokens"].(float64); ok {
					userStats[entry.UserID].TokensUsed += int64(tokens)
					stats.TotalTokensUsed += int64(tokens)
				}
				if cost, ok := entry.Details["cost"].(float64); ok {
					userStats[entry.UserID].TotalCost += cost
					stats.TotalCost += cost
				}
			}
		}

		// Track model stats
		if entry.Details != nil {
			if model, ok := entry.Details["model"].(string); ok && model != "" {
				if _, exists := modelStats[model]; !exists {
					modelStats[model] = &models.ModelStats{
						Model: model,
					}
					if provider, ok := entry.Details["provider"].(string); ok {
						modelStats[model].Provider = provider
					}
				}
				modelStats[model].RequestCount++
			}
		}
	}

	// Count unique users
	stats.UniqueUsers = int64(len(userStats))

	// Convert maps to slices
	for _, us := range userStats {
		stats.TopUsers = append(stats.TopUsers, *us)
	}
	for _, ms := range modelStats {
		stats.TopModels = append(stats.TopModels, *ms)
	}

	return stats, nil
}

// GetDashboardMetrics returns metrics for the dashboard
func (l *Logger) GetDashboardMetrics(ctx context.Context) (*models.DashboardMetrics, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	now := time.Now()
	last24h := now.Add(-24 * time.Hour)
	prev24h := now.Add(-48 * time.Hour)

	metrics := &models.DashboardMetrics{
		Overview: models.OverviewMetrics{},
		Security: models.SecurityMetrics{
			ThreatsByLevel: make(map[string]int64),
			TopThreatTypes: make(map[string]int64),
		},
		Usage: models.UsageMetrics{
			RequestsByModel:    make(map[string]int64),
			RequestsByProvider: make(map[string]int64),
		},
		Spending: models.SpendingMetrics{
			SpendByUser:  make(map[string]float64),
			SpendByModel: make(map[string]float64),
		},
		RecentAlerts: l.getRecentAlerts(10),
		TopPolicies:  []models.PolicyMetric{},
	}

	var current24h, prev24hCount int64
	var currentUsers, prevUsers = make(map[string]bool), make(map[string]bool)
	var currentBlocked, prevBlocked int64
	var currentSpend, prevSpend float64

	for _, entry := range l.logs {
		if entry.Timestamp.After(last24h) {
			current24h++
			if entry.UserID != "" {
				currentUsers[entry.UserID] = true
			}
			if entry.Status == models.AuditStatusBlocked {
				currentBlocked++
			}

			// Security metrics
			if entry.EventType == models.EventTypeSecurityAlert {
				metrics.Security.InjectionAttempts24h++
				if entry.Details != nil {
					if level, ok := entry.Details["threat_level"].(string); ok {
						metrics.Security.ThreatsByLevel[level]++
					}
					if threatType, ok := entry.Details["threat_type"].(string); ok {
						metrics.Security.TopThreatTypes[threatType]++
					}
				}
			}

			// Usage metrics
			if entry.Details != nil {
				if tokens, ok := entry.Details["total_tokens"].(float64); ok {
					metrics.Usage.TotalTokens24h += int64(tokens)
				}
				if promptTokens, ok := entry.Details["prompt_tokens"].(float64); ok {
					metrics.Usage.PromptTokens24h += int64(promptTokens)
				}
				if completionTokens, ok := entry.Details["completion_tokens"].(float64); ok {
					metrics.Usage.CompletionTokens24h += int64(completionTokens)
				}
				if model, ok := entry.Details["model"].(string); ok {
					metrics.Usage.RequestsByModel[model]++
				}
				if provider, ok := entry.Details["provider"].(string); ok {
					metrics.Usage.RequestsByProvider[provider]++
				}
				if cost, ok := entry.Details["cost"].(float64); ok {
					currentSpend += cost
					if entry.UserID != "" {
						metrics.Spending.SpendByUser[entry.UserID] += cost
					}
					if model, ok := entry.Details["model"].(string); ok {
						metrics.Spending.SpendByModel[model] += cost
					}
				}
				if piiCount, ok := entry.Details["pii_count"].(float64); ok && piiCount > 0 {
					metrics.Security.PIIDetections24h += int64(piiCount)
				}
			}
		} else if entry.Timestamp.After(prev24h) {
			prev24hCount++
			if entry.UserID != "" {
				prevUsers[entry.UserID] = true
			}
			if entry.Status == models.AuditStatusBlocked {
				prevBlocked++
			}
			if entry.Details != nil {
				if cost, ok := entry.Details["cost"].(float64); ok {
					prevSpend += cost
				}
			}
		}
	}

	// Calculate overview metrics
	metrics.Overview.TotalRequests24h = current24h
	metrics.Overview.ActiveUsers24h = int64(len(currentUsers))
	metrics.Overview.BlockedRequests24h = currentBlocked
	metrics.Overview.TotalSpend24h = currentSpend

	// Calculate percentage changes
	if prev24hCount > 0 {
		metrics.Overview.RequestsChange = float64(current24h-prev24hCount) / float64(prev24hCount) * 100
	}
	if len(prevUsers) > 0 {
		metrics.Overview.UsersChange = float64(len(currentUsers)-len(prevUsers)) / float64(len(prevUsers)) * 100
	}
	if prevBlocked > 0 {
		metrics.Overview.BlockedChange = float64(currentBlocked-prevBlocked) / float64(prevBlocked) * 100
	}
	if prevSpend > 0 {
		metrics.Overview.SpendChange = (currentSpend - prevSpend) / prevSpend * 100
	}

	metrics.Spending.TotalSpendToday = currentSpend

	return metrics, nil
}

// CreateAlert creates a new alert
func (l *Logger) CreateAlert(ctx context.Context, alert *models.Alert) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	if alert.CreatedAt.IsZero() {
		alert.CreatedAt = time.Now()
	}

	l.alerts = append(l.alerts, *alert)

	log.Warn().
		Str("alert_id", alert.ID).
		Str("type", alert.Type).
		Str("severity", alert.Severity).
		Str("title", alert.Title).
		Msg("Alert created")

	return nil
}

// GetAlerts retrieves alerts
func (l *Logger) GetAlerts(ctx context.Context, limit int, includeAcked bool) ([]models.Alert, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var filtered []models.Alert
	for i := len(l.alerts) - 1; i >= 0 && len(filtered) < limit; i-- {
		alert := l.alerts[i]
		if includeAcked || alert.AckedAt == nil {
			filtered = append(filtered, alert)
		}
	}

	return filtered, nil
}

func (l *Logger) getRecentAlerts(limit int) []models.Alert {
	var recent []models.Alert
	for i := len(l.alerts) - 1; i >= 0 && len(recent) < limit; i-- {
		recent = append(recent, l.alerts[i])
	}
	return recent
}

// AckAlert acknowledges an alert
func (l *Logger) AckAlert(ctx context.Context, alertID, userID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i := range l.alerts {
		if l.alerts[i].ID == alertID {
			now := time.Now()
			l.alerts[i].AckedAt = &now
			l.alerts[i].AckedBy = userID
			return nil
		}
	}

	return nil
}
