package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	noWorkspace = ""
	repoName    = "test-flags-repo"
)

var (
	sha             = ""
	tempSha         = ""
	_, caller, _, _ = runtime.Caller(0)
	testPath        = filepath.Dir(caller) + "/test"
	repoPath        = "/tmp/" + repoName
)

func TestMain(m *testing.M) {
	initCmd := exec.Command("./git_init.sh")
	initCmd.Dir = testPath
	out, err := initCmd.Output()
	exitVal := 1
	if err != nil {
		fmt.Printf("Error initializing test git repo: %+v\n", err)
	}

	output := strings.Split(string(out), "\n")
	if len(output) > 3 {
		// grab the last 2 text line of the git init script, which contains the current revision for the test repo's master and temp branches
		tempSha = strings.TrimSpace(output[len(output)-2])
		sha = strings.TrimSpace(output[len(output)-3])
		exitVal = m.Run()
	}

	deinitCmd := exec.Command("./git_deinit.sh")
	deinitCmd.Dir = testPath
	_, err = deinitCmd.Output()
	if err != nil {
		fmt.Printf("Failed to cleanup test git repo after running tests: %+v\n", err)
		exitVal = 1
	}
	os.Exit(exitVal)
}

func TestCommander_RevParse(t *testing.T) {
	tests := []struct {
		name string
		head string
		want string
	}{
		{
			name: "succeeds on branch head",
			head: "master",
			want: sha,
		},
		{
			name: "succeeds on sha head",
			head: sha,
			want: sha,
		},
		{
			name: "fails on invalid head",
			head: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commander{
				Workspace: repoPath,
				Head:      tt.head,
				RepoName:  repoName,
			}
			got, err := c.RevParse()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCommander_Clone(t *testing.T) {
	tests := []struct {
		name     string
		head     string
		endpoint string
		wantErr  bool
	}{
		{
			name:     "succeeds on valid repo",
			head:     "master",
			endpoint: "file://" + repoPath,
		},
		{
			name:     "fails on invalid endpoint",
			endpoint: "file:///not-a-repo/",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commander{
				Workspace: noWorkspace,
				RepoName:  repoName,
				Head:      tt.head,
			}
			dir, err := ioutil.TempDir("", c.RepoName)
			require.NoError(t, err)
			defer os.RemoveAll(dir)
			c.Workspace = dir
			err = c.Clone(tt.endpoint)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			localCmd := Commander{
				Workspace: repoPath,
				Head:      "master",
				RepoName:  repoName,
			}
			localRev, _ := localCmd.RevParse()
			rev, _ := c.RevParse()
			require.Equal(t, localRev, rev)
		})
	}
}

func TestCommander_Checkout(t *testing.T) {
	tests := []struct {
		name    string
		head    string
		want    string
		wantErr bool
	}{
		{
			name: "succeeds on idemptoment checkout",
			head: "master",
			want: sha,
		},
		{
			name: "succeeds on checkout",
			head: "temp",
			want: tempSha,
		},
		{
			name:    "fails on non-existant branch",
			head:    "not-a-head",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commander{
				Workspace: repoPath,
				Head:      tt.head,
				RepoName:  repoName,
			}
			err := c.Checkout()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			rev, _ := c.RevParse()
			assert.Equal(t, tt.want, rev)
		})
	}
}

func TestCommander_Grep(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		ctxLines int
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "succeeds",
			flags:    []string{"someFlag"},
			ctxLines: 0,
			wantLen:  5,
		},
		{
			name:     "succeeds with 1 context line",
			flags:    []string{"someFlag"},
			ctxLines: 1,
			wantLen:  10,
		},
		{
			name:     "succeeds with 2 context lines",
			flags:    []string{"someFlag"},
			ctxLines: 2,
			wantLen:  14,
		},
		{
			name:     "succeeds with multiple flags",
			flags:    []string{"someFlag", "anotherFlag"},
			ctxLines: 0,
			wantLen:  6,
		},
		{
			name:     "succeeds with escaped flag",
			flags:    []string{"escaped.flag"},
			ctxLines: 0,
			wantLen:  1,
		},
		{
			name:     "succeeds with ctxLines < 0",
			flags:    []string{"someFlag"},
			ctxLines: -1,
			wantLen:  5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commander{
				Workspace: repoPath,
				Head:      "master",
				RepoName:  repoName,
			}
			got, err := c.Grep(tt.flags, tt.ctxLines)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, tt.wantLen)
		})
	}
}
