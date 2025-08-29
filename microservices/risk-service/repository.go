package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQLRepository implements Repository interface with PostgreSQL
type PostgreSQLRepository struct {
	db *sql.DB
}

// NewPostgreSQLRepository creates a new PostgreSQL repository
func NewPostgreSQLRepository(connectionString string) (*PostgreSQLRepository, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgreSQLRepository{db: db}, nil
}

// CreateRiskProfile creates a new risk profile
func (r *PostgreSQLRepository) CreateRiskProfile(ctx context.Context, profile *RiskProfile) error {
	query := `
		INSERT INTO risk_profiles (
			id, entity_id, entity_type, risk_score, risk_level, 
			factors, rules_applied, last_assessment, metadata, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	factorsJSON, err := json.Marshal(profile.Factors)
	if err != nil {
		return fmt.Errorf("failed to marshal factors: %w", err)
	}

	rulesJSON, err := json.Marshal(profile.RulesApplied)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	metadataJSON, err := json.Marshal(profile.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		profile.ID,
		profile.EntityID,
		profile.EntityType,
		profile.RiskScore,
		profile.RiskLevel,
		factorsJSON,
		rulesJSON,
		profile.LastAssessment,
		metadataJSON,
		profile.CreatedAt,
		profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create risk profile: %w", err)
	}

	return nil
}

// GetRiskProfile retrieves a risk profile by entity ID
func (r *PostgreSQLRepository) GetRiskProfile(ctx context.Context, entityID string) (*RiskProfile, error) {
	query := `
		SELECT id, entity_id, entity_type, risk_score, risk_level, 
			   factors, rules_applied, last_assessment, metadata, 
			   created_at, updated_at
		FROM risk_profiles 
		WHERE entity_id = $1`

	var profile RiskProfile
	var factorsJSON, rulesJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, entityID).Scan(
		&profile.ID,
		&profile.EntityID,
		&profile.EntityType,
		&profile.RiskScore,
		&profile.RiskLevel,
		&factorsJSON,
		&rulesJSON,
		&profile.LastAssessment,
		&metadataJSON,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("risk profile not found: %s", entityID)
		}
		return nil, fmt.Errorf("failed to get risk profile: %w", err)
	}

	// Parse JSON fields
	if len(factorsJSON) > 0 {
		if err := json.Unmarshal(factorsJSON, &profile.Factors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal factors: %w", err)
		}
	}

	if len(rulesJSON) > 0 {
		if err := json.Unmarshal(rulesJSON, &profile.RulesApplied); err != nil {
			return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
		}
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &profile.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &profile, nil
}

// UpdateRiskProfile updates an existing risk profile
func (r *PostgreSQLRepository) UpdateRiskProfile(ctx context.Context, profile *RiskProfile) error {
	query := `
		UPDATE risk_profiles 
		SET entity_type = $2, risk_score = $3, risk_level = $4, 
			factors = $5, rules_applied = $6, last_assessment = $7, 
			metadata = $8, updated_at = $9
		WHERE entity_id = $1`

	factorsJSON, err := json.Marshal(profile.Factors)
	if err != nil {
		return fmt.Errorf("failed to marshal factors: %w", err)
	}

	rulesJSON, err := json.Marshal(profile.RulesApplied)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	metadataJSON, err := json.Marshal(profile.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	profile.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		profile.EntityID,
		profile.EntityType,
		profile.RiskScore,
		profile.RiskLevel,
		factorsJSON,
		rulesJSON,
		profile.LastAssessment,
		metadataJSON,
		profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update risk profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("risk profile not found: %s", profile.EntityID)
	}

	return nil
}

// CreateRiskRule creates a new risk rule
func (r *PostgreSQLRepository) CreateRiskRule(ctx context.Context, rule *RiskRule) error {
	query := `
		INSERT INTO risk_rules (
			id, name, description, rule_type, conditions, 
			actions, priority, enabled, metadata, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	metadataJSON, err := json.Marshal(rule.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		rule.ID,
		rule.Name,
		rule.Description,
		rule.RuleType,
		conditionsJSON,
		actionsJSON,
		rule.Priority,
		rule.IsActive,
		metadataJSON,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create risk rule: %w", err)
	}

	return nil
}

// GetRiskRules retrieves all active risk rules
func (r *PostgreSQLRepository) GetRiskRules(ctx context.Context) ([]*RiskRule, error) {
	query := `
		SELECT id, name, description, rule_type, conditions, 
			   actions, priority, enabled, metadata, 
			   created_at, updated_at
		FROM risk_rules 
		WHERE enabled = true
		ORDER BY priority DESC, created_at ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk rules: %w", err)
	}
	defer rows.Close()

	var rules []*RiskRule
	for rows.Next() {
		var rule RiskRule
		var conditionsJSON, actionsJSON, metadataJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.Description,
			&rule.RuleType,
			&conditionsJSON,
			&actionsJSON,
			&rule.Priority,
			&rule.IsActive,
			&metadataJSON,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan risk rule: %w", err)
		}

		// Parse JSON fields
		if len(conditionsJSON) > 0 {
			if err := json.Unmarshal(conditionsJSON, &rule.Conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
		}

		if len(actionsJSON) > 0 {
			if err := json.Unmarshal(actionsJSON, &rule.Actions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
			}
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &rule.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating risk rules: %w", err)
	}

	return rules, nil
}

// CreateAssessment creates a new risk assessment
func (r *PostgreSQLRepository) CreateAssessment(ctx context.Context, assessment *RiskAssessment) error {
	query := `
		INSERT INTO risk_assessments (
			id, entity_id, entity_type, assessment_type, 
			risk_score, risk_level, factors, rules_triggered, 
			recommendations, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	factorsJSON, err := json.Marshal(assessment.Factors)
	if err != nil {
		return fmt.Errorf("failed to marshal factors: %w", err)
	}

	rulesJSON, err := json.Marshal(assessment.RulesApplied)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	// Remove recommendations field as it doesn't exist in our type
	recommendationsJSON := []byte("[]")

	metadataJSON, err := json.Marshal(assessment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		assessment.ID,
		assessment.EntityID,
		assessment.EntityType,
		"standard", // Default assessment type since field doesn't exist
		assessment.RiskScore,
		assessment.RiskLevel,
		factorsJSON,
		rulesJSON,
		recommendationsJSON,
		metadataJSON,
		assessment.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create risk assessment: %w", err)
	}

	return nil
}

// GetAssessments retrieves assessments for an entity
func (r *PostgreSQLRepository) GetAssessments(ctx context.Context, entityID string) ([]*RiskAssessment, error) {
	query := `
		SELECT id, entity_id, entity_type, assessment_type, 
			   risk_score, risk_level, factors, rules_triggered, 
			   recommendations, metadata, created_at
		FROM risk_assessments 
		WHERE entity_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assessments: %w", err)
	}
	defer rows.Close()

	var assessments []*RiskAssessment
	for rows.Next() {
		var assessment RiskAssessment
		var factorsJSON, rulesJSON, recommendationsJSON, metadataJSON []byte

		var assessmentType string // Temporary variable for unused field
		err := rows.Scan(
			&assessment.ID,
			&assessment.EntityID,
			&assessment.EntityType,
			&assessmentType,
			&assessment.RiskScore,
			&assessment.RiskLevel,
			&factorsJSON,
			&rulesJSON,
			&recommendationsJSON,
			&metadataJSON,
			&assessment.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan assessment: %w", err)
		}

		// Parse JSON fields
		if len(factorsJSON) > 0 {
			if err := json.Unmarshal(factorsJSON, &assessment.Factors); err != nil {
				return nil, fmt.Errorf("failed to unmarshal factors: %w", err)
			}
		}

		if len(rulesJSON) > 0 {
			if err := json.Unmarshal(rulesJSON, &assessment.RulesApplied); err != nil {
				return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
			}
		}

		// Skip recommendations field as it doesn't exist in our type

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &assessment.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		assessments = append(assessments, &assessment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating assessments: %w", err)
	}

	return assessments, nil
}

// Close closes the database connection
func (r *PostgreSQLRepository) Close() error {
	return r.db.Close()
}

// Health checks database connectivity
func (r *PostgreSQLRepository) Health(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
