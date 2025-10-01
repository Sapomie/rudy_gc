// internal/spider/logic/a_FD_parse_details.go
package logic

import (
	"fmt"
	"time"

	"rudy_gc/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// ParseDetails
// 1) 找到 DetailNeedScan=YES 的 item 列表
// 2) 逐条抓取并解析 -> 生成/更新电影相关数据
// 3) 成功后更新 item.DetailNeedScan
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
	logx.WithContext(l.ctx).Infof("共有 %d 个 Item 需要解析 Detail", total)

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
		logx.WithContext(l.ctx).Infof("%s 处理Detail，已完成 %d/%d，还剩 %d 部 movie",
			it.Name, done, total, total-done)
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
	//_, err = l.insertRaw(rawJavMovie)
	//if err != nil {
	//	return "", err
	//}
	return nil, nil
}
