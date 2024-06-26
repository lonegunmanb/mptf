package backup

import (
	filesystem "github.com/Azure/mapotf/pkg/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"

	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
)

func TestMetaProgrammingTFPlan_OnlyTransformThatHasTargetShouldBeInThePlan(t *testing.T) {
	dir := "cfg"
	expectedContent := `resource "fake_resource" this {
}`
	stub := gostub.Stub(&filesystem.Fs, fakeFs(map[string]string{
		filepath.Join(dir, "main.tf"):                expectedContent,
		filepath.Join(dir, "non-terraform-file.txt"): "",
		filepath.Join("etc", "terraform.tf"):         "should_not_be_copied",
	}))
	defer stub.Reset()
	err := BackupFolder(dir)
	require.NoError(t, err)
	content, err := afero.ReadFile(filesystem.Fs, filepath.Join(dir, "main.tf"+BackupExtension))
	require.NoError(t, err)
	assert.Equal(t, expectedContent, string(content))
	exists, err := afero.Exists(filesystem.Fs, filepath.Join(dir, "non-terraform-file.txt"+BackupExtension))
	require.NoError(t, err)
	assert.False(t, exists)
	exists, err = afero.Exists(filesystem.Fs, filepath.Join("etc", "terraform.tf"+BackupExtension))
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestBackupFolder_BackupFileAlreadyExists(t *testing.T) {
	dir := "cfg"
	originalContent := `resource "fake_resource" this {
}`
	backupContent := `resource "fake_resource" this {
} backup`
	stub := gostub.Stub(&filesystem.Fs, fakeFs(map[string]string{
		filepath.Join(dir, "main.tf"):                 originalContent,
		filepath.Join(dir, "main.tf"+BackupExtension): backupContent,
	}))
	defer stub.Reset()
	err := BackupFolder(dir)
	require.NoError(t, err)
	content, err := afero.ReadFile(filesystem.Fs, filepath.Join(dir, "main.tf"+BackupExtension))
	require.NoError(t, err)
	assert.Equal(t, backupContent, string(content))
}

func TestRestoreBackup(t *testing.T) {
	dir := "cfg"
	originalContent := `resource "fake_resource" this {
}`
	backupContent := `resource "fake_resource" this {
} backup`
	stub := gostub.Stub(&filesystem.Fs, fakeFs(map[string]string{
		filepath.Join(dir, "main.tf"):                 originalContent,
		filepath.Join(dir, "main.tf"+BackupExtension): backupContent,
	}))
	defer stub.Reset()
	err := Reset(dir)
	require.NoError(t, err)
	content, err := afero.ReadFile(filesystem.Fs, filepath.Join(dir, "main.tf"))
	require.NoError(t, err)
	assert.Equal(t, backupContent, string(content))
	exists, err := afero.Exists(filesystem.Fs, filepath.Join(dir, "main.tf"+BackupExtension))
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestClearBackup_NewFileShouldBeRemoved(t *testing.T) {
	dir := "cfg"
	newFileContent := `resource "new_fake_resource" this {
}`
	stub := gostub.Stub(&filesystem.Fs, fakeFs(map[string]string{
		filepath.Join(dir, "main.tf"):                  newFileContent,
		filepath.Join(dir, "main.tf"+NewFileExtension): "",
	}))
	defer stub.Reset()
	err := ClearBackup(dir)
	require.NoError(t, err)
	exists, err := afero.Exists(filesystem.Fs, filepath.Join(dir, "main.tf"+NewFileExtension))
	require.NoError(t, err)
	assert.False(t, exists)
	exists, err = afero.Exists(filesystem.Fs, filepath.Join(dir, "main.tf"))
	require.NoError(t, err)
	assert.True(t, exists)
	content, err := afero.ReadFile(filesystem.Fs, filepath.Join(dir, "main.tf"))
	require.NoError(t, err)
	assert.Equal(t, newFileContent, string(content))
}

func TestReset_NewFileShouldBeRemoved(t *testing.T) {
	dir := "cfg"
	newFileContent := `resource "new_fake_resource" this {
}`
	stub := gostub.Stub(&filesystem.Fs, fakeFs(map[string]string{
		filepath.Join(dir, "main.tf"):                  newFileContent,
		filepath.Join(dir, "main.tf"+NewFileExtension): "",
	}))
	defer stub.Reset()
	err := Reset(dir)
	require.NoError(t, err)
	exists, err := afero.Exists(filesystem.Fs, filepath.Join(dir, "main.tf"+NewFileExtension))
	require.NoError(t, err)
	assert.False(t, exists)
	exists, err = afero.Exists(filesystem.Fs, filepath.Join(dir, "main.tf"))
	require.NoError(t, err)
	assert.False(t, exists)
}

func fakeFs(files map[string]string) afero.Fs {
	fs := afero.NewMemMapFs()
	for n, content := range files {
		_ = afero.WriteFile(fs, n, []byte(content), 0644)
	}
	return fs
}

func TestClearBackup(t *testing.T) {
	dir := "cfg"
	stub := gostub.Stub(&filesystem.Fs, fakeFs(map[string]string{
		filepath.Join(dir, "main.tf"):                 "terraform content",
		filepath.Join(dir, "main.tf"+BackupExtension): "backupContent",
	}))
	defer stub.Reset()
	err := ClearBackup(dir)
	require.NoError(t, err)
	exists, err := afero.Exists(filesystem.Fs, filepath.Join(dir, "main.tf"+BackupExtension))
	require.NoError(t, err)
	assert.False(t, exists)
	exists, err = afero.Exists(filesystem.Fs, filepath.Join(dir, "main.tf"))
	require.NoError(t, err)
	assert.True(t, exists)
}
