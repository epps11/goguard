package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/epps11/goguard/internal/models"
	"github.com/google/uuid"
)

// Repository provides database operations
type Repository struct {
	db *DB
}

// NewRepository creates a new repository
func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

// User operations

func (r *Repository) CreateUser(ctx context.Context, user *models.User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()

	groupsJSON, _ := json.Marshal(user.Groups)
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, email, name, role, status, groups, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.Email, user.Name, user.Role, user.Status, groupsJSON, metadataJSON, user.CreatedAt)
	return err
}

func (r *Repository) GetUser(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	var groupsJSON, metadataJSON []byte
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, name, role, status, groups, metadata, created_at, last_login_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.Status,
		&groupsJSON, &metadataJSON, &user.CreatedAt, &lastLoginAt)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(groupsJSON, &user.Groups)
	json.Unmarshal(metadataJSON, &user.Metadata)
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (r *Repository) ListUsers(ctx context.Context) ([]*models.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, email, name, role, status, groups, metadata, created_at, last_login_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var groupsJSON, metadataJSON []byte
		var lastLoginAt sql.NullTime

		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.Status,
			&groupsJSON, &metadataJSON, &user.CreatedAt, &lastLoginAt); err != nil {
			return nil, err
		}

		json.Unmarshal(groupsJSON, &user.Groups)
		json.Unmarshal(metadataJSON, &user.Metadata)
		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}
		users = append(users, &user)
	}
	return users, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *models.User) error {
	groupsJSON, _ := json.Marshal(user.Groups)
	metadataJSON, _ := json.Marshal(user.Metadata)

	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET email = $2, name = $3, role = $4, status = $5,
		groups = $6, metadata = $7, updated_at = NOW()
		WHERE id = $1
	`, user.ID, user.Email, user.Name, user.Role, user.Status, groupsJSON, metadataJSON)
	return err
}

func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

// Policy operations

func (r *Repository) CreatePolicy(ctx context.Context, policy *models.Policy) error {
	policy.ID = uuid.New().String()
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	configJSON, _ := json.Marshal(policy.Config)
	rulesJSON, _ := json.Marshal(policy.Rules)
	targetsJSON, _ := json.Marshal(policy.Targets)
	actionsJSON, _ := json.Marshal(policy.Actions)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO policies (id, name, description, type, status, priority, config, rules, targets, actions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, policy.ID, policy.Name, policy.Description, policy.Type, policy.Status, policy.Priority,
		configJSON, rulesJSON, targetsJSON, actionsJSON, policy.CreatedAt, policy.UpdatedAt)
	return err
}

func (r *Repository) GetPolicy(ctx context.Context, id string) (*models.Policy, error) {
	var policy models.Policy
	var configJSON, rulesJSON, targetsJSON, actionsJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, type, status, priority, config, rules, targets, actions, created_at, updated_at
		FROM policies WHERE id = $1
	`, id).Scan(&policy.ID, &policy.Name, &policy.Description, &policy.Type, &policy.Status,
		&policy.Priority, &configJSON, &rulesJSON, &targetsJSON, &actionsJSON, &policy.CreatedAt, &policy.UpdatedAt)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(configJSON, &policy.Config)
	json.Unmarshal(rulesJSON, &policy.Rules)
	json.Unmarshal(targetsJSON, &policy.Targets)
	json.Unmarshal(actionsJSON, &policy.Actions)

	return &policy, nil
}

func (r *Repository) ListPolicies(ctx context.Context) ([]*models.Policy, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, type, status, priority, config, rules, targets, actions, created_at, updated_at
		FROM policies ORDER BY priority ASC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*models.Policy
	for rows.Next() {
		var policy models.Policy
		var configJSON, rulesJSON, targetsJSON, actionsJSON []byte

		if err := rows.Scan(&policy.ID, &policy.Name, &policy.Description, &policy.Type, &policy.Status,
			&policy.Priority, &configJSON, &rulesJSON, &targetsJSON, &actionsJSON, &policy.CreatedAt, &policy.UpdatedAt); err != nil {
			return nil, err
		}

		json.Unmarshal(configJSON, &policy.Config)
		json.Unmarshal(rulesJSON, &policy.Rules)
		json.Unmarshal(targetsJSON, &policy.Targets)
		json.Unmarshal(actionsJSON, &policy.Actions)
		policies = append(policies, &policy)
	}
	return policies, nil
}

func (r *Repository) UpdatePolicy(ctx context.Context, policy *models.Policy) error {
	policy.UpdatedAt = time.Now()
	configJSON, _ := json.Marshal(policy.Config)
	rulesJSON, _ := json.Marshal(policy.Rules)
	targetsJSON, _ := json.Marshal(policy.Targets)
	actionsJSON, _ := json.Marshal(policy.Actions)

	_, err := r.db.ExecContext(ctx, `
		UPDATE policies SET name = $2, description = $3, type = $4, status = $5, priority = $6,
		config = $7, rules = $8, targets = $9, actions = $10, updated_at = $11
		WHERE id = $1
	`, policy.ID, policy.Name, policy.Description, policy.Type, policy.Status, policy.Priority,
		configJSON, rulesJSON, targetsJSON, actionsJSON, policy.UpdatedAt)
	return err
}

func (r *Repository) DeletePolicy(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM policies WHERE id = $1`, id)
	return err
}

// SpendingLimit operations

func (r *Repository) CreateSpendingLimit(ctx context.Context, limit *models.SpendingLimit) error {
	limit.ID = uuid.New().String()
	limit.CreatedAt = time.Now()
	limit.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO spending_limits (id, user_id, limit_type, limit_amount, current_spend, currency, reset_at, alert_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, limit.ID, limit.UserID, limit.LimitType, limit.LimitAmount, limit.CurrentSpend,
		limit.Currency, limit.ResetAt, limit.AlertAt, limit.CreatedAt, limit.UpdatedAt)
	return err
}

func (r *Repository) GetSpendingLimit(ctx context.Context, id string) (*models.SpendingLimit, error) {
	var limit models.SpendingLimit
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, limit_type, limit_amount, current_spend, currency, reset_at, alert_at, created_at, updated_at
		FROM spending_limits WHERE id = $1
	`, id).Scan(&limit.ID, &limit.UserID, &limit.LimitType, &limit.LimitAmount,
		&limit.CurrentSpend, &limit.Currency, &limit.ResetAt, &limit.AlertAt,
		&limit.CreatedAt, &limit.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

func (r *Repository) ListSpendingLimits(ctx context.Context) ([]*models.SpendingLimit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, limit_type, limit_amount, current_spend, currency, reset_at, alert_at, created_at, updated_at
		FROM spending_limits ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var limits []*models.SpendingLimit
	for rows.Next() {
		var limit models.SpendingLimit
		if err := rows.Scan(&limit.ID, &limit.UserID, &limit.LimitType, &limit.LimitAmount,
			&limit.CurrentSpend, &limit.Currency, &limit.ResetAt, &limit.AlertAt,
			&limit.CreatedAt, &limit.UpdatedAt); err != nil {
			return nil, err
		}
		limits = append(limits, &limit)
	}
	return limits, nil
}

func (r *Repository) UpdateSpendingLimit(ctx context.Context, limit *models.SpendingLimit) error {
	limit.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, `
		UPDATE spending_limits SET user_id = $2, limit_type = $3, limit_amount = $4,
		current_spend = $5, currency = $6, reset_at = $7, alert_at = $8, updated_at = $9
		WHERE id = $1
	`, limit.ID, limit.UserID, limit.LimitType, limit.LimitAmount, limit.CurrentSpend,
		limit.Currency, limit.ResetAt, limit.AlertAt, limit.UpdatedAt)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no spending limit found with id: %s", limit.ID)
	}
	return nil
}

// AuditLog operations

func (r *Repository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	log.ID = uuid.New().String()
	log.Timestamp = time.Now()

	detailsJSON, _ := json.Marshal(log.Details)
	durationMs := int(log.Duration.Milliseconds())

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_logs (id, request_id, event_type, action, user_id, user_email, resource_type, resource_id, status, ip_address, user_agent, duration_ms, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, log.ID, log.RequestID, log.EventType, log.Action, log.UserID, log.UserEmail,
		log.ResourceType, log.ResourceID, log.Status, log.IPAddress, log.UserAgent,
		durationMs, detailsJSON, log.Timestamp)
	return err
}

func (r *Repository) ListAuditLogs(ctx context.Context, limit int) ([]*models.AuditLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, request_id, event_type, action, user_id, user_email, resource_type, resource_id, status, ip_address, user_agent, duration_ms, details, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var detailsJSON []byte
		var durationMs int

		if err := rows.Scan(&log.ID, &log.RequestID, &log.EventType, &log.Action, &log.UserID,
			&log.UserEmail, &log.ResourceType, &log.ResourceID, &log.Status, &log.IPAddress,
			&log.UserAgent, &durationMs, &detailsJSON, &log.Timestamp); err != nil {
			return nil, err
		}

		log.Duration = time.Duration(durationMs) * time.Millisecond
		json.Unmarshal(detailsJSON, &log.Details)
		logs = append(logs, &log)
	}
	return logs, nil
}

// Settings operations

func (r *Repository) GetSetting(ctx context.Context, key string) (interface{}, error) {
	var valueJSON []byte
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = $1`, key).Scan(&valueJSON)
	if err != nil {
		return nil, err
	}

	var value interface{}
	json.Unmarshal(valueJSON, &value)
	return value, nil
}

func (r *Repository) SetSetting(ctx context.Context, key string, value interface{}) error {
	valueJSON, _ := json.Marshal(value)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()
	`, key, valueJSON)
	return err
}

func (r *Repository) GetAllSettings(ctx context.Context) (map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]interface{})
	for rows.Next() {
		var key string
		var valueJSON []byte
		if err := rows.Scan(&key, &valueJSON); err != nil {
			return nil, err
		}
		var value interface{}
		json.Unmarshal(valueJSON, &value)
		settings[key] = value
	}
	return settings, nil
}
