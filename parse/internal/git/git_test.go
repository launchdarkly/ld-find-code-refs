package git

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	workspace = ""
	repoName  = "test-repo"
)

func TestCommander_RevParse(t *testing.T) {
	type fields struct {
		Workspace string
		Head      string
		RepoName  string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commander{
				Workspace: tt.fields.Workspace,
				Head:      tt.fields.Head,
				RepoName:  tt.fields.RepoName,
			}
			got, err := c.RevParse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Commander.RevParse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Commander.RevParse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommander_Clone(t *testing.T) {
	tests := []struct {
		name     string
		head     string
		endpoint string
	}{
		{
			"succeeds on valid https repo",
			"master",
			"https://github.com/zpqrtbnk/test-repo.git",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commander{
				Workspace: workspace,
				RepoName:  repoName,
				Head:      tt.head,
			}
			dir, err := ioutil.TempDir("", c.RepoName)
			require.NoError(t, err)
			defer os.RemoveAll(dir)
			c.Workspace = dir
			err = c.Clone(tt.endpoint)
			require.NoError(t, err)
		})
	}
}

func TestCommander_Checkout(t *testing.T) {
	type fields struct {
		Workspace string
		Head      string
		RepoName  string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commander{
				Workspace: tt.fields.Workspace,
				Head:      tt.fields.Head,
				RepoName:  tt.fields.RepoName,
			}
			if err := c.Checkout(); (err != nil) != tt.wantErr {
				t.Errorf("Commander.Checkout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommander_Grep(t *testing.T) {
	type fields struct {
		Workspace string
		Head      string
		RepoName  string
	}
	type args struct {
		flags    []string
		ctxLines int
		exclude  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Commander{
				Workspace: tt.fields.Workspace,
				Head:      tt.fields.Head,
				RepoName:  tt.fields.RepoName,
			}
			got, err := c.Grep(tt.args.flags, tt.args.ctxLines, tt.args.exclude)
			if (err != nil) != tt.wantErr {
				t.Errorf("Commander.Grep() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Commander.Grep() = %v, want %v", got, tt.want)
			}
		})
	}
}
