package sc

import (
	"testing"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	mediaservice "rudy_gc/local/api/internal/service/media"
	"rudy_gc/local/api/internal/service/sc"
	"rudy_gc/local/api/internal/types"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

func TestValidateScSmartPickReq(t *testing.T) {
	tests := []struct {
		name    string
		req     scSmartPickReq
		wantErr bool
	}{
		{name: "both zero", req: scSmartPickReq{}, wantErr: true},
		{name: "negative JAV", req: scSmartPickReq{JAV: scSmartPickJAVReq{PickN: -1}, FC2: scSmartPickFC2Req{PickN: 1, Reqs: []scPickReq{{Weight: 1}}}}, wantErr: true},
		{name: "negative FC2", req: scSmartPickReq{JAV: scSmartPickJAVReq{PickN: 1, Reqs: []scPickReq{{Weight: 1}}}, FC2: scSmartPickFC2Req{PickN: -1}}, wantErr: true},
		{name: "JAV reqs missing", req: scSmartPickReq{JAV: scSmartPickJAVReq{PickN: 1}}, wantErr: true},
		{name: "FC2 reqs missing", req: scSmartPickReq{FC2: scSmartPickFC2Req{PickN: 1}}, wantErr: true},
		{name: "JAV only", req: scSmartPickReq{JAV: scSmartPickJAVReq{PickN: 2, Reqs: []scPickReq{{Weight: 1}}}}},
		{name: "FC2 only", req: scSmartPickReq{FC2: scSmartPickFC2Req{PickN: 3, Reqs: []scPickReq{{Weight: 1}}}}},
		{name: "mixed", req: scSmartPickReq{
			JAV: scSmartPickJAVReq{PickN: 2, Reqs: []scPickReq{{Weight: 1}}},
			FC2: scSmartPickFC2Req{PickN: 3, Reqs: []scPickReq{{Weight: 1}}},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateScSmartPickReq(test.req)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateScSmartPickReq() error = %v, wantErr = %v", err, test.wantErr)
			}
		})
	}
}

func TestFormatSmartPickSizeGB(t *testing.T) {
	if got := formatSmartPickSizeGB(3 * 1024 * 1024 * 1024); got != "3.00" {
		t.Fatalf("formatSmartPickSizeGB() = %q", got)
	}
}

func TestBuildScAddPreviewMovieResponsesUsesSharedCardData(t *testing.T) {
	javCard := &types.MovieType{Name: "ABP-123", Title: "JAV 标题"}
	fc2Movie := &moviex.FC2Movie{ID: 42, FC2Key: "FC2PPV-4556413-1", Title: "FC2 标题"}
	fc2Card := &mediaservice.FC2CardView{FC2Movie: fc2Movie, CoverImageURL: "/covers/fc2.png"}

	responses := buildScAddPreviewMovieResponses([]*sc.AddScPreviewMovie{
		{MovieType: moviedomain.SehuatangThreadMovieTypeJAV, MovieJavId: "ABP-123", JAVCard: javCard},
		{MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV, FC2MovieId: 42, FC2Movie: fc2Movie},
	}, []*mediaservice.FC2CardView{fc2Card})

	if len(responses) != 2 {
		t.Fatalf("len(responses) = %d", len(responses))
	}
	if responses[0].Card != javCard {
		t.Fatal("JAV preview did not use MovieCard data")
	}
	if responses[1].Card != fc2Card {
		t.Fatal("FC2 preview did not use FC2MovieCard data")
	}
}

func TestApplyScEventMovieCardsUsesCanonicalCardData(t *testing.T) {
	javCard := &types.MovieType{Name: "ABP-123"}
	fc2Movie := &moviex.FC2Movie{ID: 42, FC2Key: "FC2PPV-4556413-1"}
	fc2Card := &mediaservice.FC2CardView{FC2Movie: fc2Movie, CoverImageURL: "/covers/fc2.png"}
	items := []*types.ScEventMovie{
		{Entry: &types.GList{MovieType: moviedomain.SehuatangThreadMovieTypeJAV}, MovieType: javCard},
		{Entry: &types.GList{MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV, FC2MovieId: 42}},
	}

	applyScEventMovieCards(items, []*mediaservice.FC2CardView{fc2Card})

	if items[0].Card != javCard {
		t.Fatal("JAV event movie did not use MovieCard data")
	}
	if items[1].Card != fc2Card {
		t.Fatal("FC2 event movie did not use FC2MovieCard data")
	}
}
