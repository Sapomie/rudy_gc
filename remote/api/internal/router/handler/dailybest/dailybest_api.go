package dailybest

import dailybestservice "rudy_gc/remote/api/internal/service/dailybest"

type DailyBestAPI struct {
	service *dailybestservice.Service
}

func NewDailyBestAPI(service *dailybestservice.Service) *DailyBestAPI {
	return &DailyBestAPI{service: service}
}
