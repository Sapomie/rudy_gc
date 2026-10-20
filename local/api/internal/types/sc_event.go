package types

import "io"

type ScEventListItem struct {
	Event           *GSc
	TotalMovieCount int64
	ScMovieCount    int64
	JavMovieCount   int64
	FC2MovieCount   int64
	ComeMovieType   string
	MovieCastHref   string
	ComeMovieHref   string
}

type ScEventListPage struct {
	Items []*ScEventListItem
	Total int64
}

type ScEventListFilter struct {
	StartTime      int64
	EndTime        int64
	KindOptionID   int64
	FGOptionID     int64
	VesselOptionID int64
}

type ScEventCardItem struct {
	Event              *GSc
	DetailHref         string
	ComeMovieHref      string
	ComeMovieName      string
	ComeMovieJacketImg string
	HasComeMovieCover  bool
}

type ScEventCardPage struct {
	Items []*ScEventCardItem
	Total int64
}

type ScEventDetail struct {
	Event                       *GSc
	Items                       []*ScEventMovie
	FailedItems                 []*ScEventMovie
	ComeCount                   int64
	ComeMovieHref               string
	ComeMovieType               string
	TotalMovieCount             int64
	ScMovieCount                int64
	NoScMovieCount              int64
	JavMovieCount               int64
	FC2MovieCount               int64
	EditImageName               string
	EditMovieCount              int64
	EditScMovieCount            int64
	EditComeGListId             int64
	EditComeMovieOptions        []*ScEventEditMovieOption
	EditCurrentMovieCastOptions []*ScEventCastOption
}

type ScEventEditForm struct {
	Name              string
	ComeGListId       int64
	KindOptionID      int64
	DurationMinutes   int64
	FGOptionID        int64
	VesselOptionID    int64
	MovieCastPersonID int64
	Remarks           string
	Image             io.Reader
}

type ScOption struct {
	ID              int64  `json:"id"`
	OptionType      string `json:"optionType"`
	OptionName      string `json:"optionName"`
	SortOrder       int64  `json:"sortOrder"`
	IsEnabled       int64  `json:"isEnabled"`
	ScEventTimes    int64  `json:"scEventTimes"`
	LastScEventTime int64  `json:"lastScEventTime"`
	CreatedOn       int64  `json:"createdOn"`
	UpdatedOn       int64  `json:"updatedOn"`
}

type ScOptionWriteForm struct {
	OptionType string
	OptionName string
	SortOrder  int64
	IsEnabled  int64
}

type ScEventDeleteResult struct {
	Name              string
	DeletedMovieCount int64
	DerivedDataFresh  bool
}

type ScEventEditMovieOption struct {
	GListId     int64
	MovieType   string
	MovieJavId  string
	FC2MovieId  int64
	MovieKey    string
	MovieName   string
	MovieHref   string
	CastOptions []*ScEventCastOption
	IsCome      bool
}

type ScEventCastOption struct {
	PersonID int64  `json:"personId"`
	Name     string `json:"name"`
}

type ScEventMovie struct {
	Entry      *GList
	MovieType  *MovieType
	FC2Movie   *FC2ScEventMovie
	Card       any
	MovieName  string
	IsCome     bool
	LoadError  string
	CastNames  []string
	GenreNames []string
}

type FC2ScEventMovie struct {
	ID                     int64
	FC2Key                 string
	Title                  string
	PerformerText          string
	PostTime               int64
	VideoURL               string
	PlayButtonClass        string
	PlayButtonText         string
	ShowPlayButton         bool
	DownloadTime           int64
	ScTimes                int64
	ComeTimes              int64
	LastSelectedScTime     int64
	LastParticipatedScName string
}
