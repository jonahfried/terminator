package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type app struct {
	startDir string

	ignored      map[string]struct{}
	ignoreHidden bool
}

func newApp() *app {
	a := app{
		ignored: make(map[string]struct{}),
	}
	var ignoreGlobs string

	flag.BoolVar(&(a.ignoreHidden), "ignoreHidden", false, "whether to evaluate hidden files")
	flag.StringVar(&(a.startDir), "dir", ".", "directory to terminate")
	flag.StringVar(&ignoreGlobs, "ignore", "", "comma seperated list of globs to ignore")
	flag.Parse()

	cwd, err := os.Getwd()
	_ = err
	os.Chdir(a.startDir)
	for _, g := range strings.Split(ignoreGlobs, ",") {
		matched, err := filepath.Glob(g)
		_ = err

		for _, m := range matched {
			a.ignored[filepath.Join(a.startDir, m)] = struct{}{}
		}

	}
	os.Chdir(cwd)
	return &a
}

func main() {
	a := newApp()
	unterminatedFiles := a.launchTerminator()
	display(unterminatedFiles)

	if len(unterminatedFiles) > 0 {
		os.Exit(1)
	}

	os.Exit(0)
}

func (a *app) launchTerminator() []string {
	unterminatedFiles := a.terminateDir(a.startDir)
	slices.Sort(unterminatedFiles)
	return unterminatedFiles
}

func (a *app) terminateDir(dirName string) []string {
	a.extendIgnored(dirName)

	currentDir, err := os.ReadDir(dirName)
	if err != nil {
		return nil
	}

	failedTerminations := make([]string, 0, len(currentDir))

	for _, e := range currentDir {
		if a.ignore(dirName, e) {
			continue
		}

		failedTerminations = append(
			failedTerminations,
			a.terminateEntry(dirName, e)...,
		)
	}

	return failedTerminations
}

func (a *app) extendIgnored(dirName string) {
	for k := range getIgnored(dirName) {
		a.ignored[filepath.Join(dirName, k)] = struct{}{}
	}
}

func (a *app) ignore(dirName string, entry os.DirEntry) bool {
	name := entry.Name()

	currentPath := filepath.Join(dirName, name)
	if _, ok := a.ignored[currentPath]; ok {
		return true
	}

	if a.ignoreHidden && len(name) > 0 && name[0] == '.' {
		return true
	}

	return false
}

func (a *app) terminateEntry(path string, entry os.DirEntry) []string {
	entryPath := filepath.Join(path, entry.Name())

	if entry.IsDir() {
		return a.terminateDir(entryPath)
	}

	terminated, err := isTerminated(entryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error determining termination for file '%s': %v", path, err)
	}

	if !terminated {
		return []string{entryPath}
	}

	return nil
}

func isTerminated(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("couldn't open file '%s': %w", path, err)
	}

	stat, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("couldn't get file info: %w", err)
	}

	if stat.Size() == 0 { // allow empty files
		return true, nil
	}

	buf := make([]byte, 1)
	_, err = file.ReadAt(buf, stat.Size()-1)
	if err != nil {
		return false, fmt.Errorf("unable to read from file '%s': %w", path, err)
	}

	if buf[0] == '\n' {
		return true, nil
	}

	return false, nil
}

func display(paths []string) {
	for _, p := range paths {
		fmt.Println(p)
	}
}

func getIgnored(dir string) map[string]struct{} {
	ignored := make(map[string]struct{})

	cwd, err := os.Getwd()
	if err != nil {
		return ignored
	}

	gitignorePath := filepath.Join(dir, ".gitignore")
	gitignore, err := os.ReadFile(gitignorePath)
	if err != nil {
		return map[string]struct{}{}
	}

	os.Chdir(dir)

	for _, glob := range strings.Split(string(gitignore), "\n") {
		if len(glob) > 0 && glob[0] == '#' {
			continue
		}

		matches, err := filepath.Glob(glob)
		if err != nil {
			continue
		}

		for _, m := range matches {
			ignored[m] = struct{}{}
		}
	}

	os.Chdir(cwd)

	return ignored
}
