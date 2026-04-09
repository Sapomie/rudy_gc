package movie

import (
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

func mapAMovieToTypes(mv *moviex.AMovie) *types.Movie {
	if mv == nil {
		return nil
	}
	return &types.Movie{
		Id:                   mv.Id,
		Name:                 mv.Name,
		JavId:                mv.JavId,
		EncodeName:           mv.EncodeName,
		Title:                mv.Title,
		ReleasingDate:        mv.ReleasingDate,
		Length:               mv.Length,
		Score:                mv.Score,
		ViewersNumberWant:    mv.ViewersNumberWant,
		ViewersNumberOwned:   mv.ViewersNumberOwned,
		ViewersNumberWatched: mv.ViewersNumberWatched,
		PrefixId:             mv.PrefixId,
		MakerId:              mv.MakerId,
		LabelId:              mv.LabelId,
		DirectorId:           mv.DirectorId,
		CastNumber:           mv.CastNumber,
		CastAverageAge:       mv.CastAverageAge,
		DetailUpdateTime:     mv.DetailUpdateTime,
		CreatedOn:            mv.CreatedOn,
		UpdatedOn:            mv.UpdatedOn,
	}
}

func mapBmMinfoToTypes(m *moviex.BmMinfo) *types.Minfo {
	if m == nil {
		return nil
	}
	return &types.Minfo{
		Id:                 m.Id,
		JavId:              m.JavId,
		Name:               m.Name,
		Chinese:            m.Chinese,
		FirstRankDayNumber: m.FirstRankDayNumber,
		HighestRank:        m.HighestRank,
		DaysInRank:         m.DaysInRank,
		CreatedOn:          m.CreatedOn,
		UpdatedOn:          m.UpdatedOn,
		ReleasingDate:      m.ReleasingDate,
	}
}

func mapBmMurlToTypes(m *moviex.BmMurl) *types.Murl {
	if m == nil {
		return nil
	}
	return &types.Murl{
		Id:             m.Id,
		JavId:          m.JavId,
		Name:           m.Name,
		JacketImg:      m.JacketImg,
		JacketImgLocal: m.JacketImgLocal,
		CreatedOn:      m.CreatedOn,
		UpdatedOn:      m.UpdatedOn,
	}
}

func mapVFilmToTypes(mv *moviex.LegacyFilm) *types.Film {
	if mv == nil {
		return nil
	}
	return &types.Film{
		Id:            mv.Id,
		MovieJavId:    mv.MovieJavId,
		MovieName:     mv.MovieName,
		FileName:      mv.FileName,
		DirectoryId:   mv.DirectoryId,
		RootDir:       mv.RootDir,
		FullDir:       mv.FullDir,
		Dir1Id:        mv.Dir1Id,
		Dir2Id:        mv.Dir2Id,
		Dir3Id:        mv.Dir3Id,
		Dir4Id:        mv.Dir4Id,
		Alias:         mv.Alias,
		Size:          mv.Size,
		Width:         mv.Width,
		Height:        mv.Height,
		BitRate:       mv.BitRate,
		Duration:      mv.Duration,
		FrameAverage:  mv.FrameAverage,
		HasSub:        mv.HasSub,
		SelfMake:      mv.SelfMake,
		HasMask:       mv.HasMask,
		NeedScanMeta:  mv.NeedScanMeta,
		IsRemoved:     mv.IsRemoved,
		RemoveTime:    mv.RemoveTime,
		ScTimes:       mv.ScTimes,
		ComeTimes:     mv.ComeTimes,
		LastScTime:    mv.LastScTime,
		BirthTime:     mv.BirthTime,
		ReleasingDate: mv.ReleasingDate,
		CreatedOn:     mv.CreatedOn,
		UpdatedOn:     mv.UpdatedOn,
	}
}

func mapWMediaToTypes(mv *moviex.WMedia) *types.Media {
	if mv == nil {
		return nil
	}
	return &types.Media{
		Id:            mv.Id,
		MovieJavId:    mv.MovieJavId,
		MovieName:     mv.MovieName,
		FileName:      mv.FileName,
		DirectoryId:   mv.DirectoryId,
		RootDir:       mv.RootDir,
		FullDir:       mv.FullDir,
		Alias:         mv.Alias,
		Size:          mv.Size,
		Width:         mv.Width,
		Height:        mv.Height,
		BitRate:       mv.BitRate,
		Duration:      mv.Duration,
		FrameAverage:  mv.FrameAverage,
		HasSub:        mv.HasSub,
		SelfMake:      mv.SelfMake,
		HasMask:       mv.HasMask,
		NeedScanMeta:  mv.NeedScanMeta,
		IsRemoved:     mv.IsRemoved,
		RemoveTime:    mv.RemoveTime,
		BirthTime:     mv.BirthTime,
		ReleasingDate: mv.ReleasingDate,
		CreatedOn:     mv.CreatedOn,
		UpdatedOn:     mv.UpdatedOn,
	}
}
