//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadRegistrationPromoInfoUsesFirstPromoAndGroupFallback(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	firstUser, err := client.User.Create().
		SetEmail("first-promo@example.com").
		SetPasswordHash("hash").
		Save(ctx)
	require.NoError(t, err)
	secondUser, err := client.User.Create().
		SetEmail("fallback-promo@example.com").
		SetPasswordHash("hash").
		Save(ctx)
	require.NoError(t, err)

	firstPromo, err := client.PromoCode.Create().
		SetCode("EARLYCODE").
		SetNotes("  第一组  ").
		Save(ctx)
	require.NoError(t, err)
	laterPromo, err := client.PromoCode.Create().
		SetCode("LATERCODE").
		SetNotes("第二组").
		Save(ctx)
	require.NoError(t, err)
	fallbackPromo, err := client.PromoCode.Create().
		SetCode("NOGROUP").
		Save(ctx)
	require.NoError(t, err)

	now := time.Now()
	_, err = client.PromoCodeUsage.Create().
		SetPromoCodeID(laterPromo.ID).
		SetUserID(firstUser.ID).
		SetBonusAmount(0).
		SetUsedAt(now).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.PromoCodeUsage.Create().
		SetPromoCodeID(firstPromo.ID).
		SetUserID(firstUser.ID).
		SetBonusAmount(0).
		SetUsedAt(now.Add(-time.Minute)).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.PromoCodeUsage.Create().
		SetPromoCodeID(fallbackPromo.ID).
		SetUserID(secondUser.ID).
		SetBonusAmount(0).
		SetUsedAt(now).
		Save(ctx)
	require.NoError(t, err)

	infos, err := loadRegistrationPromoInfo(ctx, client, []int64{firstUser.ID, secondUser.ID, 999999})
	require.NoError(t, err)
	require.Equal(t, RegistrationPromoInfo{Code: "EARLYCODE", Group: "第一组"}, infos[firstUser.ID])
	require.Equal(t, RegistrationPromoInfo{Code: "NOGROUP", Group: "NOGROUP"}, infos[secondUser.ID])
	_, exists := infos[999999]
	require.False(t, exists)
}
