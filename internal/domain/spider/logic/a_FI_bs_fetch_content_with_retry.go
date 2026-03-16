// internal/spiderx/logic/a_fetch_inventories_with_retry.go
package logic

import (
	"bytes"
	"context"
	"fmt"
	"rudy_gc/pkg/mylog"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const maxRetries = 45

func (l *CrawlLogic) fetchInventoryContentWithRetry(ctx context.Context, fullURL string) (string, error) {
	tryAttempts := 0
	var lastErr error
	var lastStatusCode int
	var lastBody string

	for {
		tryAttempts++
		l.deps.Log.WithContext(ctx).Infof("第 %d 次尝试: %s", tryAttempts, fullURL)

		// 使用已注入的 Fetcher（直连）
		resp, err := l.deps.Fetcher.Get(ctx, fullURL)
		statusCode := 0
		var body []byte
		if resp != nil {
			statusCode = resp.Status
			body = resp.Body
		}

		// 记录本次尝试的结果
		lastErr = err
		lastStatusCode = statusCode
		lastBody = string(body)

		if err == nil && statusCode == 200 {
			content, ferr := filterInventoryContent(body) // 可在此做解码/清洗；当前直接转字符串
			if ferr == nil {
				if isValidContent(content) {
					l.deps.Log.WithContext(ctx).Infof("成功获取内容 - url: %s, attempts: %d, content_length: %d",
						fullURL, tryAttempts, len(content))
					return content, nil
				}
				// 请求成功但页面为空/无结果 → 作为"空白页"处理给上游
				l.deps.Log.WithContext(ctx).Warnf("页面内容无效 - url: %s, attempts: %d, status_code: %d, body_length: %d",
					fullURL, tryAttempts, statusCode, len(string(body)))
				return "", ErrBlankPage
			} else {
				// 过滤内容失败
				l.deps.Log.WithContext(ctx).Warnf("过滤内容失败 - url: %s, attempts: %d, filter_error: %v",
					fullURL, tryAttempts, ferr)
				// 根据你的业务逻辑决定是否返回原始内容
				return string(body), nil
			}
		}

		// 记录请求失败的错误信息
		if err != nil {
			l.deps.Log.WithContext(ctx).Warnf("请求失败 - url: %s, attempts: %d, error: %v, status_code: %d, body_preview: %.200s",
				fullURL, tryAttempts, err, statusCode, truncateBody(string(body)))
		} else if statusCode != 200 {
			l.deps.Log.WithContext(ctx).Warnf("非200状态码 - url: %s, attempts: %d, status_code: %d, body_preview: %.200s",
				fullURL, tryAttempts, statusCode, truncateBody(string(body)))
		}

		// 重试/退避
		if !l.retryHandler(ctx, tryAttempts, err, string(body)) {
			// 到达上限或需要终止
			if lastErr == nil {
				// 无具体 err 时给一个描述
				lastErr = fmt.Errorf("fetch failed after %d attempts, status=%d, body_preview=%s",
					tryAttempts, lastStatusCode, truncateBody(lastBody))
			} else {
				// 丰富错误信息
				lastErr = fmt.Errorf("fetch failed after %d attempts: %v, status=%d, body_preview=%s",
					tryAttempts, lastErr, lastStatusCode, truncateBody(lastBody))
			}

			l.deps.Log.WithContext(ctx).Errorf("达到重试上限 - url: %s, attempts: %d, final_error: %v",
				fullURL, tryAttempts, lastErr)
			return "", lastErr
		}
	}
}

// 辅助函数：截断过长的响应体用于日志
func truncateBody(body string) string {
	if len(body) > 200 {
		return body[:200] + "..."
	}
	return body
}

// 与老代码相同：每 20 次长退避；超过 maxRetries 终止；其余每次 3s。
func (l *CrawlLogic) retryHandler(ctx context.Context, attempts int, err error, response string) bool {
	if attempts%20 == 0 {
		mylog.Warn(ctx, "多次重试失败(第%d次)，10分钟后重试。resp=%.200s", attempts, response)
		time.Sleep(10 * time.Minute)
		return true
	}
	if attempts >= maxRetries {
		mylog.Warn(ctx, "多次重试失败(到达上限%d)，停止。resp=%.200s", maxRetries, response)
		return false
	}
	mylog.Warn(ctx, "GetHttpResponse error: %v（3秒后重试）", err)
	time.Sleep(3 * time.Second)
	return true
}

// 与老代码一致的“有效页面”判定：包含 videothumblist，且不含“搜寻没有结果/空的列表”
func isValidContent(content string) bool {
	return strings.Contains(content, "videothumblist") &&
		!strings.Contains(content, "搜寻没有结果") &&
		!strings.Contains(content, "空的列表")
}

// 与老代码一致：Label 类 inventory 名称附加日期后缀
func buildInventoryName(queryWithPage string, nameType int64, t time.Time) string {
	if nameType == nameTypeLabel {
		// cons.TimeTemplate8 等价：20060102
		return queryWithPage + "date" + t.Format("2006-01")
	}
	return queryWithPage
}

// filterInventoryContent：留接口位置，你要处理编码/去广告可写在这里。
// 现在直接将字节转字符串。
func filterInventoryContent(byts []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(byts))
	if err != nil {
		return "", err
	}
	content, err := doc.Find("div[id=rightcolumn]").Html()
	if err != nil {
		return "", err
	}
	return content, nil
}
