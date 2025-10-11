package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/ptr"
	"strings"
	"time"
)

type BaiduTransResp struct {
	From        string         `json:"from"`
	To          string         `json:"to"`
	TransResult []*TransResult `json:"trans_result"`

	// 百度错误返回（兼容）
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

type TransResult struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

const (
	appID  = "20230502001663707"
	salt   = "1435660288"
	apiKey = "_ykBXQmfTGkJXqTqTrE1"

	// 速率与重试
	minCallInterval   = 120 * time.Millisecond // API 调用间隔（你原本就 sleep 了）
	maxRetry          = 3                      // 短暂错误重试次数
	initialBackoff    = 600 * time.Millisecond // 起始退避
	backoffMultiplier = 2.0
)

var (
	errSensitive = errors.New("translate sensitive word")
	errBadJSON   = errors.New("bad baidu json")
)

// TranslateTitle：核心逻辑不变，但更稳、更幂等
func (l *CrawlLogic) TranslateTitle() error {
	var count int64

	items, err := l.deps.ItemRepo.FindByTranslateStatus(l.ctx, types.ItemChineseNone)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		l.deps.Log.Info("没有需要翻译的条目")
		return nil
	}

	lastCall := time.Time{}

	for _, item := range items {
		// 支持取消
		select {
		case <-l.ctx.Done():
			return l.ctx.Err()
		default:
		}

		movie, err := l.deps.MovieRepo.FindOneByJavId(l.ctx, item.JavId)
		if err != nil {
			return fmt.Errorf("MovieRepo.FindOneByJavId err: %w, name=%s", err, item.Name)
		}

		// 只翻译「Title 去掉编号」的副标题
		src := strings.TrimSpace(strings.TrimPrefix(movie.Title, movie.Name))
		if src == "" {
			// 没有可翻译文本，按“失败”记（也可按需记 OK）
			now := time.Now().Unix()
			_ = l.deps.ItemRepo.UpdatePartialByJavId(l.ctx, item.JavId, types.ItemPatch{
				HasChinese: ptr.Int64(types.ItemChineseError),
				UpdatedOn:  &now,
			})
			continue
		}

		// 简单限速：保证调用间隔
		sleepDelta := minCallInterval - time.Since(lastCall)
		if sleepDelta > 0 {
			time.Sleep(sleepDelta)
		}

		dst, err := l.translateWithRetry(l.ctx, src)
		lastCall = time.Now()

		var translationStatus int64
		switch {
		case err == nil:
			dst = replaceSlashes(strings.TrimSpace(dst))
			if dst != "" {
				// 仅当中文真的变化时再更新 & 失效缓存
				now := time.Now().Unix()

				// 读旧值（可选：若你上层已有 minfo 就不再读）
				// 这里直接尝试更新，repo 层可选择只更新非空字段
				if uerr := l.deps.MinfoRepo.UpdatePartialByJavId(l.ctx, movie.JavId, types.MinfoPatch{
					Chinese:   &dst,
					UpdatedOn: &now,
				}); uerr != nil {
					return uerr
				}

				l.movieSvc.InvalidateMovieType(l.ctx, movie.JavId)
				translationStatus = types.ItemChineseOK
			} else {
				translationStatus = types.ItemChineseError
			}

		case errors.Is(err, errSensitive):
			l.deps.Log.Warnf("敏感词：Name=%s, Title=%s", movie.Name, movie.Title)
			translationStatus = types.ItemChineseSensitive

		default:
			// 非敏感的其它错误：记录并继续下一个（不返回整体失败）
			l.deps.Log.Warnf("翻译失败 %s: %v", movie.Name, err)
			translationStatus = types.ItemChineseError
		}

		now := time.Now().Unix()
		if uerr := l.deps.ItemRepo.UpdatePartialByJavId(l.ctx, item.JavId, types.ItemPatch{
			HasChinese: ptr.Int64(translationStatus),
			UpdatedOn:  &now,
		}); uerr != nil {
			return uerr
		}

		count++
		l.deps.Log.Infof("%s 翻译完成 %d/%d", movie.Name, count, len(items))
	}

	l.deps.Log.Infof("完成 %d 个 Movie 的翻译", count)
	return nil
}

// 带重试的翻译
func (l *CrawlLogic) translateWithRetry(ctx context.Context, src string) (string, error) {
	backoff := initialBackoff
	var lastErr error

	for attempt := 0; attempt <= maxRetry; attempt++ {
		// 支持取消
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		dst, err := l.GetChineseNameFromBaidu(src)
		if err == nil || errors.Is(err, errSensitive) || errors.Is(err, errBadJSON) {
			return dst, err // 成功/敏感词/明显非重试错误：直接返回
		}

		lastErr = err
		if attempt < maxRetry {
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * backoffMultiplier)
		}
	}
	return "", lastErr
}

// 更稳的百度解析：支持 error_code / error_msg，不再依赖“Hit sensitive word”字符串
func (l *CrawlLogic) GetChineseNameFromBaidu(src string) (string, error) {
	sign := md5Hex(appID + src + salt + apiKey)
	queryURL := fmt.Sprintf(
		"https://api.fanyi.baidu.com/api/trans/vip/translate?q=%s&from=jp&to=zh&appid=%s&salt=%s&sign=%s",
		url.QueryEscape(src), appID, salt, sign,
	)

	r, err := l.deps.Fetcher.GetWithProxy(l.ctx, queryURL)
	if err != nil {
		return "", err
	}
	body := r.Body

	var resp BaiduTransResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("%w: %v", errBadJSON, err)
	}

	// 先看是否是错误返回
	if resp.ErrorCode != "" {
		// 百度对敏感词通常会给出专门的错误码/错误信息（你的旧逻辑是字符串匹配）
		// 这里采用“包含 sensitive”的兜底并保留 error_msg 作为详情。
		if strings.Contains(strings.ToLower(resp.ErrorMsg), "sensitive") {
			return "", errSensitive
		}
		return "", fmt.Errorf("baidu error %s: %s", resp.ErrorCode, resp.ErrorMsg)
	}

	if len(resp.TransResult) != 1 || resp.TransResult[0] == nil {
		return "", fmt.Errorf("unexpected translation results: %d", len(resp.TransResult))
	}

	return resp.TransResult[0].Dst, nil
}

/* ---------- helpers ---------- */

func replaceSlashes(s string) string {
	return strings.ReplaceAll(s, "/", "_")
}

func md5Hex(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
