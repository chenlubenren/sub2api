package service

import (
	"context"
	"database/sql"
	"fmt"
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

	// Annotate invitees with the group of the first ancestor who registered
	// through a grouped promo code. This keeps the existing promo-code data
	// unchanged while exposing labels such as "曲瑶组-二次邀请" in admin views.
	for _, userID := range userIDs {
		if group, depth, ok := resolveInheritedPromoGroup(ctx, client, userID, result); ok && depth > 0 {
			info := result[userID]
			// A user's own registration promo remains authoritative. Inheritance
			// is only used when the user has no direct promo information.
			if info.Group == "" {
				info.Group = fmt.Sprintf("%s-%s", group, inviteLevelLabel(depth+1))
				result[userID] = info
			}
		}
	}

	return result, nil
}

// resolveInheritedPromoGroup walks the inviter chain. depth is the number of
// edges from userID to the ancestor whose direct registration promo provides
// the group label. A missing affiliate table (for older installations) is
// treated as no inheritance so existing pages continue to work.
func resolveInheritedPromoGroup(ctx context.Context, client *dbent.Client, userID int64, direct map[int64]RegistrationPromoInfo) (string, int, bool) {
	if client == nil || userID <= 0 {
		return "", 0, false
	}
	seen := map[int64]bool{userID: true}
	current := userID
	depth := 0
	for depth < 100 {
		var inviter sql.NullInt64
		rows, err := client.QueryContext(ctx, `SELECT inviter_id FROM user_affiliates WHERE user_id = $1`, current)
		if err != nil {
			return "", 0, false
		}
		if rows.Next() {
			_ = rows.Scan(&inviter)
		}
		_ = rows.Close()
		if !inviter.Valid || inviter.Int64 <= 0 || seen[inviter.Int64] {
			return "", 0, false
		}
		seen[inviter.Int64] = true
		depth++
		parentID := inviter.Int64
		if info, ok := direct[parentID]; ok && info.Group != "" {
			return info.Group, depth, true
		}
		// Ancestors are normally absent from the current page; load only their
		// earliest direct promo usage when needed.
		parentInfo := loadDirectPromoInfo(ctx, client, parentID)
		if parentInfo.Group != "" {
			return parentInfo.Group, depth, true
		}
		current = parentID
	}
	return "", 0, false
}

func loadDirectPromoInfo(ctx context.Context, client *dbent.Client, userID int64) RegistrationPromoInfo {
	if client == nil || userID <= 0 {
		return RegistrationPromoInfo{}
	}
	usage, err := client.PromoCodeUsage.Query().
		Where(promocodeusage.UserIDEQ(userID)).
		Order(dbent.Asc(promocodeusage.FieldUsedAt)).
		WithPromoCode().
		First(ctx)
	if err != nil || usage == nil || usage.Edges.PromoCode == nil {
		return RegistrationPromoInfo{}
	}
	code := strings.ToUpper(strings.TrimSpace(usage.Edges.PromoCode.Code))
	group := ""
	if usage.Edges.PromoCode.Notes != nil {
		group = strings.TrimSpace(*usage.Edges.PromoCode.Notes)
	}
	if group == "" {
		group = code
	}
	return RegistrationPromoInfo{Code: code, Group: group}
}

func inviteLevelLabel(level int) string {
	labels := []string{"", "一次邀请", "二次邀请", "三次邀请", "四次邀请", "五次邀请", "六次邀请", "七次邀请", "八次邀请", "九次邀请", "十次邀请"}
	if level > 0 && level < len(labels) {
		return labels[level]
	}
	return fmt.Sprintf("%d次邀请", level)
}
