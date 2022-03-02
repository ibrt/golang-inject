package injectz_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/ibrt/golang-inject/injectz"
	"github.com/ibrt/golang-inject/injectz/internal/fixtures"
)

func TestNewSingletonInjector(t *testing.T) {
	type contextKey int
	const myContextKey contextKey = iota

	ctx := injectz.NewSingletonInjector(myContextKey, "v1")(context.Background())
	require.Equal(t, "v1", ctx.Value(myContextKey))
}

func TestNewInjectors(t *testing.T) {
	ctrl := gomock.NewController(t)
	firstInjector := fixtures.NewMockInjector(ctrl)
	secondInjector := fixtures.NewMockInjector(ctrl)

	firstInjector.EXPECT().Inject(
		fixtures.NewMatcher(func(v interface{}) bool {
			ctx, ok := v.(context.Context)
			return ok && ctx.Value(fixtures.FirstContextKey) == nil && ctx.Value(fixtures.SecondContextKey) == nil
		}, "first injection")).
		DoAndReturn(func(ctx context.Context) context.Context {
			return context.WithValue(ctx, fixtures.FirstContextKey, "v1")
		})

	secondInjector.EXPECT().Inject(
		fixtures.NewMatcher(func(v interface{}) bool {
			ctx, ok := v.(context.Context)
			return ok && ctx.Value(fixtures.FirstContextKey) != nil && ctx.Value(fixtures.SecondContextKey) == nil
		}, "second injection")).
		DoAndReturn(func(ctx context.Context) context.Context {
			return context.WithValue(ctx, fixtures.SecondContextKey, "v2")
		})

	ctx := injectz.NewInjectors(firstInjector.Inject, secondInjector.Inject)(context.Background())
	require.Equal(t, "v1", ctx.Value(fixtures.FirstContextKey))
	require.Equal(t, "v2", ctx.Value(fixtures.SecondContextKey))
	ctrl.Finish()
}
