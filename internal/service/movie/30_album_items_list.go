package movie

import (
	"context"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
)

const (
	albumSourceTypeJavbusMagnet = "javbus_magnet"
	albumSourceTypeSukebei      = "sukebei_torrent"
	albumSourceTypeSehuatang    = "sehuatang_magnet"
)

type AlbumItemsPageQuery struct {
	AlbumName          string
	Page               int64
	PageSize           int64
	Sort               string
	Order              string
	Keyword            string
	SourceType         string
	InfoHash           string
	PublishTimeFrom    int64
	PublishTimeTo      int64
	HasPublishTimeFrom bool
	HasPublishTimeTo   bool
	CreatedOnFrom      int64
	CreatedOnTo        int64
	HasCreatedOnFrom   bool
	HasCreatedOnTo     bool
}

type AlbumItemsPageResult struct {
	SelectedAlbumID   int64
	SelectedAlbumName string
	Albums            []*AlbumOption
	Items             []*AlbumItemPageRow
	Total             int64
	Page              int64
	PageSize          int64
}

type AlbumOption struct {
	ID   int64
	Name string
}

type AlbumItemPageRow struct {
	ID             int64
	AlbumID        int64
	MovieJavID     string
	MovieName      string
	SourceType     string
	SourceTypeText string
	InfoHash       string
	Size           string
	PublishTime    int64
	CreatedOn      int64
}

func (s *Service) BuildAlbumItemsPage(ctx context.Context, query AlbumItemsPageQuery) (*AlbumItemsPageResult, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	selectedAlbumName := strings.TrimSpace(query.AlbumName)
	if selectedAlbumName == "" {
		selectedAlbumName = defaultAlbumName
	}

	albums, err := s.deps.AlbumModel.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	selectedAlbumID := int64(0)
	for _, album := range albums {
		if album == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(album.Name), selectedAlbumName) {
			selectedAlbumID = album.Id
			selectedAlbumName = strings.TrimSpace(album.Name)
			break
		}
	}

	if selectedAlbumID <= 0 && selectedAlbumName == defaultAlbumName {
		id, createErr := s.getDefaultAlbumID(ctx, true)
		if createErr != nil {
			return nil, createErr
		}
		selectedAlbumID = id
		albums, err = s.deps.AlbumModel.ListAll(ctx)
		if err != nil {
			return nil, err
		}
	}

	options := buildAlbumOptions(albums, selectedAlbumName)
	result := &AlbumItemsPageResult{
		SelectedAlbumID:   selectedAlbumID,
		SelectedAlbumName: selectedAlbumName,
		Albums:            options,
		Items:             []*AlbumItemPageRow{},
		Total:             0,
		Page:              page,
		PageSize:          pageSize,
	}
	if selectedAlbumID <= 0 {
		return result, nil
	}

	filter := moviex.AlbumItemPageFilter{
		Keyword:            query.Keyword,
		SourceType:         strings.TrimSpace(query.SourceType),
		SourceTypeSet:      strings.TrimSpace(query.SourceType) != "",
		InfoHash:           query.InfoHash,
		PublishTimeFrom:    query.PublishTimeFrom,
		PublishTimeTo:      query.PublishTimeTo,
		HasPublishTimeFrom: query.HasPublishTimeFrom,
		HasPublishTimeTo:   query.HasPublishTimeTo,
		CreatedOnFrom:      query.CreatedOnFrom,
		CreatedOnTo:        query.CreatedOnTo,
		HasCreatedOnFrom:   query.HasCreatedOnFrom,
		HasCreatedOnTo:     query.HasCreatedOnTo,
	}
	total, err := s.deps.AlbumItemModel.CountPageRows(ctx, selectedAlbumID, filter)
	if err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	rows, err := s.deps.AlbumItemModel.ListPageRows(ctx, selectedAlbumID, offset, pageSize, albumItemPageOrderBy(query.Sort, query.Order), filter)
	if err != nil {
		return nil, err
	}

	items := make([]*AlbumItemPageRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, &AlbumItemPageRow{
			ID:             row.Id,
			AlbumID:        row.AlbumId,
			MovieJavID:     strings.TrimSpace(row.MovieJavId),
			MovieName:      strings.TrimSpace(row.MovieName),
			SourceType:     strings.TrimSpace(row.SourceType),
			SourceTypeText: albumItemSourceTypeText(row.SourceType),
			InfoHash:       strings.TrimSpace(row.InfoHash),
			Size:           strings.TrimSpace(row.Size),
			PublishTime:    row.PublishTime,
			CreatedOn:      row.CreatedOn,
		})
	}

	result.Items = items
	result.Total = total
	return result, nil
}

func buildAlbumOptions(rows []*moviex.TAlbum, selectedAlbumName string) []*AlbumOption {
	options := make([]*AlbumOption, 0, len(rows)+1)
	hasSelected := false
	for _, row := range rows {
		if row == nil {
			continue
		}
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		if strings.EqualFold(name, strings.TrimSpace(selectedAlbumName)) {
			hasSelected = true
		}
		options = append(options, &AlbumOption{
			ID:   row.Id,
			Name: name,
		})
	}
	if !hasSelected && strings.TrimSpace(selectedAlbumName) != "" {
		options = append(options, &AlbumOption{
			ID:   0,
			Name: strings.TrimSpace(selectedAlbumName),
		})
	}
	return options
}

func albumItemSourceTypeText(sourceType string) string {
	switch strings.TrimSpace(sourceType) {
	case albumSourceTypeJavbusMagnet:
		return "JavBus 磁力"
	case albumSourceTypeSukebei:
		return "Sukebei 种子"
	case albumSourceTypeSehuatang:
		return "Sehuatang 磁力"
	default:
		return "-"
	}
}

func albumItemPageOrderBy(sortField string, sortOrder string) string {
	column := mapAlbumItemSortColumn(sortField)
	order := normalizeAlbumItemsSortOrder(sortOrder)
	return column + " " + order + ", `id` DESC"
}

func mapAlbumItemSortColumn(sortField string) string {
	switch strings.TrimSpace(sortField) {
	case "movie_name":
		return "`movie_name`"
	case "source_type":
		return "`source_type`"
	case "publish_time":
		return "`publish_time`"
	case "created_on":
		return "`created_on`"
	default:
		return "`created_on`"
	}
}

func normalizeAlbumItemsSortOrder(sortOrder string) string {
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		return "ASC"
	}
	return "DESC"
}
