package dailybest

import (
	"context"
	"errors"
	"fmt"
	"strings"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

var ErrInvalidSourceConfig = errors.New("站点配置无效")

type SourceConfigRepository interface {
	FindSourceConfig(context.Context) (dailybesttypes.SourceConfig, error)
	InsertSourceConfigIfMissing(context.Context, string) error
	SaveSourceConfig(context.Context, string) (dailybesttypes.SourceConfig, error)
}

func (s *Service) sourceConfigRepository() (SourceConfigRepository, error) {
	repository, ok := s.repository.(SourceConfigRepository)
	if !ok {
		return nil, fmt.Errorf("DailyBest 站点配置仓库为空")
	}
	return repository, nil
}

// PrepareSourceConfig 仅在配置缺失时导入旧 YAML，重启不会覆盖已保存的域名。
func (s *Service) PrepareSourceConfig(ctx context.Context, legacyAddress string) error {
	repository, err := s.sourceConfigRepository()
	if err != nil {
		return err
	}
	_, err = repository.FindSourceConfig(ctx)
	if err == nil {
		return nil
	}
	if !errors.Is(err, ErrNotFound) {
		return err
	}
	if strings.TrimSpace(legacyAddress) == "" {
		return nil
	}
	address, err := normalizeJavAddress(legacyAddress)
	if err != nil {
		return err
	}
	return repository.InsertSourceConfigIfMissing(ctx, address)
}

func (s *Service) GetSourceConfig(ctx context.Context) (dailybesttypes.SourceConfig, error) {
	repository, err := s.sourceConfigRepository()
	if err != nil {
		return dailybesttypes.SourceConfig{}, err
	}
	config, err := repository.FindSourceConfig(ctx)
	if errors.Is(err, ErrNotFound) {
		return dailybesttypes.SourceConfig{}, nil
	}
	if err != nil {
		return dailybesttypes.SourceConfig{}, err
	}
	config.TargetProbeURL = dailybesttypes.SourcePageURL(config.JavAddress, 1)
	return config, nil
}

func (s *Service) UpdateSourceConfig(ctx context.Context, input dailybesttypes.UpdateSourceConfigRequest) (dailybesttypes.SourceConfig, error) {
	address, err := normalizeJavAddress(input.JavAddress)
	if err != nil {
		return dailybesttypes.SourceConfig{}, err
	}
	repository, err := s.sourceConfigRepository()
	if err != nil {
		return dailybesttypes.SourceConfig{}, err
	}
	config, err := repository.SaveSourceConfig(ctx, address)
	if err != nil {
		return dailybesttypes.SourceConfig{}, err
	}
	config.TargetProbeURL = dailybesttypes.SourcePageURL(config.JavAddress, 1)
	return config, nil
}

func normalizeJavAddress(raw string) (string, error) {
	address := strings.ToLower(strings.TrimSpace(raw))
	if address == "" {
		return "", fmt.Errorf("%w: JAV 域名不能为空", ErrInvalidSourceConfig)
	}
	labels := strings.Split(address, ".")
	if len(labels) < 2 {
		return "", fmt.Errorf("%w: 请填写完整域名，省略协议、端口和路径", ErrInvalidSourceConfig)
	}
	for _, label := range labels {
		if label == "" || label[0] == '-' || label[len(label)-1] == '-' {
			return "", fmt.Errorf("%w: JAV 域名格式无效", ErrInvalidSourceConfig)
		}
		for _, char := range label {
			if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '-') {
				return "", fmt.Errorf("%w: 请填写域名，省略协议、端口和路径", ErrInvalidSourceConfig)
			}
		}
	}
	return address, nil
}
