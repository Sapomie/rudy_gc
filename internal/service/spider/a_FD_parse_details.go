// internal/spider/logic/a_FD_parse_details.go
package spider

import (
	"context"
	"fmt"
	"rudy_gc/internal/consts"
	"rudy_gc/pkg/ptr"
	"time"

	"rudy_gc/internal/types"
)

func (l *CrawlLogic) ParseDetails(ctx context.Context) (*affectedMovieNumbers, error) {
	log := l.deps.Log.WithContext(ctx)
	// 查找需要解析的 Item
	items, err := l.deps.ItemRepo.FindByDetailNeedScan(ctx, consts.ItemDetailStatusNeedScan)
	if err != nil {
		return nil, fmt.Errorf("查询待解析的 Item 失败: %w", err)
	}

	total := len(items)
	if total == 0 {
		log.Info("没有需要解析的 Detail")
		return newAffectedMovieNumbers(), nil
	}
	log.Infof("共有 %d 个 Item 需要解析 Detail", total)

	affected := newAffectedMovieNumbers()
	var done int
	for _, it := range items {
		resp, err := l.handleDetailParse(ctx, it)
		if err != nil {
			return nil, err
		}
		affected.addFromResponse(resp)
		done++
		log.Infof("%s 处理Detail，已完成 %d/%d", it.Name, done, total)
	}
	return affected, nil
}

// handleDetailParse: 负责解析详情并更新状态
func (l *CrawlLogic) handleDetailParse(ctx context.Context, it *types.Item) (*saveParsedMovieResponse, error) {
	// 解析并入库
	resp, err := l.parseDetailAndInsertMovie(ctx, it)
	if err != nil {
		return nil, fmt.Errorf("解析并入库失败 %s(%s): %w", it.Name, it.JavId, err)
	}

	// 更新 item.DetailNeedScan -> 已解析
	now := time.Now().Unix()
	patch := types.ItemPatch{
		DetailNeedScan: ptr.Int64(consts.ItemDetailStatusNoNeedScan),
		UpdatedOn:      &now,
	}
	if err := l.deps.ItemRepo.UpdatePartialByJavId(ctx, it.JavId, patch); err != nil {
		return nil, fmt.Errorf("更新 Item 详情状态失败 %s: %w", it.Name, err)
	}

	return resp, nil
}
func (l *CrawlLogic) parseDetailAndInsertMovie(ctx context.Context, it *types.Item) (*saveParsedMovieResponse, error) {
	//build rawMovie
	rawJavMovie, err := l.buildRawMovieByDetail(ctx, it)
	if err != nil {
		return nil, err
	}

	//insert raw
	resp, err := l.saveParsedMovie(ctx, rawJavMovie)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
