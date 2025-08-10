package main

import (
	"reflect"
	"testing"
)

func Test_terminationChecker(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "simple_test",
			path: "./testdata/test1/",
			want: []string{"testdata/test1/not1.txt"},
		},
		{
			name: "with_gitignore",
			path: "./testdata/test2/",
			want: []string{"testdata/test2/not2.txt"},
		},
		{
			name: "with_gitignore",
			path: "./testdata/test3/",
			want: []string{"testdata/test3/inner/not3.txt", "testdata/test3/not2.txt"},
		},
		{
			name: "nested_gitignore",
			path: "./testdata/test4/",
			want: []string{"testdata/test4/not1.txt", "testdata/test4/not2.txt"},
		},
	}

	a := terminationChecker{
		startDir:     "",
		ignored:      map[string]struct{}{},
		ignoreHidden: false,
		quiet:        false,
		noGitIgnore:  false,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a.startDir = tt.path
			if got := a.checkDir(tt.path); !reflect.DeepEqual(tt.want, got) {
				t.Errorf("got = %v, want = %v", got, tt.want)
			}
		})
	}
}
