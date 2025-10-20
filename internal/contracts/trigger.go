package contracts

type TriggerKind int

const (
	ProcDailyBest TriggerKind = iota + 1
	ProcSeeds
	ProcStop // ✅ 新增：停止当前正在运行的大流程
)

type TriggerMsg struct {
	Kind  TriggerKind
	Seeds []string // 可选
}
