package media

import (
	"context"
	"time"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/service/wfoldertree"
)

func ensureFC2TargetDirectory(ctx context.Context, model moviex.WFolderModel, rootDir string, now time.Time) (string, int64, error) {
	layout := buildRootLayout(rootDir)
	targetDir, err := allocateTargetDirectoryPathUnder(layout.media, now)
	if err != nil {
		return "", 0, err
	}
	folder, err := wfoldertree.EnsurePathChain(ctx, model, consts.WFolderSourceNative, targetDir, now.Unix())
	if err != nil {
		return "", 0, err
	}
	return targetDir, folder.Id, nil
}
