package service

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// SubscriptionAnalyticsItem is the per-subscription row used by the admin
// analytics page. Costs are in USD and usage is averaged over the last seven
// calendar days (including zero-use days).
type SubscriptionAnalyticsItem struct {
	SubscriptionID           int64   `json:"subscription_id"`
	SubscriptionName         string  `json:"subscription_name"`
	UserID                   int64   `json:"user_id"`
	Email                    string  `json:"email"`
	AverageRechargeAmount    float64 `json:"average_recharge_amount"`
	AverageMonthlyActualCost float64 `json:"average_monthly_actual_cost"`
	AverageDailyUsage        float64 `json:"average_daily_usage"`
	DailyLimit               float64 `json:"daily_limit"`
	DailyLimitUtilization    float64 `json:"daily_limit_utilization"`
	sortOrder                int
}

type SubscriptionPackageAnalytics struct {
	SubscriptionName  string  `json:"subscription_name"`
	ActiveUsers       int     `json:"active_users"`
	AverageDailyUsage float64 `json:"average_daily_usage_7d"`
	SortOrder         int     `json:"sort_order,omitempty"`
}

type SubscriptionUtilizationTrendPoint struct {
	Date        string  `json:"date"`
	Utilization float64 `json:"utilization"`
}

type SubscriptionUtilizationTrend struct {
	SubscriptionID   int64                               `json:"subscription_id"`
	SubscriptionName string                              `json:"subscription_name"`
	PeriodStart      string                              `json:"period_start"`
	PeriodEnd        string                              `json:"period_end"`
	DailyLimit       float64                             `json:"daily_limit"`
	Trend            []SubscriptionUtilizationTrendPoint `json:"trend"`
}

type SubscriptionAnalyticsResponse struct {
	Items             []SubscriptionAnalyticsItem    `json:"items"`
	PackageSummaries  []SubscriptionPackageAnalytics `json:"package_summaries"`
	UtilizationTrends []SubscriptionUtilizationTrend `json:"utilization_trends"`
	GeneratedAt       time.Time                      `json:"generated_at"`
}

// GetSubscriptionAnalytics returns current active subscription analytics.
// It intentionally reads usage_logs by subscription_id so the page reflects
// the same billing records that enforce subscription quotas.
func (s *SubscriptionService) GetSubscriptionAnalytics(ctx context.Context) (*SubscriptionAnalyticsResponse, error) {
	if s == nil || s.entClient == nil {
		return &SubscriptionAnalyticsResponse{Items: []SubscriptionAnalyticsItem{}, PackageSummaries: []SubscriptionPackageAnalytics{}, UtilizationTrends: []SubscriptionUtilizationTrend{}, GeneratedAt: time.Now().UTC()}, nil
	}

	rows, err := s.entClient.QueryContext(ctx, `
SELECT us.id,
       g.name,
       us.user_id,
       u.email,
       COALESCE(plan.recharge_amount, 0)::double precision,
       COALESCE(SUM(ul.actual_cost) FILTER (WHERE ul.created_at >= NOW() - INTERVAL '30 days'), 0)::double precision,
       (COALESCE(SUM(ul.actual_cost) FILTER (WHERE ul.created_at >= NOW() - INTERVAL '7 days'), 0) / 7)::double precision,
       COALESCE(g.daily_limit_usd, 0)::double precision,
       COALESCE(plan.sort_order, g.sort_order, 0)
FROM user_subscriptions us
JOIN users u ON u.id = us.user_id AND u.deleted_at IS NULL
JOIN groups g ON g.id = us.group_id AND g.deleted_at IS NULL
LEFT JOIN LATERAL (
    SELECT AVG(price) AS recharge_amount, MIN(sort_order) AS sort_order
    FROM subscription_plans
    WHERE group_id = g.id AND for_sale = TRUE
) plan ON TRUE
LEFT JOIN usage_logs ul
       ON ul.subscription_id = us.id
      AND ul.actual_cost > 0
      AND ul.created_at >= NOW() - INTERVAL '30 days'
WHERE us.deleted_at IS NULL
  AND us.status = 'active'
  AND us.expires_at > NOW()
GROUP BY us.id, g.name, us.user_id, u.email, plan.recharge_amount, plan.sort_order, g.daily_limit_usd, g.sort_order
ORDER BY COALESCE(plan.sort_order, g.sort_order, 0), us.id`)
	if err != nil {
		return nil, fmt.Errorf("query subscription analytics: %w", err)
	}
	defer rows.Close()

	result := &SubscriptionAnalyticsResponse{
		Items:             make([]SubscriptionAnalyticsItem, 0),
		PackageSummaries:  make([]SubscriptionPackageAnalytics, 0),
		UtilizationTrends: make([]SubscriptionUtilizationTrend, 0),
		GeneratedAt:       time.Now().UTC(),
	}
	for rows.Next() {
		var item SubscriptionAnalyticsItem
		if err := rows.Scan(&item.SubscriptionID, &item.SubscriptionName, &item.UserID, &item.Email,
			&item.AverageRechargeAmount, &item.AverageMonthlyActualCost, &item.AverageDailyUsage,
			&item.DailyLimit, &item.sortOrder); err != nil {
			return nil, fmt.Errorf("scan subscription analytics: %w", err)
		}
		if item.DailyLimit > 0 {
			item.DailyLimitUtilization = item.AverageDailyUsage / item.DailyLimit * 100
		}
		result.Items = append(result.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscription analytics: %w", err)
	}

	type packageAccumulator struct {
		users map[int64]struct{}
		total float64
		sort  int
	}
	packages := make(map[string]*packageAccumulator)
	for _, item := range result.Items {
		p := packages[item.SubscriptionName]
		if p == nil {
			p = &packageAccumulator{users: make(map[int64]struct{}), sort: item.sortOrder}
			packages[item.SubscriptionName] = p
		}
		p.users[item.UserID] = struct{}{}
		p.total += item.AverageDailyUsage
		if item.sortOrder < p.sort {
			p.sort = item.sortOrder
		}
	}
	for name, p := range packages {
		avg := 0.0
		if len(p.users) > 0 {
			avg = p.total / float64(len(p.users))
		}
		result.PackageSummaries = append(result.PackageSummaries, SubscriptionPackageAnalytics{SubscriptionName: name, ActiveUsers: len(p.users), AverageDailyUsage: avg, SortOrder: p.sort})
	}
	sort.SliceStable(result.PackageSummaries, func(i, j int) bool {
		if result.PackageSummaries[i].SortOrder == result.PackageSummaries[j].SortOrder {
			return result.PackageSummaries[i].SubscriptionName < result.PackageSummaries[j].SubscriptionName
		}
		return result.PackageSummaries[i].SortOrder < result.PackageSummaries[j].SortOrder
	})

	trendRows, err := s.entClient.QueryContext(ctx, `
SELECT us.id,
       g.name,
       d.day::date,
       COALESCE(SUM(ul.actual_cost), 0)::double precision
FROM user_subscriptions us
JOIN groups g ON g.id = us.group_id AND g.deleted_at IS NULL
CROSS JOIN LATERAL generate_series(
    GREATEST(us.starts_at::date, (CURRENT_DATE - 29)::date),
    LEAST(us.expires_at::date, CURRENT_DATE),
    INTERVAL '1 day'
) d(day)
LEFT JOIN usage_logs ul
       ON ul.subscription_id = us.id
      AND ul.actual_cost > 0
      AND ul.created_at >= d.day
      AND ul.created_at < d.day + INTERVAL '1 day'
WHERE us.deleted_at IS NULL
  AND us.status = 'active'
  AND us.expires_at > NOW()
GROUP BY us.id, g.name, d.day
ORDER BY us.id, d.day`)
	if err != nil {
		return nil, fmt.Errorf("query subscription utilization trends: %w", err)
	}
	defer trendRows.Close()

	trendMap := make(map[int64]*SubscriptionUtilizationTrend)
	for trendRows.Next() {
		var id int64
		var name string
		var day time.Time
		var usage float64
		if err := trendRows.Scan(&id, &name, &day, &usage); err != nil {
			return nil, fmt.Errorf("scan subscription utilization trend: %w", err)
		}
		item := findSubscriptionAnalyticsItem(result.Items, id)
		if item == nil {
			continue
		}
		trend := trendMap[id]
		if trend == nil {
			trend = &SubscriptionUtilizationTrend{SubscriptionID: id, SubscriptionName: name, DailyLimit: item.DailyLimit, PeriodStart: day.Format("2006-01-02"), PeriodEnd: day.Format("2006-01-02"), Trend: make([]SubscriptionUtilizationTrendPoint, 0)}
			trendMap[id] = trend
		}
		date := day.Format("2006-01-02")
		trend.PeriodEnd = date
		utilization := 0.0
		if item.DailyLimit > 0 {
			utilization = usage / item.DailyLimit * 100
		}
		trend.Trend = append(trend.Trend, SubscriptionUtilizationTrendPoint{Date: date, Utilization: utilization})
	}
	if err := trendRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscription utilization trends: %w", err)
	}
	for _, item := range result.Items {
		if trend := trendMap[item.SubscriptionID]; trend != nil {
			result.UtilizationTrends = append(result.UtilizationTrends, *trend)
		}
	}
	return result, nil
}

func findSubscriptionAnalyticsItem(items []SubscriptionAnalyticsItem, id int64) *SubscriptionAnalyticsItem {
	for i := range items {
		if items[i].SubscriptionID == id {
			return &items[i]
		}
	}
	return nil
}
