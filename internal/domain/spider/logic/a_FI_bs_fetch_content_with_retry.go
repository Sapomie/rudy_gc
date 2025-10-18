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

		if err == nil && statusCode == 200 {
			content, ferr := filterInventoryContent(body) // 可在此做解码/清洗；当前直接转字符串
			if ferr == nil {
				if isValidContent(content) {
					return content, nil
				}
				// 请求成功但页面为空/无结果 → 作为“空白页”处理给上游
				return "", ErrBlankPage
			}
		}

		// 重试/退避
		if !l.retryHandler(ctx, tryAttempts, err, string(body)) {
			// 到达上限或需要终止
			if err == nil {
				// 无具体 err 时给一个描述
				err = fmt.Errorf("fetch failed after %d attempts, status=%d", tryAttempts, statusCode)
			}
			return "", err
		}
	}
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
