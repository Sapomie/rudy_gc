package movie

import (
	"testing"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

func TestTorrentAlbumMovieIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		row           *moviex.TmAlbumItem
		magnet        *moviex.ShMagnet
		wantMovieType string
		wantFC2Key    string
	}{
		{
			name:          "JavBus 来源",
			row:           &moviex.TmAlbumItem{SourceType: albumSourceTypeJavbusMagnet, MovieName: "ABP-123"},
			wantMovieType: moviedomain.SehuatangThreadMovieTypeJAV,
		},
		{
			name: "SH FC2 来源",
			row:  &moviex.TmAlbumItem{SourceType: albumSourceTypeSehuatang, MovieName: "FC2PPV-4556413"},
			magnet: &moviex.ShMagnet{
				MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV,
				MovieName: "FC2PPV-4556413",
				FC2Key:    "FC2PPV-4556413-1",
			},
			wantMovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV,
			wantFC2Key:    "FC2PPV-4556413-1",
		},
		{
			name:          "SH 历史 FC2 来源",
			row:           &moviex.TmAlbumItem{SourceType: albumSourceTypeSehuatang, MovieName: "FC2PPV-4556413"},
			wantMovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV,
		},
		{
			name:          "空记录",
			wantMovieType: moviedomain.SehuatangThreadMovieTypeJAV,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			movieType, fc2Key := torrentAlbumMovieIdentity(tt.row, tt.magnet)
			if movieType != tt.wantMovieType || fc2Key != tt.wantFC2Key {
				t.Fatalf("torrentAlbumMovieIdentity() = (%q, %q), want (%q, %q)", movieType, fc2Key, tt.wantMovieType, tt.wantFC2Key)
			}
		})
	}
}
