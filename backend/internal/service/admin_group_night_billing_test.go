package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type groupNightBillingRepoStub struct {
	GroupRepository
	group   *Group
	updated *Group
}

func (r *groupNightBillingRepoStub) GetByID(_ context.Context, _ int64) (*Group, error) {
	group := *r.group
	return &group, nil
}

func (r *groupNightBillingRepoStub) Update(_ context.Context, group *Group) error {
	updated := *group
	r.updated = &updated
	return nil
}

func TestAdminServiceUpdateGroupPersistsNightBillingMultipliers(t *testing.T) {
	existingGroup := &Group{
		ID:                       1,
		Name:                     "night-billing-group",
		Platform:                 PlatformOpenAI,
		Status:                   StatusActive,
		NightRateEnabled:         true,
		NightStart:               "01:30",
		NightEnd:                 "06:30",
		NightRateMultiplier:      1.5,
		CacheReadMultiplier:      1.1,
		NightCacheReadMultiplier: 1.2,
	}
	repo := &groupNightBillingRepoStub{group: existingGroup}
	svc := &adminServiceImpl{groupRepo: repo}

	nightRateEnabled := true
	nightStart := "01:30"
	nightEnd := "06:30"
	nightRateMultiplier := 1.0
	nightCacheReadMultiplier := 1.1
	group, err := svc.UpdateGroup(context.Background(), existingGroup.ID, &UpdateGroupInput{
		NightRateEnabled:         &nightRateEnabled,
		NightStart:               &nightStart,
		NightEnd:                 &nightEnd,
		NightRateMultiplier:      &nightRateMultiplier,
		NightCacheReadMultiplier: &nightCacheReadMultiplier,
	})

	require.NoError(t, err)
	require.NotNil(t, group)
	require.NotNil(t, repo.updated)
	require.True(t, repo.updated.NightRateEnabled)
	require.Equal(t, "01:30", repo.updated.NightStart)
	require.Equal(t, "06:30", repo.updated.NightEnd)
	require.InDelta(t, 1.0, repo.updated.NightRateMultiplier, 1e-12)
	require.InDelta(t, 1.1, repo.updated.NightCacheReadMultiplier, 1e-12)
}
