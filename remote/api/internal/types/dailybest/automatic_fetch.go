package dailybest

import contracts "rudy_gc/shared/dailybest"

type AutomaticFetchState struct {
	Successful *contracts.DailyBestFetchRun
	Active     *contracts.DailyBestFetchRun
	Failed     *contracts.DailyBestFetchRun
}
