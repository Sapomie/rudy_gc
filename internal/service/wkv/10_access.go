package wkv

import (
	"context"
	"errors"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) GetValue(ctx context.Context, itemKey string) (string, error) {
	itemKey = strings.TrimSpace(itemKey)
	if itemKey == "" {
		return "", nil
	}

	row, err := s.deps.WKvModel.FindOneByItemKey(ctx, itemKey)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) || isWKvTableMissing(err) {
			return "", nil
		}
		return "", err
	}
	if row == nil {
		return "", nil
	}
	return strings.TrimSpace(row.ItemValue), nil
}

func (s *Service) UpsertDate(ctx context.Context, itemKey string, itemValue string) (string, error) {
	itemKey = strings.TrimSpace(itemKey)
	itemValue = strings.TrimSpace(itemValue)
	if itemKey == "" {
		return "", errors.New("item_key 不能为空")
	}
	if itemValue == "" {
		return "", errors.New("日期不能为空")
	}

	t, err := time.ParseInLocation("2006-01-02", itemValue, time.Local)
	if err != nil {
		return "", errors.New("日期格式必须为 YYYY-MM-DD")
	}
	value := t.Format("2006-01-02")
	now := time.Now().Unix()

	row, err := s.deps.WKvModel.FindOneByItemKey(ctx, itemKey)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return "", err
	}
	if err == nil && row != nil {
		if strings.TrimSpace(row.ItemValue) == value {
			return value, nil
		}
		row.ItemValue = value
		row.UpdatedTime = now
		if err := s.deps.WKvModel.Update(ctx, row); err != nil {
			return "", err
		}
		return value, nil
	}

	_, err = s.deps.WKvModel.Insert(ctx, &moviex.WKv{
		ItemKey:     itemKey,
		ItemValue:   value,
		CreatedTime: now,
		UpdatedTime: now,
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

func isWKvTableMissing(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "error 1146") && strings.Contains(msg, "w_kv")
}
