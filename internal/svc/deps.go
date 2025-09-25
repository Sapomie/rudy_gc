package svc

import "rudy_gc/internal/repo"

// Deps / ServiceContext：全局依赖聚合（可被 spider、transport、domain 等共用）
type Deps struct {
	SeedRepo repo.SeedRepo
	// 将来可继续加：RawInventoryRepo、RawDetailRepo、Logger、Tracer、Cfg 等
}
