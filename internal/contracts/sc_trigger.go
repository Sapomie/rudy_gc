// internal/contracts/sc_trigger.go
package contracts

type ScTriggerKind int

const (
	ScMove ScTriggerKind = iota + 1 // 执行 MoveScFilm(scName)
	ScAdd                           // 执行 AddSc(dir)
)

type ScTriggerMsg struct {
	Kind   ScTriggerKind
	ScName string // ScMove 用
	Dir    string // ScAdd  用
}
