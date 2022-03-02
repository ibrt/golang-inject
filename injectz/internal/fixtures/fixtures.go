//go:generate go run github.com/golang/mock/mockgen@v1.6.0 -source ./fixtures.go -destination ./mocks.go -package fixtures

package fixtures

import (
	"context"

	"github.com/golang/mock/gomock"

	"github.com/ibrt/golang-inject/injectz"
)

// ContextKey represents a context key.
type ContextKey int

// Known context keys.
const (
	FirstContextKey  ContextKey = iota
	SecondContextKey ContextKey = iota
	ThirdContextKey  ContextKey = iota
	FourthContextKey ContextKey = iota
)

// Initializer allows to mock an Initializer func.
type Initializer interface {
	Initialize(ctx context.Context) (injectz.Injector, injectz.Releaser)
}

// Injector allows to mock an Injector func.
type Injector interface {
	Inject(context.Context) context.Context
}

// Releaser allows to mock a Releaser func.
type Releaser interface {
	Release()
}

// Closer allows to mock an io.Closer.
type Closer interface {
	Close() error
}

// NewMatcher returns a new gomock.Matcher that uses the given callback to match.
func NewMatcher(f func(interface{}) bool, s string) gomock.Matcher {
	return &matcherFunc{
		f: f,
		s: s,
	}
}

type matcherFunc struct {
	f func(interface{}) bool
	s string
}

func (m *matcherFunc) Matches(x interface{}) bool {
	return m.f(x)
}

func (m *matcherFunc) String() string {
	return m.s
}
