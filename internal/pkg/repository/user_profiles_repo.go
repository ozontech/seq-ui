package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/ozontech/seq-ui/internal/app/types"
	sqlb "github.com/ozontech/seq-ui/internal/pkg/repository/sql_builder"
)

type userProfilesRepository struct {
	*pool
}

func newUserProfilesRepository(pool *pool) *userProfilesRepository {
	return &userProfilesRepository{pool}
}

func (r *userProfilesRepository) GetOrCreate(ctx context.Context, req types.GetOrCreateUserProfileRequest) (types.UserProfile, error) {
	userProfile := types.UserProfile{
		UserName: req.UserName,
	}

	query, args := "SELECT id, timezone, onboarding_version, log_columns FROM user_profiles WHERE user_name = $1 LIMIT 1",
		[]any{req.UserName}

	metricLabelsSelect := []string{"user_profiles", "SELECT"}
	var logColumns string
	err := r.queryRow(ctx, metricLabelsSelect, query, args...).Scan(
		&userProfile.ID,
		&userProfile.Timezone,
		&userProfile.OnboardingVersion,
		&logColumns,
	)

	// create user profile if it doesn't exist
	if errors.Is(err, pgx.ErrNoRows) {
		query, args = "INSERT INTO user_profiles (user_name,timezone,onboarding_version,log_columns) VALUES ($1,$2,$3,$4) RETURNING id",
			[]any{req.UserName, "", "", "[]"}

		metricLabelsInsert := []string{"user_profiles", "INSERT"}
		if err = r.queryRow(ctx, metricLabelsInsert, query, args...).Scan(&userProfile.ID); err != nil {
			incErrorMetric(err, metricLabelsInsert)
			return userProfile, fmt.Errorf("failed to create user profile: %w", err)
		}
		return userProfile, nil
	}
	if err != nil {
		incErrorMetric(err, metricLabelsSelect)
		return userProfile, fmt.Errorf("failed to get user profile: %w", err)
	}

	err = json.Unmarshal([]byte(logColumns), &userProfile.LogColumns.LogColumns)
	if err != nil {
		return userProfile, fmt.Errorf("failed to parse log columns: %w", err)
	}

	return userProfile, nil
}

func (r *userProfilesRepository) Update(ctx context.Context, req types.UpdateUserProfileRequest) error {
	qb := sqlb.Update("user_profiles").
		Where(sq.Eq{
			"user_name": req.UserName,
		})
	if req.Timezone != nil {
		qb = qb.Set("timezone", *req.Timezone)
	}
	if req.OnboardingVersion != nil {
		qb = qb.Set("onboarding_version", *req.OnboardingVersion)
	}
	if req.LogColumns != nil {
		logColumns, _ := json.Marshal(req.LogColumns.LogColumns)
		qb = qb.Set("log_columns", logColumns)
	}

	query, args := qb.MustSql()

	metricLabels := []string{"user_profiles", "UPDATE"}
	tag, err := r.exec(ctx, metricLabels, query, args...)
	if err != nil {
		err = fmt.Errorf("failed to update user profile: %w", err)
	} else if tag.RowsAffected() == 0 {
		err = types.NewErrNotFound("user")
	}

	if err != nil {
		incErrorMetric(err, metricLabels)
		return err
	}

	return nil
}
