package filmmove

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func (s *Service) Commit(ctx context.Context, planID string) (*CommitResult, error) {
	plan := s.takePlan(planID)
	if plan == nil {
		return nil, fmt.Errorf("plan not found or expired")
	}

	result := &CommitResult{
		PlanID:       plan.ID,
		Total:        len(plan.Items),
		Items:        make([]*CommitItem, 0, len(plan.Items)),
		SuccessItems: make([]*CommitItem, 0, len(plan.Items)),
		FailedItems:  make([]*CommitItem, 0, len(plan.Items)),
	}

	for _, planItem := range plan.Items {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if planItem == nil {
			continue
		}

		item := &CommitItem{
			MovieName:  strings.TrimSpace(planItem.MovieName),
			MovieJavID: strings.TrimSpace(planItem.MovieJavID),
			SourcePath: strings.TrimSpace(planItem.SourcePath),
			TargetPath: strings.TrimSpace(planItem.TargetPath),
			Status:     "fail",
			Error:      strings.TrimSpace(planItem.Error),
		}

		if !planItem.CanMove {
			if item.Error == "" {
				item.Error = "预处理校验未通过"
			}
			result.Failed++
			result.Items = append(result.Items, item)
			result.FailedItems = append(result.FailedItems, item)
			continue
		}

		if item.SourcePath == "" || item.TargetPath == "" {
			item.Error = "源路径或目标路径为空"
			result.Failed++
			result.Items = append(result.Items, item)
			result.FailedItems = append(result.FailedItems, item)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(item.TargetPath), 0o755); err != nil {
			item.Error = "创建目标目录失败: " + err.Error()
			result.Failed++
			result.Items = append(result.Items, item)
			result.FailedItems = append(result.FailedItems, item)
			continue
		}

		if _, err := os.Stat(item.SourcePath); err != nil {
			item.Error = "源文件不存在或不可读"
			result.Failed++
			result.Items = append(result.Items, item)
			result.FailedItems = append(result.FailedItems, item)
			continue
		}

		if _, err := os.Stat(item.TargetPath); err == nil {
			item.Error = "目标文件已存在"
			result.Failed++
			result.Items = append(result.Items, item)
			result.FailedItems = append(result.FailedItems, item)
			continue
		} else if !os.IsNotExist(err) {
			item.Error = "检查目标文件失败: " + err.Error()
			result.Failed++
			result.Items = append(result.Items, item)
			result.FailedItems = append(result.FailedItems, item)
			continue
		}

		if err := os.Rename(item.SourcePath, item.TargetPath); err != nil {
			if errors.Is(err, syscall.EXDEV) {
				item.Error = "跨卷移动被禁止: " + item.SourcePath + " => " + item.TargetPath
			} else {
				item.Error = "移动失败: " + err.Error()
			}
			result.Failed++
			result.Items = append(result.Items, item)
			result.FailedItems = append(result.FailedItems, item)
			continue
		}

		item.Status = "success"
		item.Error = ""
		result.Success++
		result.Items = append(result.Items, item)
		result.SuccessItems = append(result.SuccessItems, item)
	}

	return result, nil
}
