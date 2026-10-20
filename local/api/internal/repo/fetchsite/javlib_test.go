package fetchsite

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	fetchsitetypes "rudy_gc/local/api/internal/types/fetchsite"
)

type javlibConfigTestModel struct {
	moviex.TFetchSiteModel
	row     *moviex.TFetchSite
	readErr error
	inserts int
}

func (m *javlibConfigTestModel) FindOneBySiteCode(_ context.Context, code string) (*moviex.TFetchSite, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if m.row == nil || m.row.SiteCode != code {
		return nil, moviex.ErrNotFound
	}
	row := *m.row
	return &row, nil
}

func (m *javlibConfigTestModel) InsertIfMissing(_ context.Context, row *moviex.TFetchSite) (sql.Result, error) {
	m.inserts++
	if m.row == nil {
		copy := *row
		copy.Id = 7
		m.row = &copy
	}
	return nil, nil
}

func TestEnsureJavlibConfigSeedsLegacyValuesAndPreservesSavedConfig(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{Fetcher: config.FetcherConf{
		JavAddress: "www.legacy.test", UserAgent: "legacy-agent", Cookie: "legacy-cookie", Timeout: 60,
	}}
	model := &javlibConfigTestModel{}
	if err := EnsureJavlibConfig(ctx, cfg, model); err != nil {
		t.Fatal(err)
	}
	row := model.row
	if row == nil || row.SiteCode != "javlib" || row.BaseUrl != "https://www.legacy.test" ||
		row.UserAgent != "legacy-agent" || row.Cookie != "legacy-cookie" || row.TimeoutSeconds != 60 ||
		row.CreatedOn == 0 || row.UpdatedOn != row.CreatedOn {
		t.Fatalf("legacy Javlib settings were not preserved: %+v", row)
	}
	row.BaseUrl = "https://saved.test"
	row.Cookie = "saved-cookie"
	// 已入库后允许移除旧域名配置，重启仍使用数据库中的新地址。
	cfg.Fetcher.JavAddress = ""
	if err := EnsureJavlibConfig(ctx, cfg, model); err != nil {
		t.Fatal(err)
	}
	loaded := Load(ctx, cfg, model, nil)[fetchsitetypes.CodeJavlib]
	if model.inserts != 1 || loaded.BaseURL != "https://saved.test" || loaded.Cookie != "saved-cookie" {
		t.Fatalf("saved config was replaced: inserts=%d config=%+v", model.inserts, loaded)
	}
	if loaded := Load(ctx, cfg, model, nil)[fetchsitetypes.CodeSukebei]; loaded.BaseURL != "https://sukebei.nyaa.si" {
		t.Fatalf("unrelated site config changed: %+v", loaded)
	}
}

func TestEnsureJavlibConfigRejectsMissingSeedAndDatabaseFailures(t *testing.T) {
	for _, cfg := range []config.Config{
		{Fetcher: config.FetcherConf{Timeout: 60}},
		{Fetcher: config.FetcherConf{JavAddress: "valid.test"}},
	} {
		model := &javlibConfigTestModel{}
		if err := EnsureJavlibConfig(context.Background(), cfg, model); err == nil || model.inserts != 0 {
			t.Fatalf("invalid seed was inserted: err=%v inserts=%d", err, model.inserts)
		}
	}
	dbErr := errors.New("database unavailable")
	model := &javlibConfigTestModel{readErr: dbErr}
	if err := EnsureJavlibConfig(context.Background(), config.Config{}, model); !errors.Is(err, dbErr) || model.inserts != 0 {
		t.Fatalf("database failure was ignored: err=%v inserts=%d", err, model.inserts)
	}
}
