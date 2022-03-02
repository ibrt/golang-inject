package injectz

import (
	"context"
)

// Initializer initializes a value, returning a corresponding Injector and Releaser.
type Initializer func(ctx context.Context) (Injector, Releaser)

// Initialize calls all initializers and returns a compound Injector and Releaser.
// If an initializer panics, previous releasers are invoked before propagating it.
func Initialize(initializers ...Initializer) (Injector, Releaser) {
	ctx := context.Background()

	injectors := make([]Injector, 0, len(initializers))
	releasers := make([]Releaser, 0, len(initializers))

	defer func() {
		if r := recover(); r != nil {
			NewReleasers(releasers...)()
			panic(r)
		}
	}()

	for _, initializer := range initializers {
		injector, releaser := initializer(ctx)

		injectors = append(injectors, injector)
		releasers = append(releasers, releaser)

		ctx = injector(ctx)
	}

	return NewInjectors(injectors...), NewReleasers(releasers...)
}
