package contracts

type TriggerKind int

const (
	ProcDailyBest TriggerKind = iota + 1
	ProcSeeds
)

type TriggerMsg struct {
	Kind TriggerKind
}
