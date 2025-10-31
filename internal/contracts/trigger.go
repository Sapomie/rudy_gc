// internal/contracts/trigger.go
package contracts

type TriggerKind int

const (
	ProcDailyBest TriggerKind = iota + 1
	ProcSeeds
	ProcSeedByName
	ProcSyncBest // ✅ 新增：同步每日榜
)

type TriggerMsg struct {
	Kind TriggerKind
	Name string // 给 ProcSeedByName 用
}
