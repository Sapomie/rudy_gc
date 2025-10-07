package consts

const (
	MovieTypeNotOwned int64 = 1 + iota
	MovieTypeOwned
	MovieTypeOwnedAndHasSub
	MovieTypeIsRemoved
)
