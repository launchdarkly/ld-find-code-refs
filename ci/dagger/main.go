// A generated module for Ci functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
)

type Ci struct{}

// Returns a container that echoes whatever string argument is provided
func (m *Ci) ContainerEcho(stringArg string) *Container {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg})
}

func (m *Ci) BuildAndTest(ctx context.Context, source *Directory) (string, error) {
	if _, err := m.TestRepo(ctx, source); err != nil {
		return "", err
	}

	out, err := m.BuildBinary(ctx, source)

	if err != nil {
		return out, err
	}

	return out, nil
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *Ci) TestRepo(ctx context.Context, source *Directory) (string, error) {
	return dag.Container().
		From("cimg/go:1.21").
		WithMountedCache("/go/pkg/mod", m.goModCacheVolume()).
		WithExec([]string{"sudo", "apt-get", "update"}).
		WithExec([]string{"go", "install", "github.com/jstemmer/go-junit-report@v1.0.0"}).
		WithExec([]string{"go", "install", "github.com/kyoh86/richgo@v0.3.10"}).
		WithExec([]string{"sudo", "apt-get", "update"}).
		WithExec([]string{"sudo", "apt-get", "install", "python3-pip"}).
		WithExec([]string{"sudo", "pip", "install", "pre-commit"}).
		WithDirectory("/src", source).WithWorkdir("/src").
		WithExec([]string{"go", "test", "./..."}).
		Stdout(ctx)
}

func (m *Ci) BuildBinary(ctx context.Context, source *dagger.Directory) (string, error) {
	fmt.Printf("Building binary")

	return dag.Goreleaser(source).
		WithGoCache().
		Release(ctx)
}

func (m *Ci) goModCacheVolume() *CacheVolume {
	return dag.CacheVolume("go-mod")
}

func (m *Ci) goBuildCacheVolume() *CacheVolume {
	return dag.CacheVolume("go-build")
}
