package grainxpilot

import "context"

type StubGrainAdapter struct{}

func (StubGrainAdapter) ListRecordings(context.Context, ListRecordingsRequest) (ListRecordingsResponse, error) {
	return ListRecordingsResponse{}, ErrNotImplemented
}

func (StubGrainAdapter) GetRecording(context.Context, GetRecordingRequest) (Recording, error) {
	return Recording{}, ErrNotImplemented
}

func (StubGrainAdapter) GetRecordingTranscriptJSON(context.Context, GetTranscriptRequest) ([]TranscriptSegment, error) {
	return nil, ErrNotImplemented
}

func (StubGrainAdapter) GetRecordingTranscriptTXT(context.Context, GetTranscriptRequest) (string, error) {
	return "", ErrNotImplemented
}

func (StubGrainAdapter) GetRecordingTranscriptVTT(context.Context, GetTranscriptRequest) (string, error) {
	return "", ErrNotImplemented
}

func (StubGrainAdapter) GetRecordingTranscriptSRT(context.Context, GetTranscriptRequest) (string, error) {
	return "", ErrNotImplemented
}

func (StubGrainAdapter) DownloadRecording(context.Context, DownloadRecordingRequest) ([]byte, error) {
	return nil, ErrNotImplemented
}

func (StubGrainAdapter) CreateHook(context.Context, CreateHookRequest) (Hook, error) {
	return Hook{}, ErrNotImplemented
}

func (StubGrainAdapter) ListHooks(context.Context, ListHooksRequest) ([]Hook, error) {
	return nil, ErrNotImplemented
}

func (StubGrainAdapter) DeleteHook(context.Context, DeleteHookRequest) error {
	return ErrNotImplemented
}

type StubNormalizer struct{}

func (StubNormalizer) Normalize(context.Context, NormalizeRequest) (NormalizeResponse, error) {
	return NormalizeResponse{}, ErrNotImplemented
}

type StubBrowserWorker struct{}

func (StubBrowserWorker) Attach(context.Context, BrowserAttachRequest) error {
	return ErrNotImplemented
}
func (StubBrowserWorker) VerifyAuth(context.Context) error             { return ErrNotImplemented }
func (StubBrowserWorker) OpenUploadPage(context.Context) error         { return ErrNotImplemented }
func (StubBrowserWorker) UploadDocument(context.Context, string) error { return ErrNotImplemented }
func (StubBrowserWorker) WaitForQueueContains(context.Context, string) error {
	return ErrNotImplemented
}
func (StubBrowserWorker) WaitForParser(context.Context) error               { return ErrNotImplemented }
func (StubBrowserWorker) OpenItem(context.Context, string) error            { return ErrNotImplemented }
func (StubBrowserWorker) AssertTitlePresent(context.Context) error          { return ErrNotImplemented }
func (StubBrowserWorker) AssertStoryboardPresent(context.Context) error     { return ErrNotImplemented }
func (StubBrowserWorker) TriggerRender(context.Context, string) error       { return ErrNotImplemented }
func (StubBrowserWorker) DownloadMP4(context.Context, string, string) error { return ErrNotImplemented }
func (StubBrowserWorker) DownloadSCORM(context.Context, string, string) error {
	return ErrNotImplemented
}
