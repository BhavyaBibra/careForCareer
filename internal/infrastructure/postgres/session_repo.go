package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	authdomain "careergps/internal/domain/auth"
	"careergps/pkg/apperrors"
)

type SessionRepo struct {
	pool *pgxpool.Pool
}

func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{pool: pool}
}

func (r *SessionRepo) Create(ctx context.Context, s *authdomain.Session) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, refresh_token, expires_at, revoked, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$6)`,
		s.ID, s.UserID, s.RefreshToken, s.ExpiresAt, s.Revoked, s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("session: create: %w", err)
	}
	return nil
}

func (r *SessionRepo) GetByRefreshToken(ctx context.Context, token string) (*authdomain.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, refresh_token, expires_at, revoked, created_at
		FROM sessions WHERE refresh_token=$1`, token)
	var s authdomain.Session
	err := row.Scan(&s.ID, &s.UserID, &s.RefreshToken, &s.ExpiresAt, &s.Revoked, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("session: get by token: %w", err)
	}
	return &s, nil
}

func (r *SessionRepo) Revoke(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE sessions SET revoked=true, updated_at=$1 WHERE refresh_token=$2`,
		time.Now().UTC(), token,
	)
	return err
}

func (r *SessionRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*authdomain.Session, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, refresh_token, expires_at, revoked, created_at
		FROM sessions WHERE user_id=$1 AND revoked=false AND expires_at > NOW()`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []*authdomain.Session
	for rows.Next() {
		var s authdomain.Session
		if err := rows.Scan(&s.ID, &s.UserID, &s.RefreshToken, &s.ExpiresAt, &s.Revoked, &s.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, &s)
	}
	return sessions, rows.Err()
}
