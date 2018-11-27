package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	noWorkspace = ""
	repoName    = "test-flags-repo"
	sha         = "ce110109a6eee2a653dc7a311d6d4b9c93555010"
)

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
				Workspace: filepath.Dir(caller) + "/" + repoName,
				Head:      tt.head,
				RepoName:  repoName,
			}
			got, err := c.RevParse()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

var _, caller, _, _ = runtime.Caller(0)

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
			endpoint: "file://" + filepath.Dir(caller) + "/" + repoName,
		},
		{
			name:     "fails on invalid endpoint",
			endpoint: "file:///not-a-repo/",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.endpoint)
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
				Workspace: repoName,
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
			want: "892f88313d2bb23108e0099d3f5a231bce2740b1",
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
				Workspace: filepath.Dir(caller) + "/" + repoName,
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
			wantLen:  4,
		},
		{
			name:     "succeeds with 1 context line",
			flags:    []string{"someFlag"},
			ctxLines: 1,
			wantLen:  7,
		},
		{
			name:     "succeeds with 2 context lines",
			flags:    []string{"someFlag"},
			ctxLines: 2,
			wantLen:  10,
		},
		{
			name:     "succeeds with multiple flags",
			flags:    []string{"someFlag", "anotherFlag"},
			ctxLines: 0,
			wantLen:  5,
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
			wantLen:  4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commander{
				Workspace: filepath.Dir(caller) + "/" + repoName,
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
