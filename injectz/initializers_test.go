package injectz_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/ibrt/golang-inject/injectz"
	"github.com/ibrt/golang-inject/injectz/internal/fixtures"
)

func TestInitialize_Ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	firstInitializer := fixtures.NewMockInitializer(ctrl)
	secondInitializer := fixtures.NewMockInitializer(ctrl)
	firstInjector := fixtures.NewMockInjector(ctrl)
	secondInjector := fixtures.NewMockInjector(ctrl)
	firstReleaser := fixtures.NewMockReleaser(ctrl)
	secondReleaser := fixtures.NewMockReleaser(ctrl)
	isSecondReleased := false

	firstInitializer.EXPECT().Initialize(
		fixtures.NewMatcher(func(v interface{}) bool {
			ctx, ok := v.(context.Context)
			return ok && ctx.Value(fixtures.FirstContextKey) == nil && ctx.Value(fixtures.SecondContextKey) == nil
		}, "first initialization")).
		DoAndReturn(func(ctx context.Context) (injectz.Injector, injectz.Releaser) {
			return firstInjector.Inject, firstReleaser.Release
		})

	secondInitializer.EXPECT().Initialize(
		fixtures.NewMatcher(func(v interface{}) bool {
			ctx, ok := v.(context.Context)
			return ok && ctx.Value(fixtures.FirstContextKey) != nil && ctx.Value(fixtures.SecondContextKey) == nil
		}, "second injection")).
		DoAndReturn(func(ctx context.Context) (injectz.Injector, injectz.Releaser) {
			return secondInjector.Inject, secondReleaser.Release
		})

	firstInjector.EXPECT().Inject(
		fixtures.NewMatcher(func(v interface{}) bool {
			ctx, ok := v.(context.Context)
			return ok && ctx.Value(fixtures.FirstContextKey) == nil && ctx.Value(fixtures.SecondContextKey) == nil
		}, "first injection")).
		DoAndReturn(func(ctx context.Context) context.Context {
			return context.WithValue(ctx, fixtures.FirstContextKey, "v1")
		}).
		Times(2)

	secondInjector.EXPECT().Inject(
		fixtures.NewMatcher(func(v interface{}) bool {
			ctx, ok := v.(context.Context)
			return ok && ctx.Value(fixtures.FirstContextKey) != nil && ctx.Value(fixtures.SecondContextKey) == nil
		}, "second injection")).
		DoAndReturn(func(ctx context.Context) context.Context {
			return context.WithValue(ctx, fixtures.SecondContextKey, "v2")
		}).
		Times(2)

	firstReleaser.EXPECT().Release().Do(func() { require.True(t, isSecondReleased) })
	secondReleaser.EXPECT().Release().Do(func() { isSecondReleased = true })

	injector, releaser := injectz.Initialize(firstInitializer.Initialize, secondInitializer.Initialize)

	ctx := injector(context.Background())
	require.Equal(t, "v1", ctx.Value(fixtures.FirstContextKey))
	require.Equal(t, "v2", ctx.Value(fixtures.SecondContextKey))
	releaser()

	ctrl.Finish()
}

func TestInitialize_Fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	firstInitializer := fixtures.NewMockInitializer(ctrl)
	secondInitializer := fixtures.NewMockInitializer(ctrl)
	firstInjector := fixtures.NewMockInjector(ctrl)
	firstReleaser := fixtures.NewMockReleaser(ctrl)

	firstInitializer.EXPECT().Initialize(
		fixtures.NewMatcher(func(v interface{}) bool {
			ctx, ok := v.(context.Context)
			return ok && ctx.Value(fixtures.FirstContextKey) == nil && ctx.Value(fixtures.SecondContextKey) == nil
		}, "first initialization")).
		DoAndReturn(func(ctx context.Context) (injectz.Injector, injectz.Releaser) {
			return firstInjector.Inject, firstReleaser.Release
		})

	secondInitializer.EXPECT().Initialize(
		fixtures.NewMatcher(func(v interface{}) bool {
			ctx, ok := v.(context.Context)
			return ok && ctx.Value(fixtures.FirstContextKey) != nil && ctx.Value(fixtures.SecondContextKey) == nil
		}, "second injection")).
		DoAndReturn(func(ctx context.Context) (injectz.Injector, injectz.Releaser) {
			panic(fmt.Errorf("initializer error"))
		})

	firstInjector.EXPECT().Inject(
		fixtures.NewMatcher(func(v interface{}) bool {
			ctx, ok := v.(context.Context)
			return ok && ctx.Value(fixtures.FirstContextKey) == nil && ctx.Value(fixtures.SecondContextKey) == nil
		}, "first injection")).
		DoAndReturn(func(ctx context.Context) context.Context {
			return context.WithValue(ctx, fixtures.FirstContextKey, "v1")
		})

	firstReleaser.EXPECT().Release()

	require.PanicsWithError(t, "initializer error", func() {
		injectz.Initialize(firstInitializer.Initialize, secondInitializer.Initialize)
	})

	ctrl.Finish()
}
