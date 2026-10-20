package movie

import (
	"reflect"
	"testing"

	"rudy_gc/local/api/internal/model/modelx/moviex"
)

func TestBuildMovieDetailHrefsTracksCurrentMovieExistence(t *testing.T) {
	refs := []MovieDetailLinkRef{
		{MovieJavID: "jav-existing", MovieName: "OLD-NAME"},
		{MovieName: "NAME-EXISTS"},
		{MovieName: "NAME-MISSING"},
	}

	if got := buildMovieDetailHrefs(refs, nil); !reflect.DeepEqual(got, []string{"", "", ""}) {
		t.Fatalf("影片尚未入库时链接 = %#v", got)
	}

	rows := []*moviex.AMovieIdentity{
		{JavID: "jav-existing", Name: "CANONICAL-001"},
		{JavID: "jav-by-name", Name: "NAME-EXISTS"},
	}
	want := []string{"/moviedetail/CANONICAL-001", "/moviedetail/NAME-EXISTS", ""}
	if got := buildMovieDetailHrefs(refs, rows); !reflect.DeepEqual(got, want) {
		t.Fatalf("影片入库后链接 = %#v, want %#v", got, want)
	}
}

func TestBuildMovieDetailHrefsEscapesMovieName(t *testing.T) {
	refs := []MovieDetailLinkRef{{MovieName: "ABC 123"}}
	rows := []*moviex.AMovieIdentity{{JavID: "jav-space", Name: "ABC 123"}}
	want := []string{"/moviedetail/ABC%20123"}
	if got := buildMovieDetailHrefs(refs, rows); !reflect.DeepEqual(got, want) {
		t.Fatalf("带空格影片链接 = %#v, want %#v", got, want)
	}
}
