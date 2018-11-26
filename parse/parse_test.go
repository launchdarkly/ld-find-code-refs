package parse

import (
	"reflect"
	"testing"

	"github.com/launchdarkly/git-flag-parser/parse/internal/git"
	"github.com/launchdarkly/git-flag-parser/parse/internal/ld"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Parse()
		})
	}
}

func Test_getFlags(t *testing.T) {
	type args struct {
		ldApi ld.ApiClient
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFlags(tt.args.ldApi)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_branch_MakeBranchRep(t *testing.T) {
	type fields struct {
		Name       string
		Head       string
		IsDefault  bool
		PushTime   int64
		SyncTime   int64
		References references
	}
	tests := []struct {
		name   string
		fields fields
		want   ld.BranchRep
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &branch{
				Name:       tt.fields.Name,
				Head:       tt.fields.Head,
				IsDefault:  tt.fields.IsDefault,
				PushTime:   tt.fields.PushTime,
				SyncTime:   tt.fields.SyncTime,
				References: tt.fields.References,
			}
			if got := b.MakeBranchRep(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("branch.MakeBranchRep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_references_Len(t *testing.T) {
	tests := []struct {
		name string
		r    references
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Len(); got != tt.want {
				t.Errorf("references.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_references_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		r    references
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("references.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_references_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		r    references
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Swap(tt.args.i, tt.args.j)
		})
	}
}

func Test_references_MakeReferenceReps(t *testing.T) {
	tests := []struct {
		name string
		r    references
		want []ld.ReferenceRep
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.MakeReferenceReps(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("references.MakeReferenceReps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_references_MakeHunkReps(t *testing.T) {
	tests := []struct {
		name string
		r    references
		want []ld.HunkRep
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.MakeHunkReps(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("references.MakeHunkReps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_branch_findReferences(t *testing.T) {
	type fields struct {
		Name       string
		Head       string
		IsDefault  bool
		PushTime   int64
		SyncTime   int64
		References references
	}
	type args struct {
		cmd   git.Commander
		flags []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *branch
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &branch{
				Name:       tt.fields.Name,
				Head:       tt.fields.Head,
				IsDefault:  tt.fields.IsDefault,
				PushTime:   tt.fields.PushTime,
				SyncTime:   tt.fields.SyncTime,
				References: tt.fields.References,
			}
			got, err := b.findReferences(tt.args.cmd, tt.args.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("branch.findReferences() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("branch.findReferences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateReferencesFromGrep(t *testing.T) {
	type args struct {
		flags      []string
		grepResult [][]string
	}
	tests := []struct {
		name string
		args args
		want []reference
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateReferencesFromGrep(tt.args.flags, tt.args.grepResult); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateReferencesFromGrep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findReferencedFlags(t *testing.T) {
	type args struct {
		ref   string
		flags []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findReferencedFlags(tt.args.ref, tt.args.flags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findReferencedFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makeTimestamp(t *testing.T) {
	tests := []struct {
		name string
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeTimestamp(); got != tt.want {
				t.Errorf("makeTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fatal(t *testing.T) {
	type args struct {
		msg string
		err error
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fatal(tt.args.msg, tt.args.err)
		})
	}
}
