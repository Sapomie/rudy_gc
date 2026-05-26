package types

type ItemListFilter struct {
	JavID string
	Name  string

	HasDetail    int64
	HasDetailSet bool

	HasDownloadCover    int64
	HasDownloadCoverSet bool

	HasChinese    int64
	HasChineseSet bool

	DetailNeedScan    int64
	DetailNeedScanSet bool

	DetailBirthTimeFrom    int64
	HasDetailBirthTimeFrom bool
	DetailBirthTimeTo      int64
	HasDetailBirthTimeTo   bool

	LastQueryDetailTimeFrom    int64
	HasLastQueryDetailTimeFrom bool
	LastQueryDetailTimeTo      int64
	HasLastQueryDetailTimeTo   bool
}

type ItemListRow struct {
	Id                    int64  `json:"id"`
	JavId                 string `json:"jav_id"`
	JavUrl                string `json:"jav_url"`
	Name                  string `json:"name"`
	HasDetail             int64  `json:"has_detail"`
	HasDetailValue        string `json:"has_detail_value"`
	HasDetailText         string `json:"has_detail_text"`
	HasDownloadCover      int64  `json:"has_download_cover"`
	HasDownloadCoverValue string `json:"has_download_cover_value"`
	HasDownloadCoverText  string `json:"has_download_cover_text"`
	HasChinese            int64  `json:"has_chinese"`
	HasChineseValue       string `json:"has_chinese_value"`
	HasChineseText        string `json:"has_chinese_text"`
	DetailNeedScan        int64  `json:"detail_need_scan"`
	DetailNeedScanValue   string `json:"detail_need_scan_value"`
	DetailNeedScanText    string `json:"detail_need_scan_text"`
	DetailBirthTime       int64  `json:"detail_birth_time"`
	LastQueryDetailTime   int64  `json:"last_query_detail_time"`
}
