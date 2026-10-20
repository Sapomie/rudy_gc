package movie

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/dep"
)

func TestMovieCardBatchMatchesSingleLoadIntegration(t *testing.T) {
	configPath := strings.TrimSpace(os.Getenv("RUDY_GC_INTEGRATION_CONFIG"))
	if configPath == "" {
		t.Skip("设置 RUDY_GC_INTEGRATION_CONFIG 后执行真实数据库卡片批量装配测试")
	}

	cfg := config.MustLoad(configPath)
	deps, err := dep.BuildDep(cfg)
	if err != nil {
		t.Fatalf("构建依赖失败: %v", err)
	}
	if deps.MovieCardBatchRepo == nil {
		t.Fatal("MovieCardBatchRepo 未初始化")
	}

	var javIDs []string
	query := `
SELECT am.jav_id
FROM a_movie am
LEFT JOIN w_media wm
  ON wm.movie_jav_id = am.jav_id
 AND wm.source_type = 2
LEFT JOIN g_sc_stat gss
  ON gss.movie_jav_id = am.jav_id
ORDER BY
  CASE WHEN wm.movie_jav_id IS NULL THEN 1 ELSE 0 END,
  CASE WHEN gss.movie_jav_id IS NULL THEN 1 ELSE 0 END,
  am.id DESC
LIMIT 5`
	if err := deps.SqlConn.QueryRowsCtx(context.Background(), &javIDs, query); err != nil {
		t.Fatalf("读取样本影片失败: %v", err)
	}
	if len(javIDs) == 0 {
		t.Fatal("缺少可用于批量装配对照的影片")
	}

	svc := NewService(deps)
	data, err := deps.MovieCardBatchRepo.Load(context.Background(), javIDs)
	if err != nil {
		t.Fatalf("批量读取失败: %v", err)
	}
	for _, javID := range javIDs {
		want, err := svc.buildMovieTypeFromModels(context.Background(), javID)
		if err != nil {
			t.Fatalf("逐片装配 %s 失败: %v", javID, err)
		}
		got, err := svc.buildMovieTypeFromBatch(javID, data)
		if err != nil {
			t.Fatalf("批量装配 %s 失败: %v", javID, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("影片 %s 的批量装配结果与逐片装配结果不一致\ngot: %#v\nwant: %#v", javID, got, want)
		}
	}
}
