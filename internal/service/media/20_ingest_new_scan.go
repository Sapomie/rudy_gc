package media

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type rootLayout struct {
	rootDir   string
	process   string
	ingestNew string
	tmp       string
	failed    string
	rollback  string
	media     string
}

func buildRootLayout(root string) rootLayout {
	process := filepath.Join(root, processDirName)
	return rootLayout{
		rootDir:   root,
		process:   process,
		ingestNew: filepath.Join(process, ingestNewDirName),
		tmp:       filepath.Join(process, tmpDirName),
		failed:    filepath.Join(process, failedDirName),
		rollback:  filepath.Join(process, rollbackDirName),
		media:     filepath.Join(root, mediaDirName),
	}
}

func ensureRootLayout(layout rootLayout) error {
	dirs := []string{
		layout.ingestNew,
		layout.tmp,
		layout.failed,
		layout.rollback,
		layout.media,
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, defaultFilePerm); err != nil {
			return err
		}
	}
	return nil
}

func listIngestNewFiles(layout rootLayout) ([]string, error) {
	return listProcessVideoFiles(layout.ingestNew)
}

func listRollbackFiles(layout rootLayout) ([]string, error) {
	return listProcessVideoFiles(layout.rollback)
}

func listProcessVideoFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if shouldSkipIngestEntryName(name) {
			continue
		}

		info, infoErr := entry.Info()
		if infoErr != nil {
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		if info.Size() < minMediaFileSize {
			continue
		}
		if !isVideoName(name) {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	sort.Strings(files)
	return files, nil
}

func shouldSkipIngestEntryName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return true
	}
	// macOS AppleDouble 文件（._*）与隐藏文件都不应进入 media ingest。
	return strings.HasPrefix(name, ".")
}
