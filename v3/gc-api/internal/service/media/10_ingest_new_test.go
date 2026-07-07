package media

import (
	"testing"

	"rudy_gc/internal/consts"
)

func TestMergeIngestRootResultNilRootResult(t *testing.T) {
	result := &IngestNewResult{}

	mergeIngestRootResult(result, nil)

	if result.Precheck.Total != 0 || result.Total != 0 || result.Success != 0 || result.Failed != 0 {
		t.Fatalf("nil root result changed aggregate: %+v", result)
	}
}

func TestMergeIngestRootResultCounts(t *testing.T) {
	result := &IngestNewResult{}
	rootResult := &ingestRootResult{
		Precheck: IngestPrecheckStat{
			Total:  2,
			Passed: 1,
			Failed: 1,
		},
		Items: []*IngestFileItem{
			{Status: ingestItemStatusPass},
			{Status: ingestItemStatusFail, Error: "failed"},
		},
	}

	mergeIngestRootResult(result, rootResult)

	if result.Precheck.Total != 2 || result.Precheck.Passed != 1 || result.Precheck.Failed != 1 {
		t.Fatalf("unexpected precheck aggregate: %+v", result.Precheck)
	}
	if result.Total != 2 || result.Success != 1 || result.Failed != 1 || len(result.Items) != 2 {
		t.Fatalf("unexpected item aggregate: total=%d success=%d failed=%d items=%d", result.Total, result.Success, result.Failed, len(result.Items))
	}
}

func TestParseRawMovieMetaSupportsSerialLetterSuffix(t *testing.T) {
	meta, err := parseRawMovieMeta("OLM-044N~1.mp4")
	if err != nil {
		t.Fatalf("parseRawMovieMeta failed: %v", err)
	}
	if meta.movieName != "OLM-044N" {
		t.Fatalf("unexpected movie name: %s", meta.movieName)
	}
	if meta.ext != ".mp4" {
		t.Fatalf("unexpected extension: %s", meta.ext)
	}
	if meta.selfMake != consts.FilmSelfMake {
		t.Fatalf("unexpected selfMake: %d", meta.selfMake)
	}
}
