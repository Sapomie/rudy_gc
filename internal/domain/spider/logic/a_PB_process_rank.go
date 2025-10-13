package logic

import (
	"bytes"
	"fmt"
	consts "rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/ptr"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zeromicro/go-zero/core/logx"
)

func (l *CrawlLogic) ProcessBestinvRank() error {
	log := logx.WithContext(l.ctx)

	// 取出需要排名检查的 bestinv
	ids, err := l.deps.BestinvRepo.ListIDsByRankCheck(l.ctx, consts.BestinvNeedRankCheck, 1000)
	if err != nil {
		return fmt.Errorf("查询需要排名检查的 bestinv 失败: %w", err)
	}
	if len(ids) == 0 {
		log.Info("没有需要排名检查的 bestinv")
		return nil
	}
	log.Infof("需要处理 %d 个 bestinv", len(ids))

	javIdMap := make(map[string]struct{})
	for count, id := range ids {
		bestinv, err := l.deps.BestinvRepo.FindOne(l.ctx, id)
		if err != nil {
			return fmt.Errorf("根据 id=%d 查询 bestinv 失败: %w", id, err)
		}

		if err := l.makeAndInsertRankByBestinv(bestinv, javIdMap); err != nil {
			return fmt.Errorf("处理 bestinv 排名失败(id=%d): %w", id, err)
		}

		if (count+1)%5 == 0 {
			log.Infof("已完成 %d 个 bestinv 的排名处理", count+1)
		}

		// 标记已完成 rank check
		if err := l.deps.BestinvRepo.MarkRankChecked(l.ctx, id, time.Now().Unix()); err != nil {
			return fmt.Errorf("标记 bestinv 已完成排名检查失败(id=%d): %w", id, err)
		}
	}

	// 最后更新电影的排名信息
	if err := l.UpdateMovieRankInfo(javIdMap); err != nil {
		return fmt.Errorf("更新影片排名信息失败: %w", err)
	}

	log.Info("Bestinv 排名处理完成")
	return nil
}

func (l *CrawlLogic) makeAndInsertRankByBestinv(best *types.Bestinv, javIdsMap map[string]struct{}) error {
	// 解析
	rks, err := makeRanksFromBestinv(best)
	if err != nil {
		return fmt.Errorf("解析 Bestinv(id=%d) 失败: %w", best.Id, err)
	}

	// 写库（幂等）
	for _, rk := range rks {
		if err := l.deps.RankRepo.Upsert(l.ctx, rk); err != nil {
			return fmt.Errorf("写入排名失败 rank_key=%s, javId=%s: %w", rk.RankKey, rk.MovieJavId, err)
		}
		// 收集本批涉及的影片
		javIdsMap[rk.MovieJavId] = struct{}{}
	}

	// 标记本条 Bestinv 的“需要排名检查”为已处理
	if err := l.deps.BestinvRepo.MarkRankChecked(l.ctx, best.Id, time.Now().Unix()); err != nil {
		return fmt.Errorf("标记 Bestinv(id=%d) 已完成排名检查失败: %w", best.Id, err)
	}

	return nil
}

// makeRanksFromBestinv 解析 Bestinv.Content，生成当页的排名记录切片
// 规则等同老项目：
// - rank 起始 = (page-1)*20 + 1
// - 过滤标题中包含 “蓝光盘” 标记的条目（consts.MarkBlueRay）
// - 从 <div.toolbar a id="..."> 提取 javId
// - Name = "<date>_<序号3位补零>"
func makeRanksFromBestinv(best *types.Bestinv) ([]*types.Rank, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(best.Content)))
	if err != nil {
		return nil, err
	}

	rks := make([]*types.Rank, 0, 24)
	rankPos := (best.Page-1)*20 + 1

	doc.Find("div.videos div.video").Each(func(_ int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("div.title").Text())
		if strings.Contains(title, consts.MarkBlueRay) {
			return // 跳过蓝光版
		}

		javId, _ := s.Find("div.toolbar a").Attr("id")
		if javId == "" {
			return
		}

		rks = append(rks, &types.Rank{
			RankKey:    types.BuildRankKey(best.Date, rankPos), // 生成唯一键
			MovieJavId: javId,
			DayNumber:  consts.GetRankDayNumber(best.Date),
			RankPos:    rankPos,
			Category:   best.Category,
		})
		rankPos++
	})

	return rks, nil
}

// UpdateMovieRankInfo 批量按 javId 聚合更新排行信息
func (l *CrawlLogic) UpdateMovieRankInfo(javIdSet map[string]struct{}) error {
	l.deps.Log.Infof("开始汇总并更新排行信息，待处理 %d 个 javId", len(javIdSet))

	updated := 0
	for javId := range javIdSet {
		if err := l.AddRankInfo(javId); err != nil {
			return fmt.Errorf("更新 javId=%s 的排行信息失败: %w", javId, err)
		}
		updated++
	}
	l.deps.Log.Infof("排行信息更新完成，共处理 %d 个", updated)
	return nil
}

// AddRankInfo 计算单个影片的：首次上榜日、最佳名次、上榜天数，并写回 bm_minfo
func (l *CrawlLogic) AddRankInfo(javId string) error {
	// 1) 统计聚合（SQL 聚合更快；没有就先在 Repo 做一层）
	firstDay, bestRank, daysInRank, err := l.deps.RankRepo.AggregateByJavId(l.ctx, javId)
	if err != nil {
		return fmt.Errorf("统计排行聚合失败(javId=%s): %w", javId, err)
	}

	// 2) 取得 movie.id（用于回刷关联演员缓存 / 侧写）
	mv, err := l.deps.MovieRepo.FindOneByJavId(l.ctx, javId)
	if err != nil {
		return fmt.Errorf("查询电影失败(javId=%s): %w", javId, err)
	}
	// 3) 取演员 ids（保持与老项目等价的副作用：填充内存 Map）
	castIDs, err := l.deps.MovieCastRepo.ListCastIDsByMovieJavId(l.ctx, mv.JavId)
	if err != nil {
		return fmt.Errorf("查询电影演员关系失败(movieId=%d): %w", mv.Id, err)
	}
	for _, cid := range castIDs {
		l.castIdMap[cid] = struct{}{}
	}

	// 4) 回写 bm_minfo 的排行三件套
	now := time.Now().Unix()

	patch := types.MinfoPatch{
		Chinese:            nil,
		FirstRankDayNumber: ptr.Int64(firstDay),
		HighestRank:        ptr.Int64(bestRank),
		DaysInRank:         ptr.Int64(daysInRank),
		UpdatedOn:          ptr.Int64(now),
	}
	err = l.deps.MinfoRepo.UpdatePartialByJavId(l.ctx, javId, patch)
	if err != nil {
		return fmt.Errorf("写回 bm_minfo 排行信息失败(javId=%s): %w", javId, err)
	}

	l.movieSvc.InvalidateMovieType(l.ctx, javId)

	return nil
}
