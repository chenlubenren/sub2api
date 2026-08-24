package admin

import (
	"math"
	"testing"
	"time"
)

func TestExportDividendAndProfit(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		promo    string
		dividend float64
		profit   float64
	}{
		{name: "no promo code", amount: 100, dividend: 96, profit: 96},
		{name: "special promo code", amount: 100, promo: "CHENLUREC", dividend: 96, profit: 96},
		{name: "special promo code ignores case and whitespace", amount: 100, promo: " chenlurec ", dividend: 96, profit: 96},
		{name: "other promo code", amount: 100, promo: "GROUP-A", dividend: 81.6, profit: 18.4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dividend, profit := exportDividendAndProfit(tt.amount, tt.promo)
			if math.Abs(dividend-tt.dividend) > 1e-9 || math.Abs(profit-tt.profit) > 1e-9 {
				t.Fatalf("exportDividendAndProfit(%v, %q) = (%v, %v), want (%v, %v)", tt.amount, tt.promo, dividend, profit, tt.dividend, tt.profit)
			}
		})
	}
}

func TestParseExportDateRange(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	start, end, err := parseExportDateRange("2026-08-01", "2026-08-01", loc)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := start.Format(time.RFC3339), "2026-08-01T00:00:00+08:00"; got != want {
		t.Fatalf("start = %s, want %s", got, want)
	}
	if got, want := end.Format(time.RFC3339), "2026-08-02T00:00:00+08:00"; got != want {
		t.Fatalf("end = %s, want %s", got, want)
	}

	if _, _, err := parseExportDateRange("2026-08-02", "2026-08-01", loc); err == nil {
		t.Fatal("expected invalid date range to fail")
	}
	if _, _, err := parseExportDateRange("", "2026-08-01", loc); err == nil {
		t.Fatal("expected missing start date to fail")
	}
}

func TestPaymentTypeLabel(t *testing.T) {
	tests := map[string]string{
		"wxpay_native": "微信支付",
		"alipay":       "支付宝",
		"stripe_card":  "Stripe",
		"airwallex":    "Airwallex",
		"custom":       "custom",
	}
	for input, want := range tests {
		if got := paymentTypeLabel(input); got != want {
			t.Errorf("paymentTypeLabel(%q) = %q, want %q", input, got, want)
		}
	}
}
