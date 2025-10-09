package movie

//func (s *Service) ListMovieFull(ctx context.Context, r *types.ListMovieFullRequest) (*types.ListMovieResponse, error) {
//	if r == nil {
//		return nil, errors.New("nil ListMovieLiteRequest")
//	}
//
//	// 组装 MovieType（需要完整聚合就用现有的 buildMovieTypeFromRepos）
//	out := make([]*types.MovieType, 0, len(rows))
//	for _, mv := range rows {
//		mt, err := s.GetMovieType(ctx, mv.JavId)
//		if err != nil {
//			return nil, err
//		}
//		out = append(out, mt)
//	}
//
//	return &types.ListMovieResponse{
//		List:  out,
//		Total: total,
//	}, nil
//}
