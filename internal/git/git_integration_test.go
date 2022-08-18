package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/launchdarkly/ld-find-code-refs/internal/ld"
	"github.com/launchdarkly/ld-find-code-refs/internal/log"
	"github.com/launchdarkly/ld-find-code-refs/options"
	"github.com/stretchr/testify/require"

	"github.com/launchdarkly/ld-find-code-refs/search"
)

const (
	repoDir = "testdata/repo"
	flag1   = "flag1"
	flag2   = "flag2"
	flag3   = "flag3"
)

func TestMain(m *testing.M) {
	log.Init(true)
	os.Exit(m.Run())
}

func setupRepo(t *testing.T) *git.Repository {
	os.RemoveAll(repoDir)
	require.NoError(t, os.MkdirAll(repoDir, 0700))
	repo, err := git.PlainInit(repoDir, false)
	require.NoError(t, err)
	return repo
}

// TestFindExtinctions is an integration test against a real Git repository stored under the testdata directory.
func TestFindExtinctions(t *testing.T) {
	repo := setupRepo(t)

	// Create commit history
	flagFile, err := os.Create(filepath.Join(repoDir, "flag1.txt"))
	require.NoError(t, err)
	_, err = flagFile.WriteString(flag1)
	require.NoError(t, err)
	require.NoError(t, flagFile.Close())
	flagFile, err = os.Create(filepath.Join(repoDir, "flag2.txt"))
	require.NoError(t, err)
	_, err = flagFile.WriteString(flag2)
	require.NoError(t, err)
	require.NoError(t, flagFile.Close())
	flagFile, err = os.Create(filepath.Join(repoDir, "flag3.txt"))
	require.NoError(t, err)
	_, err = flagFile.WriteString(flag3)
	require.NoError(t, err)
	require.NoError(t, flagFile.Close())

	wt, err := repo.Worktree()
	require.NoError(t, err)

	who := object.Signature{Name: "LaunchDarkly", Email: "dev@launchdarkly.com", When: time.Unix(100000000, 0)}

	wt.Add("flag1.txt")
	wt.Add("flag2.txt")
	wt.Add("flag3.txt")
	_, err = wt.Commit("add flags", &git.CommitOptions{All: true, Committer: &who, Author: &who})
	require.NoError(t, err)

	// Test with a removed file
	err = os.Remove(filepath.Join(repoDir, "flag1.txt"))
	require.NoError(t, err)

	who.When = who.When.Add(time.Minute)
	message2 := "remove flag1"
	commit2, err := wt.Commit(message2, &git.CommitOptions{All: true, Committer: &who, Author: &who})
	require.NoError(t, err)

	// Test with an updated (truncated) file
	flagFile, err = os.Create(filepath.Join(repoDir, "flag2.txt"))
	require.NoError(t, err)
	require.NoError(t, flagFile.Close())

	who.When = who.When.Add(time.Minute)
	message3 := "remove flag2"
	commit3, err := wt.Commit("remove flag2", &git.CommitOptions{All: true, Committer: &who, Author: &who})
	require.NoError(t, err)

	err = os.Remove(filepath.Join(repoDir, "flag3.txt"))
	require.NoError(t, err)

	who.When = who.When.Add(time.Minute)
	message4 := "remove flag3"
	commit4, err := wt.Commit("remove flag3", &git.CommitOptions{All: true, Committer: &who, Author: &who})
	require.NoError(t, err)

	c := Client{workspace: repoDir}
	projKey := options.Project{
		Key: "default",
	}
	addProjKey := options.Project{
		Key: "otherProject",
	}
	projects := []options.Project{projKey, addProjKey}
	missingFlags := [][]string{{flag1, flag2}, {flag3}}
	matcher := search.Matcher{
		Elements: []search.ElementMatcher{
			search.NewElementMatcher(projKey.Key, ``, ``, []string{flag1, flag2}, nil),
			search.NewElementMatcher(addProjKey.Key, ``, ``, []string{flag3}, nil),
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
			Time:     who.When.Add(-time.Minute).Unix() * 1000,
			ProjKey:  projKey.Key,
			FlagKey:  flag2,
		},
		{
			Revision: commit2.String(),
			Message:  message2,
			Time:     who.When.Add(-time.Minute*2).Unix() * 1000,
			ProjKey:  projKey.Key,
			FlagKey:  flag1,
		},
		{
			Revision: commit4.String(),
			Message:  message4,
			Time:     who.When.Unix() * 1000,
			ProjKey:  addProjKey.Key,
			FlagKey:  flag3,
		},
	}
	require.Equal(t, expected, extinctions)

}
