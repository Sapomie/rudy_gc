// internal/domain/movie/movie_type_build.go
package movie

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/url"
	"path/filepath"
	"rudy_gc/internal/consts"
	"strings"
	"time"

	"rudy_gc/internal/types"
)

// buildMovieTypeFromRepos 聚合多个表，生成用于前端展示的 MovieType
func (s *Service) buildMovieTypeFromRepos(ctx context.Context, javId string) (*types.MovieType, error) {
	// 0) 基础 Movie
	mv, err := s.deps.MovieRepo.FindOneByJavId(ctx, javId)
	if err != nil {
		return nil, fmt.Errorf("FindOneByJavId failed: %w", err)
	}

	// 1) 扩展信息：图片/视频/编码
	murl, err := s.deps.MurlRepo.FindOneByJavId(ctx, mv.JavId)
	if err != nil {
		return nil, fmt.Errorf("MurlRepo.FindOneByJavId failed: %w", err)
	}

	// 2) 关联信息：演员 & 类型
	castInfos, err := s.getCastInfos(ctx, mv.Id, mv.ReleasingDate)
	if err != nil {
		return nil, fmt.Errorf("getCastInfos failed: %w", err)
	}
	genreNames, err := s.getGenreNames(ctx, mv.Id)
	if err != nil {
		return nil, fmt.Errorf("getGenreNames failed: %w", err)
	}

	// 3) 字典信息：导演/前缀/厂牌/标签
	director, err := s.deps.DirectorRepo.FindOne(ctx, mv.DirectorId)
	if err != nil {
		return nil, fmt.Errorf("DirectorRepo.FindOne(%d) failed: %w", mv.DirectorId, err)
	}
	prefix, err := s.deps.PrefixRepo.FindOne(ctx, mv.PrefixId)
	if err != nil {
		return nil, fmt.Errorf("PrefixRepo.FindOne(%d) failed: %w", mv.PrefixId, err)
	}
	label, err := s.deps.LabelRepo.FindOne(ctx, mv.LabelId)
	if err != nil {
		return nil, fmt.Errorf("LabelRepo.FindOne(%d) failed: %w", mv.LabelId, err)
	}
	maker, err := s.deps.MakerRepo.FindOne(ctx, mv.MakerId)
	if err != nil {
		return nil, fmt.Errorf("MakerRepo.FindOne(%d) failed: %w", mv.MakerId, err)
	}

	javURL := "https://" + s.deps.Config.Fetcher.JavAddress + "/cn/?v=" + mv.JavId
	updateDate := time.Unix(mv.DetailUpdateTime, 0).Format(time.DateOnly)

	// x) 扩展信息：图片/视频/编码
	minfo, err := s.deps.MinfoRepo.FindOneByJavId(ctx, mv.JavId)
	if err != nil {
		return nil, fmt.Errorf("MinfoRepo .FindOneByJavId failed: %w", err)
	}

	item, err := s.deps.ItemRepo.FindOneByJavId(ctx, mv.JavId)
	if err != nil {
		return nil, fmt.Errorf("MinfoRepo .FindOneByJavId failed: %w", err)
	}

	// 4) 主体赋值（与老项目一致）
	out := &types.MovieType{
		DbId:                 mv.Id,
		Name:                 mv.Name,
		JavId:                mv.JavId,
		Title:                chooseTitle(mv.Title, minfo.Chinese, item.HasChinese), // 见下方 chooseTitle
		ReleasingDate:        tsToDate(mv.ReleasingDate),
		UpdateDate:           updateDate,
		Length:               mv.Length,
		Score:                round1(float64(mv.Score) / 10.0),
		ViewersNumberOwned:   mv.ViewersNumberOwned,
		ViewersNumberWatched: mv.ViewersNumberWatched,
		ViewersNumberWant:    mv.ViewersNumberWant,
		Maker:                maker.Name,
		Director:             director.Name,
		Label:                label.Name,
		Genre:                genreNames,
		Cast:                 castInfos,
		CastNumber:           mv.CastNumber,
		JavUrl:               javURL,
		SearchUrl:            getMovieSearchUrl(mv.Name),
		BusUrl:               s.getBusUrl(mv.Name),
		JacketImg:            murl.JacketImg,
		SmallImg:             murl.SmallImg,
		Prefix:               prefix.Name,
		Owned:                consts.MovieTypeNotOwned,
		NeedDownload:         minfo.NeedDownload,
		EncodeName:           minfo.EncodeName,
	}

	film, err := s.deps.FilmRepo.FindOneByMovieJavId(ctx, mv.JavId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("FilmRepo.FindOneByMovieJavId failed: %w", err)
	}

	if film != nil {
		out.FilmBirthDate = tsToDate(film.BirthTime)
		out.VideoUrl = film.FullDir + string(filepath.Separator) + film.FileName
		out.ScTimes = film.ScTimes
		out.ComeTimes = film.ComeTimes
		out.Owned = determineOwnership(film)
	}

	if minfo.HighestRank < 1000 {
		out.HighestRank = minfo.HighestRank
	}
	if minfo.DaysInRank > 0 {
		out.FirstRankingDay = fmt.Sprintf("%s(%v)", consts.GetDateStringByRankDayNumber(minfo.FirstRankDayNumber), minfo.DaysInRank)
	}

	// 6) 覆盖本地封面路径
	if item.HasDownloadCover == types.ItemCoverOK {
		imgBase := s.deps.Config.Fetcher.LocalImageDir
		if murl.JacketImgLocal != "" {
			out.JacketImg = imgBase + murl.JacketImgLocal
		}
		if murl.SmallImgLocal != "" {
			out.SmallImg = imgBase + murl.SmallImgLocal
		}
	}

	return out, nil
}

// ===== 辅助函数 =====
func (s *Service) getCastInfos(ctx context.Context, movieId int64, releasingTs int64) ([]*types.CastInfo, error) {
	// ✅ 使用你已有的方法：ListCastIDsByMovie
	castIDs, err := s.deps.MovieCastRepo.ListCastIDsByMovie(ctx, movieId)
	if err != nil {
		return nil, err
	}

	infos := make([]*types.CastInfo, 0, 9)
	for i, id := range castIDs {
		if i > 8 { // 最多 9 个
			break
		}
		// ✅ 你已有的 CastRepo.FindOne
		c, err := s.deps.CastRepo.FindOne(ctx, id)
		if err != nil {
			return nil, err
		}

		ci := &types.CastInfo{
			LastScTime: c.LastScTime,
			Name:       c.Name,
			// 防止中文/空格等字符破坏 URL，做下转义
			Url: fmt.Sprintf("moviesum?cn=%s", url.QueryEscape(c.Name)),
		}

		// ✅ 按你之前在 saveParsedMovie 里用到的 CafoRepo.FindBirthByName 来算年龄
		//    这里不强依赖中文名字段（如果你有 Cafo 的中文名方法，可再补充）
		birth, found, err := s.deps.CafoRepo.FindBirthByName(ctx, c.Name)
		if err != nil {
			return nil, err
		}
		if found && birth > 0 && releasingTs > 0 {
			age := round1(float64(releasingTs-birth) / (3600.0 * 24.0 * 365.0))
			ci.NameShow = fmt.Sprintf("%s(%v)", c.Name, age)
		} else {
			ci.NameShow = c.Name
		}

		infos = append(infos, ci)
	}
	return infos, nil
}

func (s *Service) getGenreNames(ctx context.Context, movieId int64) ([]string, error) {
	// ✅ 使用你已有的方法：ListGenreIDsByMovie
	genreIDs, err := s.deps.MovieGenreRepo.ListGenreIDsByMovie(ctx, movieId)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(genreIDs))
	for _, gid := range genreIDs {
		// ✅ 你已有的 GenreRepo.FindOne
		g, err := s.deps.GenreRepo.FindOne(ctx, gid)
		if err != nil {
			return nil, err
		}
		names = append(names, g.Name)
	}
	return names, nil
}

func determineOwnership(film *types.Film) int64 {
	if film.IsRemoved == consts.FilmIsRemoved {
		return consts.MovieTypeIsRemoved
	}
	if film.HasSub == consts.FilmHasSub {
		return consts.MovieTypeOwnedAndHasSub
	}
	return consts.MovieTypeOwned
}

// ====== 工具 ======

func getMovieSearchUrl(movieName string) string {
	parts := strings.Split(movieName, "-")
	if len(parts) != 2 {
		return ""
	}
	return fmt.Sprintf("https://sukebei.nyaa.si/?f=0&c=0_0&q=%v+%v", parts[0], parts[1])
}

func (s *Service) getBusUrl(movieName string) string {
	return fmt.Sprintf("https://%v/%v", s.deps.Config.Fetcher.BusAddress, movieName) // TODO: 对齐配置字段
}

func tsToDate(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format(time.DateOnly)
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10.0
}

func chooseTitle(title, chinese string, hasChinese int64) string {
	// TODO: 用你的常量替换判断（如 MovieHasChinese == 1）
	if hasChinese == 1 && chinese != "" {
		return chinese
	}
	if chinese != "" {
		return chinese
	}
	return title
}
