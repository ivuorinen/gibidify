package fileproc_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ivuorinen/gibidify/config"
	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/testutil"
)

func TestProdWalkerWithIgnore(t *testing.T) {
	// Create a temporary directory structure.
	rootDir := t.TempDir()

	subDir := testutil.CreateTestDirectory(t, rootDir, "vendor")

	// Write sample files
	testutil.CreateTestFiles(t, rootDir, []testutil.FileSpec{
		{Name: "file1.go", Content: "content"},
		{Name: "file2.txt", Content: "content"},
	})
	testutil.CreateTestFile(t, subDir, "file_in_vendor.txt", []byte("content")) // should be ignored

	// .gitignore that ignores *.txt and itself
	gitignoreContent := `*.txt
.gitignore
`
	testutil.CreateTestFile(t, rootDir, ".gitignore", []byte(gitignoreContent))

	// Initialize config to ignore "vendor" directory
	testutil.ResetConfig(t, "")
	config.Set("ignoreDirectories", []string{"vendor"})

	// Run walker
	w := fileproc.NewProdWalker(fileproc.WalkOptions{})
	found, err := w.Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")

	// We expect only file1.go to appear
	if len(found) != 1 {
		t.Errorf("Expected 1 file to pass filters, got %d: %v", len(found), found)
	}
	if len(found) == 1 && filepath.Base(found[0]) != "file1.go" {
		t.Errorf("Expected file1.go, got %s", found[0])
	}
}

func TestProdWalkerBinaryCheck(t *testing.T) {
	rootDir := t.TempDir()

	// Create test files
	testutil.CreateTestFiles(t, rootDir, []testutil.FileSpec{
		{Name: "somefile.exe", Content: "fake-binary-content"},
		{Name: "keep.go", Content: "package main"},
	})

	// Reset and load default config
	testutil.ResetConfig(t, "")

	// Reset FileTypeRegistry to ensure clean state
	fileproc.ResetRegistryForTesting()

	// Run walker
	w := fileproc.NewProdWalker(fileproc.WalkOptions{})
	found, err := w.Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")

	// Only "keep.go" should be returned
	if len(found) != 1 {
		t.Errorf("Expected 1 file, got %d: %v", len(found), found)
	}
	if len(found) == 1 && filepath.Base(found[0]) != "keep.go" {
		t.Errorf("Expected keep.go in results, got %s", found[0])
	}
}

func TestProdWalkerSizeLimit(t *testing.T) {
	rootDir := t.TempDir()

	// Create test files
	largeFileData := make([]byte, 6*1024*1024) // 6 MB
	testutil.CreateTestFile(t, rootDir, "largefile.txt", largeFileData)
	testutil.CreateTestFile(t, rootDir, "smallfile.go", []byte("package main"))

	// Reset and load default config, which sets size limit to 5 MB
	testutil.ResetConfig(t, "")

	w := fileproc.NewProdWalker(fileproc.WalkOptions{})
	found, err := w.Walk(rootDir)
	if err != nil {
		t.Fatalf("Walk returned error: %v", err)
	}

	// We should only get the small file
	if len(found) != 1 {
		t.Errorf("Expected 1 file under size limit, got %d", len(found))
	}
	if len(found) == 1 && filepath.Base(found[0]) != "smallfile.go" {
		t.Errorf("Expected smallfile.go, got %s", found[0])
	}
}

// relNames returns the walker results relative to root, for readable assertions.
func relNames(t *testing.T, root string, found []string) []string {
	t.Helper()
	names := make([]string, 0, len(found))
	for _, f := range found {
		rel, err := filepath.Rel(root, f)
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, filepath.ToSlash(rel))
	}

	return names
}

// TestProdWalkerSkipsSymlinks verifies links to files and directories outside
// the tree are never returned: following them leaks out-of-tree files into the output.
func TestProdWalkerSkipsSymlinks(t *testing.T) {
	testutil.ResetConfig(t, "")
	outside := t.TempDir()
	testutil.CreateTestFile(t, outside, "secret.txt", []byte("secret"))

	rootDir := t.TempDir()
	testutil.CreateTestFile(t, rootDir, "keep.go", []byte("package main"))
	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(rootDir, "link.txt")); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(rootDir, "linkdir")); err != nil {
		t.Fatal(err)
	}

	found, err := fileproc.NewProdWalker(fileproc.WalkOptions{}).Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")

	if got := relNames(t, rootDir, found); !slices.Equal(got, []string{"keep.go"}) {
		t.Errorf("expected only keep.go, got %v", got)
	}
}

// TestProdWalkerPrunesGitignoredDirectories verifies a gitignored directory is
// not descended into, so an unreadable directory inside it cannot fail the walk.
func TestProdWalkerPrunesGitignoredDirectories(t *testing.T) {
	testutil.ResetConfig(t, "")
	rootDir := t.TempDir()
	testutil.CreateTestFile(t, rootDir, "keep.go", []byte("package main"))
	testutil.CreateTestFile(t, rootDir, ".gitignore", []byte("venv/\n"))
	venv := testutil.CreateTestDirectory(t, rootDir, "venv")
	testutil.CreateTestFile(t, venv, "lib.py", []byte("x = 1"))
	// Empty, so t.TempDir's cleanup can remove it without restoring permissions.
	if err := os.Mkdir(filepath.Join(venv, "locked"), 0o000); err != nil {
		t.Fatal(err)
	}

	found, err := fileproc.NewProdWalker(fileproc.WalkOptions{}).Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")

	if got := relNames(t, rootDir, found); !slices.Equal(got, []string{".gitignore", "keep.go"}) {
		t.Errorf("expected .gitignore and keep.go, got %v", got)
	}
}

// TestProdWalkerNestedNegationReincludes verifies a deeper .gitignore's
// "!pattern" re-includes a file a parent .gitignore ignores, as git does. The
// parent rule used to win because any matching rule ignored the path.
func TestProdWalkerNestedNegationReincludes(t *testing.T) {
	testutil.ResetConfig(t, "")
	rootDir := t.TempDir()
	testutil.CreateTestFile(t, rootDir, ".gitignore", []byte("*.log\n"))
	sub := testutil.CreateTestDirectory(t, rootDir, "sub")
	testutil.CreateTestFile(t, sub, ".gitignore", []byte("!keep.log\n"))
	testutil.CreateTestFile(t, sub, "keep.log", []byte("kept"))
	testutil.CreateTestFile(t, sub, "other.log", []byte("dropped"))
	testutil.CreateTestFile(t, rootDir, "root.log", []byte("dropped"))

	found, err := fileproc.NewProdWalker(fileproc.WalkOptions{}).Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")

	want := []string{".gitignore", "sub/.gitignore", "sub/keep.log"}
	if got := relNames(t, rootDir, found); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestProdWalkerUnreadableIgnoreFile verifies an ignore file that cannot be read
// is never treated as absent: below the root its directory is skipped, at the
// root the walk fails. It used to be ignored, so the files it excluded were
// collected.
func TestProdWalkerUnreadableIgnoreFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read mode-000 files")
	}
	testutil.ResetConfig(t, "")

	t.Run("subdirectory is skipped", func(t *testing.T) {
		rootDir := t.TempDir()
		testutil.CreateTestFile(t, rootDir, "keep.go", []byte("package main"))
		sub := testutil.CreateTestDirectory(t, rootDir, "sub")
		testutil.CreateTestFile(t, sub, "secret.env.txt", []byte("TOKEN=1"))
		ignorePath := testutil.CreateTestFile(t, sub, ".gitignore", []byte("secret.env.txt\n"))
		if err := os.Chmod(ignorePath, 0o000); err != nil {
			t.Fatal(err)
		}

		found, err := fileproc.NewProdWalker(fileproc.WalkOptions{}).Walk(rootDir)
		testutil.MustSucceed(t, err, "walking directory")
		if got := relNames(t, rootDir, found); !slices.Equal(got, []string{"keep.go"}) {
			t.Errorf("expected only keep.go, got %v", got)
		}
	})

	t.Run("root fails the walk", func(t *testing.T) {
		rootDir := t.TempDir()
		testutil.CreateTestFile(t, rootDir, "keep.go", []byte("package main"))
		ignorePath := testutil.CreateTestFile(t, rootDir, ".gitignore", []byte("*.env\n"))
		if err := os.Chmod(ignorePath, 0o000); err != nil {
			t.Fatal(err)
		}

		_, err := fileproc.NewProdWalker(fileproc.WalkOptions{}).Walk(rootDir)
		testutil.AssertErrorContains(t, err, "cannot read ignore file", "walking with unreadable root .gitignore")
	})
}

// TestProdWalkerSkipsUnreadableSubdirectory verifies an unreadable directory
// that is not ignored is skipped instead of aborting the whole walk.
func TestProdWalkerSkipsUnreadableSubdirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read mode-000 directories")
	}
	testutil.ResetConfig(t, "")
	rootDir := t.TempDir()
	testutil.CreateTestFile(t, rootDir, "keep.go", []byte("package main"))
	// Empty, so t.TempDir's cleanup can remove it without restoring permissions.
	if err := os.Mkdir(filepath.Join(rootDir, "locked"), 0o000); err != nil {
		t.Fatal(err)
	}

	found, err := fileproc.NewProdWalker(fileproc.WalkOptions{}).Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")

	if got := relNames(t, rootDir, found); !slices.Equal(got, []string{"keep.go"}) {
		t.Errorf("expected only keep.go, got %v", got)
	}
}

// TestProdWalkerExcludesPaths verifies excluded absolute paths (the output
// file) are not returned, so a re-run does not ingest its own previous output.
func TestProdWalkerExcludesPaths(t *testing.T) {
	testutil.ResetConfig(t, "")
	rootDir := t.TempDir()
	testutil.CreateTestFile(t, rootDir, "keep.go", []byte("package main"))
	testutil.CreateTestFile(t, rootDir, "out.json", []byte("{}"))

	found, err := fileproc.NewProdWalker(
		fileproc.WalkOptions{Exclude: []string{filepath.Join(rootDir, "out.json")}},
	).Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")

	if got := relNames(t, rootDir, found); !slices.Equal(got, []string{"keep.go"}) {
		t.Errorf("expected only keep.go, got %v", got)
	}
}

// TestProdWalkerSkipsSensitiveFiles verifies credential files are left out and
// counted by default and kept with IncludeSensitive. They used to reach the
// LLM-bound output whenever a .gitignore did not happen to cover them.
func TestProdWalkerSkipsSensitiveFiles(t *testing.T) {
	testutil.ResetConfig(t, "")
	rootDir := t.TempDir()
	sensitive := []string{".env", ".env.local", "server.key", "cert.pem", "id_rsa", "id_ed25519.pub", ".npmrc"}
	for _, name := range append([]string{"main.go"}, sensitive...) {
		testutil.CreateTestFile(t, rootDir, name, []byte("x"))
	}

	w := fileproc.NewProdWalker(fileproc.WalkOptions{})
	found, err := w.Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")
	if got := relNames(t, rootDir, found); !slices.Equal(got, []string{"main.go"}) {
		t.Errorf("expected only main.go, got %v", got)
	}
	if w.Skipped() != len(sensitive) {
		t.Errorf("Skipped() = %d, want %d", w.Skipped(), len(sensitive))
	}

	found, err = fileproc.NewProdWalker(fileproc.WalkOptions{IncludeSensitive: true}).Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory with IncludeSensitive")
	if len(found) != len(sensitive)+1 {
		t.Errorf("IncludeSensitive: got %d files, want %d", len(found), len(sensitive)+1)
	}
}

// TestProdWalkerAlwaysSkipsVCSDirectories verifies .git stays excluded when the
// user replaces ignoreDirectories. .git used to be an entry of the replaceable
// default list, so adding one directory re-included .git/config and its
// token-bearing remote URLs.
func TestProdWalkerAlwaysSkipsVCSDirectories(t *testing.T) {
	testutil.SetConfigKeys(t, map[string]any{"ignoreDirectories": []string{"secrets"}})
	rootDir := t.TempDir()
	testutil.CreateTestFile(t, rootDir, "main.go", []byte("package main"))
	for _, vcs := range []string{".git", ".hg", ".svn"} {
		dir := testutil.CreateTestDirectory(t, rootDir, vcs)
		testutil.CreateTestFile(t, dir, "config", []byte("[remote]"))
	}

	found, err := fileproc.NewProdWalker(fileproc.WalkOptions{}).Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")
	if got := relNames(t, rootDir, found); !slices.Equal(got, []string{"main.go"}) {
		t.Errorf("expected only main.go, got %v", got)
	}
}

// TestProdWalkerTrimsIgnoredDirectoryNames verifies an ignoreDirectories entry
// with surrounding spaces, which validation accepts after trimming, still
// matches the directory.
func TestProdWalkerTrimsIgnoredDirectoryNames(t *testing.T) {
	testutil.SetConfigKeys(t, map[string]any{"ignoreDirectories": []string{" skipme "}})
	rootDir := t.TempDir()
	testutil.CreateTestFile(t, rootDir, "main.go", []byte("package main"))
	testutil.CreateTestFile(t, testutil.CreateTestDirectory(t, rootDir, "skipme"), "a.go", []byte("package a"))

	found, err := fileproc.NewProdWalker(fileproc.WalkOptions{}).Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")
	if got := relNames(t, rootDir, found); !slices.Equal(got, []string{"main.go"}) {
		t.Errorf("expected only main.go, got %v", got)
	}
}

// TestProdWalkerExcludesByIdentity verifies the output file is excluded when
// the source is walked through a symlinked root and the destination is named by
// its real path, so the two path strings differ.
func TestProdWalkerExcludesByIdentity(t *testing.T) {
	testutil.ResetConfig(t, "")
	base := t.TempDir()
	realDir := testutil.CreateTestDirectory(t, base, "real")
	testutil.CreateTestFile(t, realDir, "keep.go", []byte("package main"))
	outPath := testutil.CreateTestFile(t, realDir, "out.json", []byte("{}"))
	link := filepath.Join(base, "link")
	if err := os.Symlink(realDir, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}

	found, err := fileproc.NewProdWalker(fileproc.WalkOptions{Exclude: []string{outPath}}).Walk(link)
	testutil.MustSucceed(t, err, "walking symlinked root")
	if got := relNames(t, link, found); !slices.Equal(got, []string{"keep.go"}) {
		t.Errorf("expected only keep.go, got %v", got)
	}
}

// TestProdWalkerSkipsFIFO verifies non-regular files are skipped: reading a
// FIFO blocks until a writer appears, which would hang a worker forever.
func TestProdWalkerSkipsFIFO(t *testing.T) {
	testutil.ResetConfig(t, "")
	rootDir := t.TempDir()
	testutil.CreateTestFile(t, rootDir, "keep.go", []byte("package main"))
	if err := mkfifo(filepath.Join(rootDir, "pipe")); err != nil {
		t.Skipf("mkfifo unsupported: %v", err)
	}

	found, err := fileproc.NewProdWalker(fileproc.WalkOptions{}).Walk(rootDir)
	testutil.MustSucceed(t, err, "walking directory")

	if got := relNames(t, rootDir, found); !slices.Equal(got, []string{"keep.go"}) {
		t.Errorf("expected only keep.go, got %v", got)
	}
}
