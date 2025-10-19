package contracts

type ProcKind int

const (
	ProcDailyBest ProcKind = iota + 1
	ProcSeeds
	ProcBoth
)

type TriggerMsg struct {
	Kind  ProcKind
	Seeds []string // 可选：指定 seeds；为空则用默认配置
	Force bool     // 预留：是否强制忽略限流等
}
