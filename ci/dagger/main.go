// Dagger module for ld-find-code-refs CI

package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
)

const (
	imageId           = "cimg/go:1.21"
	goReleaserImageId = "goreleaser/goreleaser:v1.20.0"
	sourceDir         = "/src"
)

type Ci struct{}

func (m *Ci) TestAndSnapshot(ctx context.Context, source *Directory) (string, error) {
	if _, err := m.Precommit(ctx, source); err != nil {
		return "", err
	}

	if _, err := m.TestRepo(ctx, source); err != nil {
		return "", err
	}

	if _, err := m.Snapshot(ctx, source); err != nil {
		return "", err
	}

	return "Lint, test and snapshot successful", nil
}

func (m *Ci) Precommit(ctx context.Context, source *Directory) (string, error) {
	return dag.Container().
		From(imageId).
		With(m.baseImage(ctx)).
		WithExec([]string{"sudo", "pip", "install", "pre-commit"}).
		WithDirectory(sourceDir, source, dagger.ContainerWithDirectoryOpts{
			Owner: "circleci",
		}).WithWorkdir(sourceDir).
		WithExec([]string{"pre-commit", "run",  "-a", "golangci-lint"}).
		Stdout(ctx)
}

// Setup and run go tests
func (m *Ci) TestRepo(ctx context.Context, source *Directory) (string, error) {
	return dag.Container().
		From(imageId).
		With(m.testImage(ctx)).
		WithDirectory(sourceDir, source, dagger.ContainerWithDirectoryOpts{
			Owner: "circleci",
		}).WithWorkdir(sourceDir).
		WithExec([]string{"sh", "-c", "go test -race -v ./... | richgo testfilter"}).
		WithEnvVariable("RICHGO_FORCE_COLOR", "1").
		Stdout(ctx)
}

func (m *Ci) Release(ctx context.Context, source *dagger.Directory) (string, error) {
	fmt.Printf("Building release")

	return dag.Goreleaser(source).
		WithGoCache().
		Release(ctx)
}

func (m *Ci) Snapshot(ctx context.Context, source *dagger.Directory) (string, error) {
	fmt.Printf("Building snapshot")

	// Goreleaser builds docker containers and needs access to the docker daemon
	docker := dag.Container().
		From("docker:dind").
		WithMountedCache("/var/lib/docker", dag.CacheVolume("dind")).
		WithExec([]string{
			"dockerd",
			"--tls=false",
			"--host=tcp://0.0.0.0:2375",
		}, ContainerWithExecOpts{
			// Errors in GHA without `InsecureRootCapabilities: true`
			InsecureRootCapabilities:      true,
			ExperimentalPrivilegedNesting: true,
		}).
		WithExposedPort(2375).
		AsService()

	goReleaserDocker := dag.Container().
		From(goReleaserImageId).
		WithDirectory(sourceDir, source).WithWorkdir(sourceDir).
		WithMountedCache("/go/pkg/mod", m.goModCacheVolume()).
		WithServiceBinding("docker", docker).
		WithEnvVariable("DOCKER_HOST", "tcp://docker:2375")

	return dag.Goreleaser(source, dagger.GoreleaserOpts{
		Ctr: goReleaserDocker,
	}).
		WithGoCache().
		Snapshot(ctx)
}

func (m *Ci) goModCacheVolume() *CacheVolume {
	return dag.CacheVolume("go-mod")
}

func (m *Ci) goBuildCacheVolume() *CacheVolume {
	return dag.CacheVolume("go-build")
}

// Trying to return a cached base image here.
func (m *Ci) baseImage(ctx context.Context) dagger.WithContainerFunc {
	return func(ctr *dagger.Container) *dagger.Container {
		return ctr.From(imageId).
			WithExec([]string{"sudo", "apt-get", "update"}).
			WithExec([]string{"sudo", "apt-get", "install", "python3-pip"}).
			WithMountedCache("/go/pkg/mod", m.goModCacheVolume()).
			WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
			WithMountedCache("/go/build-cache", m.goBuildCacheVolume()).
			WithEnvVariable("GOCACHE", "/go/build-cache").
			WithExec([]string{"sudo", "chown", "-R", "circleci", "/go/build-cache"}).
			WithExec([]string{"sudo", "chown", "-R", "circleci", "/go/pkg/mod"})
	}

}

func (m *Ci) testImage(ctx context.Context) dagger.WithContainerFunc {
	return func(ctr *dagger.Container) *dagger.Container {
		return ctr.From(imageId).
		With(m.baseImage(ctx)).
		WithExec([]string{"go", "install", "github.com/kyoh86/richgo@v0.3.10"})
	}
}
