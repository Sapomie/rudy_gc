package logic

import (
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
}

type TransResult struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

const (
	appID  = "20230502001663707"
	salt   = "1435660288"
	apiKey = "_ykBXQmfTGkJXqTqTrE1"
)

var errorTranslateSensitiveWord = errors.New("translate sensitive word")

func (l *CrawlLogic) TranslateTitle() error {
	var count int64
	items, err := l.deps.ItemRepo.FindByTranslateStatus(l.ctx, types.ItemChineseNone)
	if err != nil {
		return err
	}

	for _, item := range items {
		movie, err := l.deps.MovieRepo.FindOneByJavId(l.ctx, item.JavId)
		if err != nil {
			return errors.New("MovieRepo FindOneByJavId err:" + err.Error() + item.Name)
		}

		src := strings.TrimPrefix(movie.Title, movie.Name)
		dst, err := l.GetChineseNameFromBaidu(src)
		if err != nil {
			if errors.Is(err, errorTranslateSensitiveWord) {
				l.deps.Log.Warnf("Sensitive word detected for movie: Name=%s, Title=%s", movie.Name, movie.Title)
			} else {
				l.deps.Log.Warnf("Translation error for movie %s: %v. Retrying...", movie.Name, err)
				time.Sleep(10 * time.Second)
				continue
			}
		}

		now := time.Now().Unix()
		var translationStatus int64
		if dst != "" {
			newChinese := replaceSlashes(dst)
			patch := types.MinfoPatch{
				Chinese: &newChinese, // 或者 nil
			}
			if err := l.deps.MinfoRepo.UpdatePartialByJavId(l.ctx, movie.JavId, patch); err != nil {
				return err
			}
			translationStatus = types.ItemChineseOK
		} else {
			translationStatus = types.ItemChineseError
		}

		err = l.deps.ItemRepo.UpdatePartialByJavId(l.ctx, item.JavId, types.ItemPatch{
			HasChinese: ptr.Int64(translationStatus),
			UpdatedOn:  &now,
		})

		l.movieSvc.InvalidateMovieType(l.ctx, movie.JavId)

		count++
		l.deps.Log.Infof("%s  Title添加中文翻译， 完成 %d/%d movies", movie.Name, count, len(items))
		time.Sleep(120 * time.Millisecond)

	}

	l.deps.Log.Infof("完成%v 个Movie 的翻译", count)
	return nil
}

func replaceSlashes(input string) string {
	return strings.ReplaceAll(input, "/", "_")
}

func (l *CrawlLogic) GetChineseNameFromBaidu(src string) (string, error) {
	sign := Md5Hash(appID + src + salt + apiKey)
	queryURL := fmt.Sprintf(
		"https://api.fanyi.baidu.com/api/trans/vip/translate?q=%s&from=jp&to=zh&appid=%s&salt=%s&sign=%s",
		url.QueryEscape(src),
		appID,
		salt,
		sign,
	)

	r, err := l.deps.Fetcher.GetWithProxy(l.ctx, queryURL)
	if err != nil {
		return "", err
	}
	responseData := r.Body

	if strings.Contains(string(responseData), "Hit sensitive word") {
		return "", errorTranslateSensitiveWord
	}

	var resp BaiduTransResp
	if err := json.Unmarshal(responseData, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(resp.TransResult) != 1 {
		return "", fmt.Errorf("unexpected number of translation results: got %d", len(resp.TransResult))
	}

	return resp.TransResult[0].Dst, nil
}

func Md5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
