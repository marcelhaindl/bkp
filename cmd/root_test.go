package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBackupDirectory verifies the behavior of the backupDirectory function.
//
// It covers a variety of scenarios to ensure directory backup works as expected:
//   - Copies files and subdirectories correctly.
//   - Returns an error when the destination is inside the source directory.
//   - Returns an error when the source directory does not exist.
//   - Handles the case where the source is a single file instead of a directory.
//
// Each test case provides:
//   - setupFunc: to prepare isolated source and destination paths for testing.
//   - expectedErr: whether an error is expected for the case.
//   - verifyFunc: optional assertions to confirm that files were copied correctly.
//
// The test uses t.TempDir() to create temporary, automatically cleaned-up directories,
// and require from testify to handle assertions safely and readably.
func TestBackupDirectory(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (src, dst string)
		expectedErr bool
		verifyFunc  func(t *testing.T, src, dst string)
	}{
		{
			name: "copy directory with files and subdirectories",
			setupFunc: func(t *testing.T) (string, string) {
				srcDir := t.TempDir()
				subDir := filepath.Join(srcDir, "subdir")
				err := os.Mkdir(subDir, 0o755)
				require.NoError(t, err)

				err = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("File 1"), 0o644)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("File 2"), 0o644)
				require.NoError(t, err)

				dstDir := t.TempDir()
				return srcDir, dstDir
			},
			expectedErr: false,
			verifyFunc: func(t *testing.T, src, dst string) {
				data, err := os.ReadFile(filepath.Join(dst, "file1.txt"))
				require.NoError(t, err)
				require.Equal(t, "File 1", string(data))

				data, err = os.ReadFile(filepath.Join(dst, "subdir", "file2.txt"))
				require.NoError(t, err)
				require.Equal(t, "File 2", string(data))
			},
		},
		{
			name: "dst inside src should error",
			setupFunc: func(t *testing.T) (string, string) {
				srcDir := t.TempDir()
				dstDir := filepath.Join(srcDir, "backup")
				return srcDir, dstDir
			},
			expectedErr: true,
		},
		{
			name: "src directory does not exist",
			setupFunc: func(t *testing.T) (string, string) {
				srcDir := filepath.Join(t.TempDir(), "nonexistent")
				dstDir := t.TempDir()
				return srcDir, dstDir
			},
			expectedErr: true,
		},
		{
			name: "src is a file, not a directory",
			setupFunc: func(t *testing.T) (string, string) {
				srcFile := filepath.Join(t.TempDir(), "file.txt")
				err := os.WriteFile(srcFile, []byte("Not a directory"), 0o644)
				require.NoError(t, err)
				dstDir := t.TempDir()
				return srcFile, dstDir
			},
			expectedErr: false,
			verifyFunc: func(t *testing.T, src, dst string) {
				data, err := os.ReadFile(filepath.Join(dst, "file.txt"))
				require.NoError(t, err)
				require.Equal(t, "Not a directory", string(data))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setupFunc(t)
			err := backupDirectory(src, dst)
			require.Equal(t, tt.expectedErr, err != nil)
			if tt.verifyFunc != nil && !tt.expectedErr {
				tt.verifyFunc(t, src, dst)
			}
		})
	}
}

// TestBackupFile verifies the behavior of the backupFile function.
//
// This test ensures that individual file backups work correctly under various scenarios.
// It covers both successful copies and expected failure conditions.
//
// The test includes the following cases:
//   - Copying a file to a new file path.
//   - Copying a file into an existing directory (destination is a folder).
//   - Overwriting an existing destination file with new content.
//   - Handling errors when the source file does not exist.
//
// Each test case defines:
//   - setupFunc: a helper that prepares the source and destination paths (using t.TempDir())
//     and creates necessary files or directories for isolation.
//   - expectedErr: whether an error is expected for the test case.
//   - verifyFunc: optional assertions that confirm the backup result, such as checking
//     that file contents match expectations.
//
// The test uses the "require" package from testify for clear, immediate failure behavior
// and automatic cleanup via t.TempDir().
func TestBackupFile(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (src, dst string)
		expectedErr bool
		verifyFunc  func(t *testing.T, src, dst string)
	}{
		{
			name: "copy file to new file path",
			setupFunc: func(t *testing.T) (string, string) {
				src := filepath.Join(t.TempDir(), "source.txt")
				dst := filepath.Join(t.TempDir(), "dest.txt")
				err := os.WriteFile(src, []byte("Hello, World!"), 0o644)
				require.NoError(t, err)
				return src, dst
			},
			expectedErr: false,
			verifyFunc: func(t *testing.T, src, dst string) {
				data, err := os.ReadFile(dst)
				require.NoError(t, err)
				require.Equal(t, "Hello, World!", string(data))
			},
		},
		{
			name: "copy file to existing directory",
			setupFunc: func(t *testing.T) (string, string) {
				src := filepath.Join(t.TempDir(), "source.txt")
				dstDir := t.TempDir()
				err := os.WriteFile(src, []byte("Hello, Directory!"), 0o644)
				require.NoError(t, err)
				return src, dstDir
			},
			expectedErr: false,
			verifyFunc: func(t *testing.T, src, dst string) {
				destFile := filepath.Join(dst, "source.txt")
				data, err := os.ReadFile(destFile)
				require.NoError(t, err)
				require.Equal(t, "Hello, Directory!", string(data))
			},
		},
		{
			name: "overwrite existing file",
			setupFunc: func(t *testing.T) (string, string) {
				src := filepath.Join(t.TempDir(), "source.txt")
				dst := filepath.Join(t.TempDir(), "dest.txt")
				err := os.WriteFile(src, []byte("New Content"), 0o644)
				require.NoError(t, err)
				err = os.WriteFile(dst, []byte("Old Content"), 0o644)
				require.NoError(t, err)
				return src, dst
			},
			expectedErr: false,
			verifyFunc: func(t *testing.T, src, dst string) {
				data, err := os.ReadFile(dst)
				require.NoError(t, err)
				require.Equal(t, "New Content", string(data))
			},
		},
		{
			name: "src file does not exist",
			setupFunc: func(t *testing.T) (string, string) {
				src := filepath.Join(t.TempDir(), "nonexistent.txt")
				dst := filepath.Join(t.TempDir(), "dest.txt")
				return src, dst
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setupFunc(t)
			err := backupFile(src, dst)
			require.Equal(t, tt.expectedErr, err != nil)
			if tt.verifyFunc != nil && !tt.expectedErr {
				tt.verifyFunc(t, src, dst)
			}
		})
	}
}

// TestEnsureDstOutsideSrc verifies the behavior of the ensureDstOutsideSrc function.
//
// This function checks that a destination path is not inside the source path.
// The test covers several scenarios:
//   - Destination is inside the source directory (should return an error).
//   - Destination is the same as the source directory (should return an error).
//   - Destination is outside the source directory (should succeed).
//   - Relative paths where the destination is inside the source (error expected).
//   - Relative paths where the destination is outside the source (success expected).
//
// Each test case specifies:
//   - src: the source path to test against.
//   - dst: the destination path to check.
//   - expectErr: whether ensureDstOutsideSrc is expected to return an error.
//
// The test uses t.Run to isolate each case and require.Equal for clear assertion.
func TestEnsureDstOutsideSrc(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		dst       string
		expectErr bool
	}{
		{
			name:      "dst inside src",
			src:       "/home/user/docs",
			dst:       "/home/user/docs/backup",
			expectErr: true,
		},
		{
			name:      "dst same as src",
			src:       "/home/user/docs",
			dst:       "/home/user/docs",
			expectErr: true,
		},
		{
			name:      "dst outside src",
			src:       "/home/user/docs",
			dst:       "/home/user/backup",
			expectErr: false,
		},
		{
			name:      "relative paths with dst inside src",
			src:       "docs",
			dst:       "docs/backup",
			expectErr: true,
		},
		{
			name:      "relative paths with dst outside src",
			src:       "docs",
			dst:       "backup",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ensureDstOutsideSrc(tt.src, tt.dst)
			require.Equal(t, tt.expectErr, err != nil)
		})
	}
}
