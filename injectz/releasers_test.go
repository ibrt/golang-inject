package injectz_test

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/ibrt/golang-inject/injectz"
	"github.com/ibrt/golang-inject/injectz/internal/fixtures"
)

func TestNewNoopReleaser(t *testing.T) {
	require.NotPanics(t, func() {
		injectz.NewNoopReleaser()()
	})
}

func TestNewCloseReleaser(t *testing.T) {
	ctrl := gomock.NewController(t)
	closer := fixtures.NewMockCloser(ctrl)
	closer.EXPECT().Close().Return(fmt.Errorf("close error"))
	injectz.NewCloseReleaser(closer)()
	ctrl.Finish()
}

func TestNewReleasers(t *testing.T) {
	ctrl := gomock.NewController(t)
	firstReleaser := fixtures.NewMockReleaser(ctrl)
	secondReleaser := fixtures.NewMockReleaser(ctrl)
	isSecondReleased := false
	firstReleaser.EXPECT().Release().Do(func() { require.True(t, isSecondReleased) })
	secondReleaser.EXPECT().Release().Do(func() { isSecondReleased = true })
	injectz.NewReleasers(firstReleaser.Release, secondReleaser.Release)()
	ctrl.Finish()
}
