package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/engagelab/captcha/internal/model"
)

// PGPolicyStore implements PolicyStore backed by PostgreSQL.
type PGPolicyStore struct {
	pool *pgxpool.Pool
}

// NewPGPolicyStore creates a new PostgreSQL-backed policy store.
func NewPGPolicyStore(pool *pgxpool.Pool) *PGPolicyStore {
	return &PGPolicyStore{pool: pool}
}

// CreatePolicy inserts a new policy into the database.
func (s *PGPolicyStore) CreatePolicy(p *model.Policy) error {
	ctx := context.Background()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO policies
		 (id, scene_type, threshold_low, threshold_high, action_low, action_mid, action_high,
		  ip_whitelist, ip_blacklist, rate_limit_rpm, rate_limit_rph, enabled, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		p.ID, string(p.SceneType), p.ThresholdLow, p.ThresholdHigh,
		string(p.ActionLow), string(p.ActionMid), string(p.ActionHigh),
		p.IPWhitelist, p.IPBlacklist, p.RateLimitRPM, p.RateLimitRPH,
		p.Enabled, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create policy: %w", err)
	}
	return nil
}

// GetPolicy retrieves a policy by its ID.
func (s *PGPolicyStore) GetPolicy(id string) (*model.Policy, error) {
	ctx := context.Background()
	return s.scanPolicy(s.pool.QueryRow(ctx,
		`SELECT id, scene_type, threshold_low, threshold_high, action_low, action_mid, action_high,
		        ip_whitelist, ip_blacklist, rate_limit_rpm, rate_limit_rph, enabled, created_at, updated_at
		 FROM policies WHERE id = $1`, id,
	))
}

// GetPolicyByScene retrieves the first enabled policy matching a scene type.
func (s *PGPolicyStore) GetPolicyByScene(sceneType model.SceneType) (*model.Policy, error) {
	ctx := context.Background()
	return s.scanPolicy(s.pool.QueryRow(ctx,
		`SELECT id, scene_type, threshold_low, threshold_high, action_low, action_mid, action_high,
		        ip_whitelist, ip_blacklist, rate_limit_rpm, rate_limit_rph, enabled, created_at, updated_at
		 FROM policies WHERE scene_type = $1 AND enabled = true
		 ORDER BY created_at DESC LIMIT 1`, string(sceneType),
	))
}

// ListPolicies returns all policies.
func (s *PGPolicyStore) ListPolicies() []*model.Policy {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx,
		`SELECT id, scene_type, threshold_low, threshold_high, action_low, action_mid, action_high,
		        ip_whitelist, ip_blacklist, rate_limit_rpm, rate_limit_rph, enabled, created_at, updated_at
		 FROM policies ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*model.Policy
	for rows.Next() {
		p, err := s.scanPolicyRow(rows)
		if err != nil {
			continue
		}
		result = append(result, p)
	}
	return result
}

func (s *PGPolicyStore) scanPolicy(row pgx.Row) (*model.Policy, error) {
	var p model.Policy
	var sceneType, actionLow, actionMid, actionHigh string
	err := row.Scan(
		&p.ID, &sceneType, &p.ThresholdLow, &p.ThresholdHigh,
		&actionLow, &actionMid, &actionHigh,
		&p.IPWhitelist, &p.IPBlacklist, &p.RateLimitRPM, &p.RateLimitRPH,
		&p.Enabled, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("policy not found: %w", err)
	}
	p.SceneType = model.SceneType(sceneType)
	p.ActionLow = model.RiskAction(actionLow)
	p.ActionMid = model.RiskAction(actionMid)
	p.ActionHigh = model.RiskAction(actionHigh)
	return &p, nil
}

func (s *PGPolicyStore) scanPolicyRow(rows pgx.Rows) (*model.Policy, error) {
	var p model.Policy
	var sceneType, actionLow, actionMid, actionHigh string
	err := rows.Scan(
		&p.ID, &sceneType, &p.ThresholdLow, &p.ThresholdHigh,
		&actionLow, &actionMid, &actionHigh,
		&p.IPWhitelist, &p.IPBlacklist, &p.RateLimitRPM, &p.RateLimitRPH,
		&p.Enabled, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.SceneType = model.SceneType(sceneType)
	p.ActionLow = model.RiskAction(actionLow)
	p.ActionMid = model.RiskAction(actionMid)
	p.ActionHigh = model.RiskAction(actionHigh)
	return &p, nil
}
