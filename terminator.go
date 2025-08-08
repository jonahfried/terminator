package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	entry := getEntryDir()
	ignored := map[string]struct{}{".git": {}}

	unterminatedFiles := terminateDir(entry, ignored)
	slices.Sort(unterminatedFiles)

	display(unterminatedFiles)
}

func getEntryDir() string {
	if len(os.Args) > 1 {
		dir := os.Args[1]

		if _, err := os.ReadDir(dir); err == nil {
			return dir
		}
	}

	return "."
}

func terminateDir(dirName string, ignored map[string]struct{}) []string {
	for k := range getIgnored(dirName) {
		ignored[filepath.Join(dirName, k)] = struct{}{}
	}
	fmt.Println(ignored)

	currentDir, err := os.ReadDir(dirName)
	if err != nil {
		return nil
	}

	failedTerminations := make([]string, 0, len(currentDir))

	for _, e := range currentDir {
		currentPath := filepath.Join(dirName, e.Name())
		if _, ok := ignored[currentPath]; ok {
			continue
		}

		failedTerminations = append(
			failedTerminations,
			terminateEntry(dirName, e, ignored)...,
		)
	}

	return failedTerminations
}

func terminateEntry(path string, entry os.DirEntry, ignored map[string]struct{}) []string {
	entryPath := filepath.Join(path, entry.Name())

	if entry.IsDir() {
		return terminateDir(entryPath, ignored)
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

// func attemptGitignoreFilter(paths []string) []string {
// 	gitignorePath := filepath.Join(getEntryDir(), ".gitignore")
// 	gitignore, err := os.ReadFile(gitignorePath)
// 	if err != nil {
// 		return paths
// 	}

// 	for _, glob := range strings.Split(string(gitignore), "\n") {
//     filepath.Glob()
// 	}
// }
