// pkg/ptr/ptr.go
package ptr

func Int64(v int64) *int64       { return &v }
func Int(v int) *int             { return &v }
func String(v string) *string    { return &v }
func Bool(v bool) *bool          { return &v }
func Float64(v float64) *float64 { return &v }
