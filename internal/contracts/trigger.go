// internal/contracts/trigger.go
package contracts

type TriggerKind int

const (
	ProcDailyBest TriggerKind = iota + 1
	ProcSeeds
	ProcSeedByName
	ProcSyncBest
	ProcRefreshOldestDetail // ✅ 新增
)

type TriggerMsg struct {
	Kind   TriggerKind
	Name   string // 给 SeedByName 用
	Number int64  // ✅ 给 RefreshOldestDetail 用
}
