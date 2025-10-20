package contracts

type FilmTriggerKind int

const (
	ProcFilmRename  FilmTriggerKind = iota + 1 // 触发 RenameFilm()
	ProcFilmProcess                            // 触发 ProcessFilm(ctx)
)

// FilmTriggerMsg 影片触发消息
type FilmTriggerMsg struct {
	Kind FilmTriggerKind
}
