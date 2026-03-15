package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/engagelab/captcha/internal/model"
)

// PGChallengeStore implements ChallengeStore backed by PostgreSQL.
type PGChallengeStore struct {
	pool *pgxpool.Pool
}

// NewPGChallengeStore creates a new PostgreSQL-backed challenge store.
func NewPGChallengeStore(pool *pgxpool.Pool) *PGChallengeStore {
	return &PGChallengeStore{pool: pool}
}

// SaveChallenge inserts or upserts a challenge session.
func (s *PGChallengeStore) SaveChallenge(ch *model.ChallengeSession) error {
	ctx := context.Background()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO challenge_sessions
		 (id, app_id, scene_id, session_id, ip, ua_hash, fingerprint_id, challenge_type,
		  risk_score, risk_label, status, created_at, expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		 ON CONFLICT (id) DO UPDATE SET
		   status = EXCLUDED.status,
		   risk_score = EXCLUDED.risk_score,
		   risk_label = EXCLUDED.risk_label`,
		ch.ID, ch.AppID, ch.SceneID, ch.SessionID, ch.IP, ch.UAHash, ch.FingerprintID,
		string(ch.ChallengeType), ch.RiskScore, ch.RiskLabel, string(ch.Status),
		ch.CreatedAt, ch.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("save challenge: %w", err)
	}
	return nil
}

// GetChallenge retrieves a challenge session by ID.
func (s *PGChallengeStore) GetChallenge(id string) (*model.ChallengeSession, error) {
	ctx := context.Background()
	return s.scanChallenge(s.pool.QueryRow(ctx,
		`SELECT id, app_id, scene_id, session_id, ip, ua_hash, fingerprint_id, challenge_type,
		        risk_score, risk_label, status, created_at, expires_at
		 FROM challenge_sessions WHERE id = $1`, id,
	))
}

// UpdateChallengeStatus updates only the status field of a challenge session.
func (s *PGChallengeStore) UpdateChallengeStatus(id string, status model.ChallengeStatus) error {
	ctx := context.Background()
	tag, err := s.pool.Exec(ctx,
		`UPDATE challenge_sessions SET status = $1 WHERE id = $2`,
		string(status), id,
	)
	if err != nil {
		return fmt.Errorf("update challenge status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("challenge not found: %s", id)
	}
	return nil
}

// ListChallenges returns all challenge sessions, ordered by creation time descending.
func (s *PGChallengeStore) ListChallenges() []*model.ChallengeSession {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx,
		`SELECT id, app_id, scene_id, session_id, ip, ua_hash, fingerprint_id, challenge_type,
		        risk_score, risk_label, status, created_at, expires_at
		 FROM challenge_sessions ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []*model.ChallengeSession
	for rows.Next() {
		ch, err := s.scanChallengeRow(rows)
		if err != nil {
			continue
		}
		result = append(result, ch)
	}
	return result
}

func (s *PGChallengeStore) scanChallenge(row pgx.Row) (*model.ChallengeSession, error) {
	var ch model.ChallengeSession
	var challengeType, status string
	err := row.Scan(
		&ch.ID, &ch.AppID, &ch.SceneID, &ch.SessionID, &ch.IP, &ch.UAHash,
		&ch.FingerprintID, &challengeType, &ch.RiskScore, &ch.RiskLabel,
		&status, &ch.CreatedAt, &ch.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}
	ch.ChallengeType = model.ChallengeType(challengeType)
	ch.Status = model.ChallengeStatus(status)
	return &ch, nil
}

func (s *PGChallengeStore) scanChallengeRow(rows pgx.Rows) (*model.ChallengeSession, error) {
	var ch model.ChallengeSession
	var challengeType, status string
	err := rows.Scan(
		&ch.ID, &ch.AppID, &ch.SceneID, &ch.SessionID, &ch.IP, &ch.UAHash,
		&ch.FingerprintID, &challengeType, &ch.RiskScore, &ch.RiskLabel,
		&status, &ch.CreatedAt, &ch.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	ch.ChallengeType = model.ChallengeType(challengeType)
	ch.Status = model.ChallengeStatus(status)
	return &ch, nil
}
