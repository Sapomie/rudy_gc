// internal/contracts/sc_trigger.go
package contracts

type ScTriggerKind int

const (
	ScAdd          ScTriggerKind = iota + 1 // 执行 AddSc(dir)
	ScRebuildStats                          // 执行 SC 统计回填
)

type ScTriggerMsg struct {
	Kind            ScTriggerKind
	Dir             string // ScAdd 用
	ComeMovieJavId  string
	MovieCast       string
	DurationMinutes int64
	Fg              string
	Vessel          string
	Remarks         string
	Movies          []ScTriggerMovie
}

type ScTriggerMovie struct {
	MovieJavId string
	IsSc       int64
}
