package dailybest

const (
	FetchLogStatusRequesting = "requesting"
	FetchLogStatusSucceeded  = "succeeded"
	FetchLogStatusFailed     = "failed"
)

type FetchLogDraft struct {
	FetchRunID    int64
	PageNumber    int64
	AttemptNumber int64
	RequestURL    string
}

type FetchLogResult struct {
	Status       string
	HTTPStatus   int64
	DurationMS   int64
	ErrorMessage string
}

type FetchLog struct {
	ID            int64
	FetchRunID    int64
	PageNumber    int64
	AttemptNumber int64
	RequestURL    string
	Status        string
	HTTPStatus    int64
	DurationMS    int64
	ErrorMessage  string
	CreatedOn     int64
	UpdatedOn     int64
}

type FetchLogFilter struct {
	PageQuery
	FetchRunID int64
	PageNumber int64
	Status     string
}
