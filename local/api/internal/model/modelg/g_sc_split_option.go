package modelg

type GScKindOption struct {
	ID              int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name            string `gorm:"column:name;type:varchar(191);not null;uniqueIndex:uk_g_sc_option_kind_name"`
	SortOrder       int64  `gorm:"column:sort_order;not null;default:0;index:idx_g_sc_option_kind_list,priority:2"`
	IsEnabled       int64  `gorm:"column:is_enabled;type:tinyint;not null;default:1;index:idx_g_sc_option_kind_list,priority:1"`
	ScEventTimes    int64  `gorm:"column:sc_event_times;not null;default:0"`
	LastScEventTime int64  `gorm:"column:last_sc_event_time;not null;default:0"`
	CreatedOn       int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn       int64  `gorm:"column:updated_on;not null;default:0"`
}

const GScKindOptionTableName = "g_sc_option_kind"

func (GScKindOption) TableName() string { return GScKindOptionTableName }

type GScFGOption struct {
	ID              int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name            string `gorm:"column:name;type:varchar(191);not null;uniqueIndex:uk_g_sc_option_fg_name"`
	SortOrder       int64  `gorm:"column:sort_order;not null;default:0;index:idx_g_sc_option_fg_list,priority:2"`
	IsEnabled       int64  `gorm:"column:is_enabled;type:tinyint;not null;default:1;index:idx_g_sc_option_fg_list,priority:1"`
	ScEventTimes    int64  `gorm:"column:sc_event_times;not null;default:0"`
	LastScEventTime int64  `gorm:"column:last_sc_event_time;not null;default:0"`
	CreatedOn       int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn       int64  `gorm:"column:updated_on;not null;default:0"`
}

const GScFGOptionTableName = "g_sc_option_fg"

func (GScFGOption) TableName() string { return GScFGOptionTableName }

type GScVesselOption struct {
	ID              int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name            string `gorm:"column:name;type:varchar(191);not null;uniqueIndex:uk_g_sc_option_vessl_name"`
	SortOrder       int64  `gorm:"column:sort_order;not null;default:0;index:idx_g_sc_option_vessl_list,priority:2"`
	IsEnabled       int64  `gorm:"column:is_enabled;type:tinyint;not null;default:1;index:idx_g_sc_option_vessl_list,priority:1"`
	ScEventTimes    int64  `gorm:"column:sc_event_times;not null;default:0"`
	LastScEventTime int64  `gorm:"column:last_sc_event_time;not null;default:0"`
	CreatedOn       int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn       int64  `gorm:"column:updated_on;not null;default:0"`
}

const GScVesselOptionTableName = "g_sc_option_vessl"

func (GScVesselOption) TableName() string { return GScVesselOptionTableName }
