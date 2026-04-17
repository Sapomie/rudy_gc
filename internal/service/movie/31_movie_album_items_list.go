package movie

import (
	"context"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
)

type MovieAlbumItemsPageQuery struct {
	AlbumName string
	Page      int64
	PageSize  int64
	Sort      string
	Order     string
	Keyword   string
}

type MovieAlbumItemsPageResult struct {
	SelectedAlbumID   int64
	SelectedAlbumName string
	Albums            []*MovieAlbumOption
	Items             []*MovieAlbumItemRow
	Total             int64
	Page              int64
	PageSize          int64
}

func (s *Service) BuildMovieAlbumItemsPage(ctx context.Context, query MovieAlbumItemsPageQuery) (*MovieAlbumItemsPageResult, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	selectedAlbumName := normalizeMovieAlbumName(query.AlbumName)
	albums, err := s.deps.MovieAlbumModel.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	options := buildMovieAlbumOptions(albums, selectedAlbumName)
	selectedAlbumID := int64(0)
	for _, option := range options {
		if option == nil || !option.Selected {
			continue
		}
		selectedAlbumID = option.ID
		selectedAlbumName = option.Name
		break
	}
	if selectedAlbumID <= 0 && len(options) > 0 {
		selectedAlbumID = options[0].ID
		selectedAlbumName = options[0].Name
		options[0].Selected = true
	}

	result := &MovieAlbumItemsPageResult{
		SelectedAlbumID:   selectedAlbumID,
		SelectedAlbumName: selectedAlbumName,
		Albums:            options,
		Items:             []*MovieAlbumItemRow{},
		Total:             0,
		Page:              page,
		PageSize:          pageSize,
	}
	if selectedAlbumID <= 0 {
		return result, nil
	}

	total, err := s.deps.MovieAlbumItemModel.CountPageRows(ctx, selectedAlbumID, strings.TrimSpace(query.Keyword))
	if err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	rows, err := s.deps.MovieAlbumItemModel.ListPageRows(ctx, selectedAlbumID, offset, pageSize, movieAlbumItemPageOrderBy(query.Sort, query.Order), strings.TrimSpace(query.Keyword))
	if err != nil {
		return nil, err
	}

	items := make([]*MovieAlbumItemRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, &MovieAlbumItemRow{
			ID:         row.Id,
			AlbumID:    row.AlbumId,
			MovieJavID: strings.TrimSpace(row.MovieJavId),
			MovieName:  strings.TrimSpace(row.MovieName),
			SortNo:     row.SortNo,
			CreatedOn:  row.CreatedOn,
		})
	}
	result.Items = items
	result.Total = total
	return result, nil
}

func buildMovieAlbumOptions(rows []*moviex.CMovieAlbum, selectedAlbumName string) []*MovieAlbumOption {
	options := make([]*MovieAlbumOption, 0, len(rows)+1)
	hasSelected := false
	selectedAlbumName = strings.TrimSpace(selectedAlbumName)
	for _, row := range rows {
		if row == nil {
			continue
		}
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		selected := strings.EqualFold(name, selectedAlbumName)
		if selected {
			hasSelected = true
		}
		options = append(options, &MovieAlbumOption{ID: row.Id, Name: name, Selected: selected})
	}
	if !hasSelected && selectedAlbumName != "" {
		options = append(options, &MovieAlbumOption{ID: 0, Name: selectedAlbumName, Selected: true})
	}
	return options
}

func movieAlbumItemPageOrderBy(sortField string, sortOrder string) string {
	column := mapMovieAlbumItemSortColumn(sortField)
	order := normalizeMovieAlbumItemsSortOrder(sortOrder)
	return column + " " + order + ", `id` DESC"
}

func mapMovieAlbumItemSortColumn(sortField string) string {
	switch strings.TrimSpace(sortField) {
	case "movie_name":
		return "`movie_name`"
	case "movie_jav_id":
		return "`movie_jav_id`"
	case "sort_no":
		return "`sort_no`"
	case "created_on":
		return "`created_on`"
	default:
		return "`created_on`"
	}
}

func normalizeMovieAlbumItemsSortOrder(sortOrder string) string {
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		return "ASC"
	}
	return "DESC"
}
