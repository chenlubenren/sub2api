package ent_test

import (
	"context"
	"database/sql"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/group"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestGroupUpdatePersistsNightCacheReadMultiplier(t *testing.T) {
	db, err := sql.Open("sqlite", "file:group_update_night_cache?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	first, err := client.Group.Create().SetName("night-cache-first").Save(ctx)
	require.NoError(t, err)
	second, err := client.Group.Create().SetName("night-cache-second").Save(ctx)
	require.NoError(t, err)

	_, err = client.Group.UpdateOneID(first.ID).
		SetNightCacheReadMultiplier(1.1).
		Save(ctx)
	require.NoError(t, err)
	first, err = client.Group.Get(ctx, first.ID)
	require.NoError(t, err)
	require.InDelta(t, 1.1, first.NightCacheReadMultiplier, 1e-12)

	updated, err := client.Group.Update().
		Where(group.IDIn(first.ID, second.ID)).
		SetNightCacheReadMultiplier(1.3).
		Save(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, updated)

	groups, err := client.Group.Query().
		Where(group.IDIn(first.ID, second.ID)).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, groups, 2)
	for _, current := range groups {
		require.InDelta(t, 1.3, current.NightCacheReadMultiplier, 1e-12)
	}
}
