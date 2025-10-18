// internal/spider/logic/a_FD_parse_details.go
package logic

import (
	"context"
	"fmt"
	"rudy_gc/internal/consts"
	"rudy_gc/pkg/ptr"
	"time"

	"rudy_gc/internal/types"
)

func (l *CrawlLogic) ParseDetails(ctx context.Context) error {
	// 查找需要解析的 Item
	items, err := l.deps.ItemRepo.FindByDetailNeedScan(ctx, consts.ItemDetailStatusNeedScan)
	if err != nil {
		return fmt.Errorf("查询待解析的 Item 失败: %w", err)
	}

	total := len(items)
	if total == 0 {
		l.deps.Log.WithContext(ctx).Info("没有需要解析的 Detail")
		return nil
	}
	l.deps.Log.Infof("共有 %d 个 Item 需要解析 Detail", total)

	var done int
	//todo: parallel parse
	for _, it := range items {
		// 解析并入库
		if err := l.parseDetailAndInsertMovie(ctx, it); err != nil {
			return fmt.Errorf("解析并入库失败 %s(%s): %w", it.Name, it.JavId, err)
		}

		// 更新 item.DetailNeedScan -> 已解析
		now := time.Now().Unix()
		patch := types.ItemPatch{
			DetailNeedScan: ptr.Int64(consts.ItemDetailStatusNoNeedScan),
			UpdatedOn:      &now,
		}
		err := l.deps.ItemRepo.UpdatePartialByJavId(ctx, it.JavId, patch)
		if err != nil {
			return fmt.Errorf("更新 Item 详情状态失败 %s: %w", it.Name, err)
		}

		done++
		l.deps.Log.Infof("%s 处理Detail，已完成 %d/%d", it.Name, done, total)
	}

	return nil
}

func (l *CrawlLogic) parseDetailAndInsertMovie(ctx context.Context, it *types.Item) error {
	//build rawMovie
	rawJavMovie, err := l.buildRawMovieByDetail(ctx, it)
	if err != nil {
		return err
	}

	//insert raw
	_, err = l.saveParsedMovie(ctx, rawJavMovie)
	if err != nil {
		return err
	}
	return nil
}
