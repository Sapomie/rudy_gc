package contracts

type TriggerKind int

const (
	ProcDailyBest TriggerKind = iota + 1
	ProcSeeds
	ProcSeedByName // ✅ 新增：按 Seed 名称触发
)

type TriggerMsg struct {
	Kind TriggerKind
	Name string // ✅ 新增：携带 seed 名称
}
