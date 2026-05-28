package media

import "testing"

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
