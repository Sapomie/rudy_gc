// internal/spider/logic/a_FD_parse_details.go
package logic

import (
	"fmt"
	"time"

	"rudy_gc/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

import (
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

func (l *CrawlLogic) ParseDetailsParallel() error {
	// 1) 拉取任务
	items, err := l.deps.ItemRepo.FindByDetailNeedScan(l.ctx, types.ItemDetailStatusNeedScan)
	if err != nil {
		return fmt.Errorf("查询待解析的 Item 失败: %w", err)
	}
	total := len(items)
	if total == 0 {
		logx.WithContext(l.ctx).Info("没有需要解析的 Detail")
		return nil
	}
	l.deps.Log.Infof("共有 %d 个 Item 需要解析 Detail", total)

	// 2) 控制并发度（可做成配置，默认 8）
	const maxParallel = 20
	sem := make(chan struct{}, maxParallel)

	// 3) 组 + 可取消上下文
	g, ctx := errgroup.WithContext(l.ctx)

	var done int64

	for _, it := range items {
		it := it // 捕获
		// 启动前先占位；如果 ctx 已取消，直接 break
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			break
		}

		g.Go(func() error {
			defer func() { <-sem }()

			// 3.1 任务前检查取消
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// 3.2 解析并入库
			if _, err := l.parseDetailAndInsertMovie(it); err != nil {
				return fmt.Errorf("解析并入库失败 %s(%s): %w", it.Name, it.JavId, err)
			}

			// 3.3 更新 item 详情元信息
			now := time.Now().Unix()
			if err := l.deps.ItemRepo.UpdateDetailMeta(
				ctx,
				it.Id,
				types.ItemDetailStatusNoNeedScan, // 已完成解析
				it.DetailBirthTime,               // 保留原值
				now,                              // DetailUpdateTime
				now,                              // UpdatedOn
				types.ItemDetailOK,               // HasDetail
			); err != nil {
				return fmt.Errorf("更新 Item 详情状态失败 %s: %w", it.Name, err)
			}

			// 3.4 进度（仅用于日志，不影响并发）
			n := atomic.AddInt64(&done, 1)
			l.deps.Log.Infof("%s 处理Detail，已完成 %d/%d", it.Name, n, total)
			return nil
		})
	}

	// 4) 等待；遇到任一错误会取消其它任务并返回该错误
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (l *CrawlLogic) ParseDetails() error {
	// 查找需要解析的 Item
	items, err := l.deps.ItemRepo.FindByDetailNeedScan(l.ctx, types.ItemDetailStatusNeedScan)
	if err != nil {
		return fmt.Errorf("查询待解析的 Item 失败: %w", err)
	}

	total := len(items)
	if total == 0 {
		logx.WithContext(l.ctx).Info("没有需要解析的 Detail")
		return nil
	}
	l.deps.Log.Infof("共有 %d 个 Item 需要解析 Detail", total)

	var done int
	for _, it := range items {
		// 解析并入库
		if _, err := l.parseDetailAndInsertMovie(it); err != nil {
			return fmt.Errorf("解析并入库失败 %s(%s): %w", it.Name, it.JavId, err)
		}

		// 更新 item.DetailNeedScan -> 已解析
		now := time.Now().Unix()
		if err := l.deps.ItemRepo.UpdateDetailMeta(
			l.ctx,
			it.Id,
			types.ItemDetailStatusNoNeedScan, // 已完成解析
			it.DetailBirthTime,               // 保留原值
			now,                              // DetailUpdateTime
			now,                              // UpdatedOn
			types.ItemDetailOK,               // HasDetail
		); err != nil {
			return fmt.Errorf("更新 Item 详情状态失败 %s: %w", it.Name, err)
		}

		done++
		l.deps.Log.Infof("%s 处理Detail，已完成 %d/%d", it.Name, done, total)
	}

	return nil
}

func (l *CrawlLogic) parseDetailAndInsertMovie(it *types.Item) (interface{}, interface{}) {
	//build rawMovie
	rawJavMovie, err := l.buildRawMovieByDetail(it)
	if err != nil {
		return "", err
	}

	//insert raw
	_, err = l.saveParsedMovie(rawJavMovie)
	if err != nil {
		return "", err
	}
	return nil, nil
}
