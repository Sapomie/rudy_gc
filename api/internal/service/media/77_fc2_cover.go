package media

import (
	"os"
	"path/filepath"
	"strings"

	"rudy_gc/internal/repository/moviex"
)

// ResolveFC2CoverImage returns the selected usable capture image, or the newest
// usable capture image when the selected image is empty or invalid.
func ResolveFC2CoverImage(mediaID, selectedImageID int64, imageRows []*moviex.FC2MediaImage) *moviex.FC2MediaImage {
	return resolveFC2MediaImage(mediaID, selectedImageID, FC2MediaImageTypeContactSheet, imageRows)
}

// ResolveFC2DetailCoverImage returns the newest usable 5x5 detail contact sheet.
func ResolveFC2DetailCoverImage(mediaID int64, imageRows []*moviex.FC2MediaImage) *moviex.FC2MediaImage {
	return resolveFC2MediaImage(mediaID, 0, FC2MediaImageTypeDetailContactSheet, imageRows)
}

func resolveFC2MediaImage(mediaID, selectedImageID int64, imageType string, imageRows []*moviex.FC2MediaImage) *moviex.FC2MediaImage {
	if mediaID <= 0 {
		return nil
	}
	var newest *moviex.FC2MediaImage
	for _, imageRow := range imageRows {
		if !isUsableFC2MediaImage(mediaID, imageType, imageRow) {
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
	return isUsableFC2MediaImage(mediaID, FC2MediaImageTypeContactSheet, imageRow)
}

func isUsableFC2MediaImage(mediaID int64, imageType string, imageRow *moviex.FC2MediaImage) bool {
	if imageRow == nil || imageRow.ID <= 0 || imageRow.FC2MediaID != mediaID {
		return false
	}
	if strings.TrimSpace(imageRow.ImageType) != imageType {
		return false
	}
	filePath := filepath.Clean(strings.TrimSpace(imageRow.FilePath))
	if filePath == "" || filePath == "." {
		return false
	}
	fileInfo, err := os.Stat(filePath)
	return err == nil && fileInfo.Mode().IsRegular()
}
