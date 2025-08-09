package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type terminationChecker struct {
	startDir string

	ignored      map[string]struct{}
	ignoreHidden bool
	quiet        bool
	noGitIgnore  bool
}

func newTerminationChecker() *terminationChecker {
	a := terminationChecker{
		ignored: make(map[string]struct{}),
	}
	var ignoreGlobs string

	flag.BoolVar(&(a.ignoreHidden), "ignoreHidden", false, "whether to evaluate hidden files")
	flag.BoolVar(&(a.quiet), "q", false, "quiet mode (don't print bad files, just return exit code)")
	flag.BoolVar(&(a.noGitIgnore), "no-ignore", false, "whether to disregard .gitignore files")
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
	a := newTerminationChecker()
	unterminatedFiles := a.launchChecker()

	if !a.quiet {
		display(unterminatedFiles)
	}

	if len(unterminatedFiles) > 0 {
		os.Exit(1)
	}

	os.Exit(0)
}

func (a *terminationChecker) launchChecker() []string {
	unterminatedFiles := a.checkDir(a.startDir)
	slices.Sort(unterminatedFiles)
	return unterminatedFiles
}

func (a *terminationChecker) checkDir(dirName string) []string {
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
			a.checkEntry(dirName, e)...,
		)
	}

	return failedTerminations
}

func (a *terminationChecker) ignore(dirName string, entry os.DirEntry) bool {
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

func (a *terminationChecker) extendIgnored(dirName string) {
	if a.noGitIgnore {
		return
	}

	for k := range getIgnored(dirName) {
		a.ignored[k] = struct{}{}
	}
}

func (a *terminationChecker) checkEntry(path string, entry os.DirEntry) []string {
	entryPath := filepath.Join(path, entry.Name())

	if entry.IsDir() {
		return a.checkDir(entryPath)
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

	gitignorePath := filepath.Join(dir, ".gitignore")
	gitignore, err := os.ReadFile(gitignorePath)
	if err != nil {
		return map[string]struct{}{}
	}

	for _, glob := range strings.Split(string(gitignore), "\n") {
		glob = strings.TrimSpace(glob)
		if len(glob) == 0 || glob[0] == '#' {
			continue
		}

		matches, err := filepath.Glob(filepath.Join(dir, glob))
		if err != nil {
			continue
		}

		for _, m := range matches {
			ignored[m] = struct{}{}
		}
	}

	return ignored
}
