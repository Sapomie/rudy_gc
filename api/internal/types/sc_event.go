package types

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
	Event                     *GSc
	Items                     []*ScEventMovie
	FailedItems               []*ScEventMovie
	ComeCount                 int64
	ComeMovieHref             string
	ComeMovieType             string
	TotalMovieCount           int64
	ScMovieCount              int64
	NoScMovieCount            int64
	JavMovieCount             int64
	FC2MovieCount             int64
	EditImageName             string
	EditMovieCount            int64
	EditScMovieCount          int64
	EditComeGListId           int64
	EditComeMovieOptions      []*ScEventEditMovieOption
	EditCurrentMovieCastNames []string
}

type ScEventEditForm struct {
	Name            string
	ComeGListId     int64
	Kind            string
	DurationMinutes int64
	Fg              string
	Vessel          string
	MovieCast       string
	Remarks         string
}

type ScEventEditMovieOption struct {
	GListId     int64
	MovieType   string
	MovieJavId  string
	FC2MovieId  int64
	MovieKey    string
	MovieName   string
	MovieHref   string
	CastOptions []string
	IsCome      bool
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
