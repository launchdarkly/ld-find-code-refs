package git

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bucketeer-io/code-refs/internal/ld"
	"github.com/bucketeer-io/code-refs/internal/log"
	"github.com/bucketeer-io/code-refs/options"
	"github.com/bucketeer-io/code-refs/search"
)

const (
	REPO_DIR = "testdata/repo"
)

var (
	flag1 = "flag1"
	flag2 = "flag2"
	flag3 = "flag3"
	flag4 = "flag4" // in bigdiff.txt
)

func TestMain(m *testing.M) {
	log.Init(true)
	os.Exit(m.Run())
}

func setupRepo(t *testing.T) *git.Repository {
	os.RemoveAll(REPO_DIR)
	require.NoError(t, os.MkdirAll(REPO_DIR, 0700))
	repo, err := git.PlainInit(REPO_DIR, false)
	require.NoError(t, err)
	return repo
}

// TestFindExtinctions is an integration test against a real Git repository stored under the testdata directory.
func TestFindExtinctions(t *testing.T) {
	repo := setupRepo(t)

	// Create commit history
	createRepoFile(t, "flag1.txt", &flag1)
	createRepoFile(t, "flag2.txt", &flag2)
	createRepoFile(t, "flag3.txt", &flag3)
	copyFile(t, "testdata/bigdiff.txt", repoPath("bigdiff.txt"))

	wt, err := repo.Worktree()
	require.NoError(t, err)

	who := object.Signature{Name: "LaunchDarkly", Email: "dev@launchdarkly.com", When: time.Unix(100000000, 0)}

	wt.Add("flag1.txt")
	wt.Add("flag2.txt")
	wt.Add("flag3.txt")
	wt.Add("bigdiff.txt")
	_, err = wt.Commit("add flags", &git.CommitOptions{All: true, Committer: &who, Author: &who})
	require.NoError(t, err)

	// Test with a removed file
	removeRepoFile(t, "flag1.txt")

	who, when2 := incrementCommitTime(who)
	message2 := "remove flag1"
	commit2, err := wt.Commit(message2, &git.CommitOptions{All: true, Committer: &who, Author: &who})
	require.NoError(t, err)

	// Test with an updated (truncated) file
	createRepoFile(t, "flag2.txt", nil)

	who, when3 := incrementCommitTime(who)
	message3 := "remove flag2"
	commit3, err := wt.Commit("remove flag2", &git.CommitOptions{All: true, Committer: &who, Author: &who})
	require.NoError(t, err)

	removeRepoFile(t, "flag3.txt")

	who, when4 := incrementCommitTime(who)
	message4 := "remove flag3"
	commit4, err := wt.Commit("remove flag3", &git.CommitOptions{All: true, Committer: &who, Author: &who})
	require.NoError(t, err)

	// Test big diff
	removeRepoFile(t, "bigdiff.txt")

	who, when5 := incrementCommitTime(who)
	message5 := "remove flag4 from bigdiff"
	commit5, err := wt.Commit(message5, &git.CommitOptions{All: true, Committer: &who, Author: &who})
	require.NoError(t, err)

	c := Client{workspace: REPO_DIR}
	projKey := options.Project{
		Key: "default",
	}
	addProjKey := options.Project{
		Key: "otherProject",
	}
	projects := []options.Project{projKey, addProjKey}
	missingFlags := [][]string{{flag1, flag2}, {flag3, flag4}}
	matcher := search.Matcher{
		Elements: []search.ElementMatcher{
			search.NewElementMatcher(projKey.Key, ``, ``, []string{flag1, flag2}, nil),
			search.NewElementMatcher(addProjKey.Key, ``, ``, []string{flag3, flag4}, nil),
		},
	}

	extinctions := make([]ld.ExtinctionRep, 0)
	for i, project := range projects {
		extinctionsByProject, err := c.FindExtinctions(project, missingFlags[i], matcher, 10)
		require.NoError(t, err)
		extinctions = append(extinctions, extinctionsByProject...)
	}

	expected := []ld.ExtinctionRep{
		{
			Revision: commit3.String(),
			Message:  message3,
			Time:     when3.UnixMilli(),
			ProjKey:  projKey.Key,
			FlagKey:  flag2,
		},
		{
			Revision: commit2.String(),
			Message:  message2,
			Time:     when2.UnixMilli(),
			ProjKey:  projKey.Key,
			FlagKey:  flag1,
		},
		{
			Revision: commit5.String(),
			Message:  message5,
			Time:     when5.UnixMilli(),
			ProjKey:  addProjKey.Key,
			FlagKey:  flag4,
		},
		{
			Revision: commit4.String(),
			Message:  message4,
			Time:     when4.UnixMilli(),
			ProjKey:  addProjKey.Key,
			FlagKey:  flag3,
		},
	}

	for i, e := range expected {
		assert.Equalf(t, e, extinctions[i], "exinction at element %d does not match", i)
	}
}

// Helper functions

func copyFile(t *testing.T, src, dst string) {
	sourceFileStat, err := os.Stat(src)
	require.NoError(t, err)
	require.Truef(t, sourceFileStat.Mode().IsRegular(), "%s is not a regular file", src)

	source, err := os.Open(src)
	require.NoError(t, err)
	defer source.Close()

	destination, err := os.Create(dst)
	require.NoError(t, err)

	defer destination.Close()
	_, err = io.Copy(destination, source)
	require.NoError(t, err)
}

func createRepoFile(t *testing.T, path string, content *string) {
	flagFile, err := os.Create(repoPath(path))
	require.NoError(t, err)
	if content != nil {
		_, err = flagFile.WriteString(*content)
		require.NoError(t, err)
	}
	require.NoError(t, flagFile.Close())
}

func removeRepoFile(t *testing.T, path string) {
	require.NoError(t, os.Remove(filepath.Join(REPO_DIR, path)))
}

func repoPath(path string) string {
	return filepath.Join(REPO_DIR, path)
}

func incrementCommitTime(who object.Signature) (object.Signature, time.Time) {
	t := who.When.Add(time.Minute)
	who.When = t

	return who, t
}
