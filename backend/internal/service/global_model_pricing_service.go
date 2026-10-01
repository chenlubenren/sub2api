package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const SettingKeyGlobalModelPricing = "global_model_pricing"

// GlobalModelPricing is the administrator-managed standard price card. Prices
// are USD per token at the API boundary, matching the existing billing and
// channel pricing APIs.
type GlobalModelPricing struct {
	Model           string  `json:"model"`
	Label           string  `json:"label"`
	Platform        string  `json:"platform"`
	InputPrice      float64 `json:"input_price"`
	OutputPrice     float64 `json:"output_price"`
	CacheWritePrice float64 `json:"cache_write_price"`
	CacheReadPrice  float64 `json:"cache_read_price"`
}

// GlobalModelPricingService owns the globally applicable, administrator-edited
// standard rates. Overrides are persisted in the settings table and loaded into
// BillingService, so the values survive a restart and are used by requests
// immediately after an administrator saves them.
type GlobalModelPricingService struct {
	settingRepo    SettingRepository
	billingService *BillingService
}

func NewGlobalModelPricingService(settingRepo SettingRepository, billingService *BillingService) *GlobalModelPricingService {
	return &GlobalModelPricingService{settingRepo: settingRepo, billingService: billingService}
}

// ProvideGlobalModelPricingService restores persisted global prices before any
// HTTP handler is exposed. An unavailable settings store must abort startup:
// otherwise a restart could silently resume billing with stale base prices.
func ProvideGlobalModelPricingService(settingRepo SettingRepository, billingService *BillingService) (*GlobalModelPricingService, error) {
	svc := NewGlobalModelPricingService(settingRepo, billingService)
	if err := svc.Load(context.Background()); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *GlobalModelPricingService) Load(ctx context.Context) error {
	if s == nil || s.settingRepo == nil || s.billingService == nil {
		return fmt.Errorf("global model pricing service is unavailable")
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyGlobalModelPricing)
	if errors.Is(err, ErrSettingNotFound) {
		s.billingService.SetGlobalModelPricingOverrides(nil)
		return nil
	}
	if err != nil {
		return fmt.Errorf("load global model pricing: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		s.billingService.SetGlobalModelPricingOverrides(nil)
		return nil
	}
	var entries []GlobalModelPricing
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return fmt.Errorf("parse global model pricing: %w", err)
	}
	overrides, err := globalPricingOverrides(entries, false)
	if err != nil {
		return err
	}
	s.billingService.SetGlobalModelPricingOverrides(overrides)
	return nil
}

func (s *GlobalModelPricingService) List(ctx context.Context) ([]GlobalModelPricing, error) {
	if s == nil || s.billingService == nil {
		return nil, fmt.Errorf("global model pricing service is unavailable")
	}
	var out []GlobalModelPricing
	for _, group := range StandardModelPricingCatalog() {
		for _, entry := range group.Models {
			pricing, err := s.billingService.GetModelPricing(entry.Model)
			if err != nil || pricing == nil {
				return nil, fmt.Errorf("resolve standard pricing for %s: %w", entry.Model, err)
			}
			cacheWrite := pricing.CacheCreationPricePerToken
			if pricing.SupportsCacheBreakdown && pricing.CacheCreation5mPrice > 0 {
				cacheWrite = pricing.CacheCreation5mPrice
			}
			out = append(out, GlobalModelPricing{
				Model:           entry.Model,
				Label:           entry.Label,
				Platform:        entry.Platform,
				InputPrice:      pricing.InputPricePerToken,
				OutputPrice:     pricing.OutputPricePerToken,
				CacheWritePrice: cacheWrite,
				CacheReadPrice:  pricing.CacheReadPricePerToken,
			})
		}
	}
	return out, nil
}

// Update replaces the standard catalogue prices atomically. Every canonical
// card must be supplied so an accidental partial save cannot reset a model to
// an unknown state.
func (s *GlobalModelPricingService) Update(ctx context.Context, entries []GlobalModelPricing) ([]GlobalModelPricing, error) {
	if s == nil || s.settingRepo == nil || s.billingService == nil {
		return nil, fmt.Errorf("global model pricing service is unavailable")
	}
	overrides, err := globalPricingOverrides(entries, true)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(entries)
	if err != nil {
		return nil, fmt.Errorf("marshal global model pricing: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyGlobalModelPricing, string(payload)); err != nil {
		return nil, fmt.Errorf("save global model pricing: %w", err)
	}
	s.billingService.SetGlobalModelPricingOverrides(overrides)
	return s.List(ctx)
}

func globalPricingOverrides(entries []GlobalModelPricing, requireComplete bool) ([]GlobalModelPricingOverride, error) {
	catalog := make(map[string]StandardModelPricingCatalogEntry)
	for _, group := range StandardModelPricingCatalog() {
		for _, entry := range group.Models {
			catalog[entry.Model] = entry
		}
	}
	seen := make(map[string]struct{}, len(entries))
	overrides := make([]GlobalModelPricingOverride, 0, len(entries))
	for _, entry := range entries {
		model := strings.ToLower(strings.TrimSpace(entry.Model))
		catalogEntry, ok := catalog[model]
		if !ok {
			return nil, infraerrors.BadRequest("UNKNOWN_GLOBAL_PRICING_MODEL", "model is not in the standard pricing catalogue")
		}
		if _, duplicate := seen[model]; duplicate {
			return nil, infraerrors.BadRequest("DUPLICATE_GLOBAL_PRICING_MODEL", "model pricing can only be supplied once")
		}
		if entry.InputPrice < 0 || entry.OutputPrice < 0 || entry.CacheWritePrice < 0 || entry.CacheReadPrice < 0 {
			return nil, infraerrors.BadRequest("INVALID_GLOBAL_PRICING", "model pricing values must be greater than or equal to zero")
		}
		seen[model] = struct{}{}
		overrides = append(overrides, GlobalModelPricingOverride{
			Model:           catalogEntry.Model,
			InputPrice:      entry.InputPrice,
			OutputPrice:     entry.OutputPrice,
			CacheWritePrice: entry.CacheWritePrice,
			CacheReadPrice:  entry.CacheReadPrice,
		})
	}
	if requireComplete && len(seen) != len(catalog) {
		return nil, infraerrors.BadRequest("INCOMPLETE_GLOBAL_PRICING", "all standard model prices must be supplied")
	}
	return overrides, nil
}
