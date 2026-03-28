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

type ActorScPage struct {
	Actor         *Person
	ActorAliases  []string
	RecentEvents  []*ActorScEventItem
	Movies        []*MovieType
	TotalMovies   int
	ActorPageHref string
	CardsHref     string
	ScAggHref     string
}
