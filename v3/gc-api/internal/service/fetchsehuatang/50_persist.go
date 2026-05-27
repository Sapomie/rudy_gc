package fetchsehuatang

import (
	"context"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

const (
	persistActionInsert = "insert"
	persistActionUpdate = "update"
)

func (s *Service) upsertMagnet(ctx context.Context, row *moviex.TSehuatangMagnet, now int64) (string, error) {
	if row == nil {
		return "", nil
	}
	if err := s.repairSehuatangRowMovieFields(ctx, row); err != nil {
		return "", err
	}
	oldRow, err := s.deps.SehuatangMagnetModel.FindOneByInfoHash(ctx, row.InfoHash)
	if err != nil && err != moviex.ErrNotFound {
		return "", err
	}

	if oldRow == nil {
		row.CreatedOn = now
		row.UpdatedOn = now
		row.LastSeenTime = now
		if row.PostTime <= 0 {
			row.PostTime = now
		}
		if row.PostDate <= 0 {
			row.PostDate = parsePostDate(row.PostTime, time.Unix(now, 0).In(time.Local))
		}
		if _, err = s.deps.SehuatangMagnetModel.Insert(ctx, row); err != nil {
			return "", err
		}
		return persistActionInsert, nil
	}

	oldRow.MovieJavId = row.MovieJavId
	oldRow.MovieName = row.MovieName
	oldRow.Tag = row.Tag
	oldRow.ThreadTitle = row.ThreadTitle
	oldRow.ThreadUrl = row.ThreadUrl
	oldRow.PostTime = row.PostTime
	oldRow.PostDate = row.PostDate
	oldRow.LastSeenTime = now
	oldRow.UpdatedOn = now
	if oldRow.PostTime <= 0 {
		oldRow.PostTime = now
	}
	if oldRow.PostDate <= 0 {
		oldRow.PostDate = parsePostDate(oldRow.PostTime, time.Unix(now, 0).In(time.Local))
	}
	if err = s.deps.SehuatangMagnetModel.Update(ctx, oldRow); err != nil {
		return "", err
	}
	return persistActionUpdate, nil
}
