package grainxpilot

import "context"

type ListRecordingsRequest struct {
	Cursor  string
	Filter  map[string]interface{}
	Include map[string]interface{}
}

type ListRecordingsResponse struct {
	Cursor     string
	Recordings []Recording
}

type GetRecordingRequest struct {
	RecordingID string
	Include     map[string]interface{}
}

type GetTranscriptRequest struct {
	RecordingID string
}

type DownloadRecordingRequest struct {
	RecordingID string
}

type CreateHookRequest struct {
	HookURL  string
	HookType string
	Include  map[string]interface{}
}

type ListHooksRequest struct {
	Filter map[string]interface{}
}

type DeleteHookRequest struct {
	HookID string
}

type GrainAdapter interface {
	ListRecordings(ctx context.Context, req ListRecordingsRequest) (ListRecordingsResponse, error)
	GetRecording(ctx context.Context, req GetRecordingRequest) (Recording, error)
	GetRecordingTranscriptJSON(ctx context.Context, req GetTranscriptRequest) ([]TranscriptSegment, error)
	GetRecordingTranscriptTXT(ctx context.Context, req GetTranscriptRequest) (string, error)
	GetRecordingTranscriptVTT(ctx context.Context, req GetTranscriptRequest) (string, error)
	GetRecordingTranscriptSRT(ctx context.Context, req GetTranscriptRequest) (string, error)
	DownloadRecording(ctx context.Context, req DownloadRecordingRequest) ([]byte, error)
	CreateHook(ctx context.Context, req CreateHookRequest) (Hook, error)
	ListHooks(ctx context.Context, req ListHooksRequest) ([]Hook, error)
	DeleteHook(ctx context.Context, req DeleteHookRequest) error
}

type NormalizeRequest struct {
	Recording        Recording
	TranscriptTXT    string
	TranscriptJSON   []TranscriptSegment
	CharBudgetPerDoc int
}

type NormalizeDoc struct {
	Slug      string
	Title     string
	Body      string
	CharCount int
}

type NormalizeResponse struct {
	Docs   []NormalizeDoc
	Review string
}

type Normalizer interface {
	Normalize(ctx context.Context, req NormalizeRequest) (NormalizeResponse, error)
}

type BrowserAttachRequest struct {
	BrowserURL  string
	WSEndpoint  string
	AutoConnect bool
}

type BrowserWorker interface {
	Attach(ctx context.Context, req BrowserAttachRequest) error
	VerifyAuth(ctx context.Context) error
	OpenUploadPage(ctx context.Context) error
	UploadDocument(ctx context.Context, docPath string) error
	WaitForQueueContains(ctx context.Context, slug string) error
	WaitForParser(ctx context.Context) error
	OpenItem(ctx context.Context, slug string) error
	AssertTitlePresent(ctx context.Context) error
	AssertStoryboardPresent(ctx context.Context) error
	TriggerRender(ctx context.Context, slug string) error
	DownloadMP4(ctx context.Context, slug string, destPath string) error
	DownloadSCORM(ctx context.Context, slug string, destPath string) error
}
