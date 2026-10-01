//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type globalModelPricingSettingRepoStub struct {
	SettingRepository
	values map[string]string
	err    error
}

func (r *globalModelPricingSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	value, ok := r.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}

func (r *globalModelPricingSettingRepoStub) Set(_ context.Context, key, value string) error {
	if r.err != nil {
		return r.err
	}
	if r.values == nil {
		r.values = make(map[string]string)
	}
	r.values[key] = value
	return nil
}

func globalPricingCardsForTest(t *testing.T, svc *GlobalModelPricingService) []GlobalModelPricing {
	t.Helper()
	entries, err := svc.List(context.Background())
	require.NoError(t, err)
	return entries
}

func setTestGlobalPrice(t *testing.T, entries []GlobalModelPricing, model string, input float64) []GlobalModelPricing {
	t.Helper()
	for i := range entries {
		if entries[i].Model == model {
			entries[i].InputPrice = input
			return entries
		}
	}
	t.Fatalf("standard model %q was not found", model)
	return nil
}

func TestGlobalModelPricingUpdatePersistsAndRestoresBillingPrice(t *testing.T) {
	ctx := context.Background()
	repo := &globalModelPricingSettingRepoStub{values: make(map[string]string)}
	billing := NewBillingService(&config.Config{}, nil)
	svc := NewGlobalModelPricingService(repo, billing)

	entries := setTestGlobalPrice(t, globalPricingCardsForTest(t, svc), "gpt-5.5", 9e-6)
	_, err := svc.Update(ctx, entries)
	require.NoError(t, err)
	require.Contains(t, repo.values, SettingKeyGlobalModelPricing)

	pricing, err := billing.GetModelPricing("gpt-5.5")
	require.NoError(t, err)
	require.InDelta(t, 9e-6, pricing.InputPricePerToken, 1e-12)

	restartedBilling := NewBillingService(&config.Config{}, nil)
	restartedService := NewGlobalModelPricingService(repo, restartedBilling)
	require.NoError(t, restartedService.Load(ctx))
	restored, err := restartedBilling.GetModelPricing("gpt-5.5")
	require.NoError(t, err)
	require.InDelta(t, 9e-6, restored.InputPricePerToken, 1e-12)
}

func TestGlobalModelPricingOverridesGroupModelCard(t *testing.T) {
	ctx := context.Background()
	repo := &globalModelPricingSettingRepoStub{values: make(map[string]string)}
	billing := NewBillingService(&config.Config{}, nil)
	svc := NewGlobalModelPricingService(repo, billing)
	entries := setTestGlobalPrice(t, globalPricingCardsForTest(t, svc), "gpt-5.5", 9e-6)
	_, err := svc.Update(ctx, entries)
	require.NoError(t, err)

	groupInputPrice := 1e-6
	resolver := NewModelPricingResolver(nil, billing)
	resolved := resolver.Resolve(ctx, PricingInput{
		Model: "gpt-5.5",
		Group: &Group{ModelPricing: []ChannelModelPricing{{
			Models:      []string{"gpt-5.5"},
			BillingMode: BillingModeToken,
			InputPrice:  &groupInputPrice,
		}}},
	})

	require.NotNil(t, resolved)
	require.NotNil(t, resolved.BasePricing)
	require.InDelta(t, 9e-6, resolved.BasePricing.InputPricePerToken, 1e-12)
}

func TestGlobalModelPricingLoadPropagatesRepositoryFailure(t *testing.T) {
	repo := &globalModelPricingSettingRepoStub{err: assertError{}}
	svc := NewGlobalModelPricingService(repo, NewBillingService(&config.Config{}, nil))
	require.Error(t, svc.Load(context.Background()))
}

type assertError struct{}

func (assertError) Error() string { return "settings unavailable" }
