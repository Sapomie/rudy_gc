package types

type ActorScEventItem struct {
	ScName     string
	ScTime     int64
	Cooldown   int64
	MovieCount int
	IsCome     bool
	Href       string
	Movies     []*ActorScEventMovie
}

type ActorScEventMovie struct {
	Name   string
	Href   string
	IsCome bool
}

type ActorScSummary struct {
	Actor         *Cast
	RecentEvents  []*ActorScEventItem
	ActorPageHref string
	CardsHref     string
	ScAggHref     string
}

type ActorScPage struct {
	Actor         *Cast
	RecentEvents  []*ActorScEventItem
	Movies        []*MovieType
	TotalMovies   int
	ActorPageHref string
	CardsHref     string
	ScAggHref     string
}
