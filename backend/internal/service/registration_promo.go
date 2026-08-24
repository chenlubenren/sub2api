package service

import (
	"context"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/promocodeusage"
)

// RegistrationPromoInfo is the first registration promo code used by a user.
// Promo-code notes are the preferred human-readable group label; the code is
// retained for commission rules and as a fallback when notes are empty.
type RegistrationPromoInfo struct {
	Code  string
	Group string
}

// loadRegistrationPromoInfo loads the first promo-code usage for each user in
// one query. Registration codes are expected to be applied once, but selecting
// the earliest usage keeps the result stable if historical data contains more
// than one record for a user.
func loadRegistrationPromoInfo(ctx context.Context, client *dbent.Client, userIDs []int64) (map[int64]RegistrationPromoInfo, error) {
	result := make(map[int64]RegistrationPromoInfo, len(userIDs))
	if client == nil || len(userIDs) == 0 {
		return result, nil
	}

	usages, err := client.PromoCodeUsage.Query().
		Where(promocodeusage.UserIDIn(userIDs...)).
		Order(dbent.Asc(promocodeusage.FieldUsedAt)).
		WithPromoCode().
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, usage := range usages {
		if _, exists := result[usage.UserID]; exists || usage.Edges.PromoCode == nil {
			continue
		}
		code := strings.ToUpper(strings.TrimSpace(usage.Edges.PromoCode.Code))
		group := ""
		if usage.Edges.PromoCode.Notes != nil {
			group = strings.TrimSpace(*usage.Edges.PromoCode.Notes)
		}
		if group == "" {
			group = code
		}
		result[usage.UserID] = RegistrationPromoInfo{Code: code, Group: group}
	}

	return result, nil
}
