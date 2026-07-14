package media

import (
	"os"
	"path/filepath"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
)

// ResolveFC2CoverImage returns the selected usable capture image, or the newest
// usable capture image when the selected image is empty or invalid.
func ResolveFC2CoverImage(mediaID, selectedImageID int64, imageRows []*moviex.FC2MediaImage) *moviex.FC2MediaImage {
	if mediaID <= 0 {
		return nil
	}
	var newest *moviex.FC2MediaImage
	for _, imageRow := range imageRows {
		if !isUsableFC2CoverImage(mediaID, imageRow) {
			continue
		}
		if selectedImageID > 0 && imageRow.ID == selectedImageID {
			return imageRow
		}
		if newest == nil || imageRow.CreatedOn > newest.CreatedOn || imageRow.CreatedOn == newest.CreatedOn && imageRow.ID > newest.ID {
			newest = imageRow
		}
	}
	return newest
}

func isUsableFC2CoverImage(mediaID int64, imageRow *moviex.FC2MediaImage) bool {
	if imageRow == nil || imageRow.ID <= 0 || imageRow.FC2MediaID != mediaID {
		return false
	}
	if strings.TrimSpace(imageRow.ImageType) != FC2MediaImageTypeContactSheet {
		return false
	}
	filePath := filepath.Clean(strings.TrimSpace(imageRow.FilePath))
	if filePath == "" || filePath == "." {
		return false
	}
	fileInfo, err := os.Stat(filePath)
	return err == nil && fileInfo.Mode().IsRegular()
}
