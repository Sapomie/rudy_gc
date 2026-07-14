package fetchsehuatang

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	fc2domain "rudy_gc/internal/domain/fc2"
	moviedomain "rudy_gc/internal/domain/movie"
	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) syncFC2Movie(ctx context.Context, row *moviex.ShSehuatangThread) error {
	if row == nil {
		return nil
	}
	if row.MovieType != moviedomain.SehuatangThreadMovieTypeFC2PPV {
		return nil
	}
	code, ppvNo, ok := fc2domain.NormalizeCode(row.MovieName)
	if !ok {
		return nil
	}
	rowMeta, err := fc2domain.Parse(row.RawContentText)
	if err != nil {
		return fmt.Errorf("解析 FC2 帖子失败 thread_id=%d: %w", row.Id, err)
	}
	movie, err := s.resolveFC2MovieForThread(ctx, row, code, ppvNo, rowMeta)
	if err != nil {
		return err
	}
	rows, err := s.deps.ShSehuatangThreadModel.ListFetchedByFC2MovieID(ctx, movie.ID)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	selected := rows[0]
	selectedMeta, err := fc2domain.Parse(selected.RawContentText)
	if err != nil {
		return fmt.Errorf("解析 FC2 主资料失败 thread_id=%d: %w", selected.Id, err)
	}
	status := "ready"
	for _, source := range rows {
		meta, parseErr := fc2domain.Parse(source.RawContentText)
		if parseErr != nil {
			return fmt.Errorf("解析 FC2 来源失败 thread_id=%d: %w", source.Id, parseErr)
		}
		metaJSON, marshalErr := fc2domain.Marshal(meta)
		if marshalErr != nil {
			return fmt.Errorf("序列化 FC2 来源失败 thread_id=%d: %w", source.Id, marshalErr)
		}
		if updateErr := s.deps.ShSehuatangThreadModel.UpdateFC2MetaJSONByID(ctx, source.Id, metaJSON); updateErr != nil {
			return updateErr
		}
		if strings.TrimSpace(meta.Title) != strings.TrimSpace(selectedMeta.Title) ||
			strings.TrimSpace(meta.PerformerText) != strings.TrimSpace(selectedMeta.PerformerText) ||
			strings.TrimSpace(meta.MaskType) != strings.TrimSpace(selectedMeta.MaskType) {
			status = "conflict"
		}
	}

	now := time.Now().Unix()
	updatedMovie, err := s.deps.FC2MovieModel.Upsert(ctx, &moviex.FC2Movie{
		ID:               movie.ID,
		FC2Key:           movie.FC2Key,
		Code:             code,
		CodeSeq:          movie.CodeSeq,
		PPVNo:            ppvNo,
		Title:            selectedMeta.Title,
		PerformerText:    selectedMeta.PerformerText,
		MaskType:         selectedMeta.MaskType,
		MetadataThreadID: selected.Id,
		PostTime:         selected.PostTime,
		DataStatus:       status,
		CreatedOn:        now,
		UpdatedOn:        now,
	})
	if err != nil {
		return err
	}
	return s.syncFC2ScStatMetadata(ctx, updatedMovie)
}

func (s *Service) resolveFC2MovieForThread(ctx context.Context, row *moviex.ShSehuatangThread, code string, ppvNo int64, meta fc2domain.Meta) (*moviex.FC2Movie, error) {
	if row.FC2MovieID > 0 {
		movie, err := s.deps.FC2MovieModel.FindOne(ctx, row.FC2MovieID)
		if err != nil {
			return nil, fmt.Errorf("查询 FC2 帖子归属失败 thread_id=%d: %w", row.Id, err)
		}
		if strings.TrimSpace(movie.Code) != code {
			return nil, fmt.Errorf("FC2 帖子归属番号不一致 thread_id=%d movie_id=%d", row.Id, movie.ID)
		}
		return movie, nil
	}

	movies, err := s.deps.FC2MovieModel.ListByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	for _, movie := range movies {
		if movie != nil && fc2MovieMetaMatches(movie, meta) {
			if err := s.deps.ShSehuatangThreadModel.UpdateFC2MovieIDByID(ctx, row.Id, movie.ID); err != nil {
				return nil, err
			}
			row.FC2MovieID = movie.ID
			return movie, nil
		}
	}

	codeSeq := int64(1)
	for _, movie := range movies {
		if movie != nil && movie.CodeSeq >= codeSeq {
			codeSeq = movie.CodeSeq + 1
		}
	}
	now := time.Now().Unix()
	movie, err := s.deps.FC2MovieModel.Upsert(ctx, &moviex.FC2Movie{
		FC2Key:           fc2domain.BuildKey(code, codeSeq),
		Code:             code,
		CodeSeq:          codeSeq,
		PPVNo:            ppvNo,
		Title:            meta.Title,
		PerformerText:    meta.PerformerText,
		MaskType:         meta.MaskType,
		MetadataThreadID: row.Id,
		PostTime:         row.PostTime,
		DataStatus:       "ready",
		CreatedOn:        now,
		UpdatedOn:        now,
	})
	if err != nil {
		return nil, err
	}
	if err := s.syncFC2ScStatMetadata(ctx, movie); err != nil {
		return nil, err
	}
	if err := s.deps.ShSehuatangThreadModel.UpdateFC2MovieIDByID(ctx, row.Id, movie.ID); err != nil {
		return nil, err
	}
	row.FC2MovieID = movie.ID
	return movie, nil
}

func (s *Service) syncFC2ScStatMetadata(ctx context.Context, movie *moviex.FC2Movie) error {
	if movie == nil || movie.ID <= 0 {
		return nil
	}
	stat, err := s.deps.FC2ScStatModel.FindOneByMovieID(ctx, movie.ID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return err
	}
	stat.FC2Key = strings.TrimSpace(movie.FC2Key)
	stat.MovieName = strings.TrimSpace(movie.Title)
	if stat.MovieName == "" {
		stat.MovieName = stat.FC2Key
	}
	stat.PostTime = movie.PostTime
	stat.UpdatedOn = time.Now().Unix()
	return s.deps.FC2ScStatModel.Update(ctx, stat)
}

func fc2MovieMetaMatches(movie *moviex.FC2Movie, meta fc2domain.Meta) bool {
	return strings.TrimSpace(movie.Title) == strings.TrimSpace(meta.Title) &&
		strings.TrimSpace(movie.PerformerText) == strings.TrimSpace(meta.PerformerText) &&
		strings.TrimSpace(movie.MaskType) == strings.TrimSpace(meta.MaskType)
}
