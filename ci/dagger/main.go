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

	return "Lint, test and build successful", nil
}

func (m *Ci) Precommit(ctx context.Context, source *Directory) (string, error) {
	return m.baseImage().
		WithExec([]string{"sudo", "pip", "install", "pre-commit"}).
		WithDirectory(sourceDir, source, dagger.ContainerWithDirectoryOpts{
			Owner: "circleci",
		}).WithWorkdir(sourceDir).
		WithExec([]string{"pre-commit", "install"}).
		WithExec([]string{"pre-commit", "run", "-a", "golangci-lint"}).
		Stdout(ctx)
}

// Setup and run go tests
func (m *Ci) TestRepo(ctx context.Context, source *Directory) (string, error) {
	return m.baseImage().
		From(imageId).
		WithExec([]string{"go", "install", "github.com/jstemmer/go-junit-report@v1.0.0"}).
		WithExec([]string{"go", "install", "github.com/kyoh86/richgo@v0.3.10"}).
		WithDirectory(sourceDir, source, dagger.ContainerWithDirectoryOpts{
			Owner: "circleci",
		}).WithWorkdir(sourceDir).
		WithExec([]string{"go", "test", "./..."}).
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
			ExperimentalPrivilegedNesting: true,
		}).
		WithExposedPort(2375).
		AsService()

	goReleaserDocker := dag.Container().
		From(goReleaserImageId).
		WithDirectory(sourceDir, source).WithWorkdir(sourceDir).
		WithServiceBinding("docker", docker).
		WithEnvVariable("DOCKER_HOST", "tcp://docker:2375")

	_, err := dag.Goreleaser(source, dagger.GoreleaserOpts{
		Ctr: goReleaserDocker,
	}).
		WithGoCache().
		Snapshot(ctx)

	if err != nil {
		return "", err
	}

	return "Snapshot successful", nil
}

func (m *Ci) goModCacheVolume() *CacheVolume {
	return dag.CacheVolume("go-mod")
}

func (m *Ci) goBuildCacheVolume() *CacheVolume {
	return dag.CacheVolume("go-build")
}

func (m *Ci) baseImage() *Container {
	return dag.Container().From(imageId).
		WithMountedCache("/go/pkg/mod", m.goModCacheVolume()).
		WithExec([]string{"sudo", "apt-get", "update"}).
		WithExec([]string{"sudo", "apt-get", "install", "python3-pip"})
}
