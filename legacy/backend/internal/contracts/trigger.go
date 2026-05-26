// internal/contracts/trigger.go
package contracts

type TriggerKind int

const (
	ProcDailyBest TriggerKind = iota + 1
	ProcSeeds
	ProcSeedByName
	ProcSyncBest
	ProcRefreshOldestDetail // ✅ 新增
	ProcRebuildCastRank
	ProcRebuildActorRank
)

type TriggerMsg struct {
	Kind      TriggerKind
	Name      string // 给 SeedByName 用
	ActorName string // 给单演员 rank 回填用
	Number    int64  // ✅ 给 RefreshOldestDetail 用
}
