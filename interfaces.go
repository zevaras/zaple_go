package zaple

import "context"

// MessagingAPI is the interface implemented by MessagingService.
// Declare this type in your application code and accept it instead of the
// concrete *MessagingService to allow substituting a mock in tests.
//
//	type MyApp struct {
//	    messaging zaple.MessagingAPI
//	}
type MessagingAPI interface {
	SendTemplate(ctx context.Context, req *SendTemplateRequest) (*SendTemplateResponse, error)
	CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*CreateTemplateResponse, error)
	GetMessageStatus(ctx context.Context, messageID string) (*MessageStatus, error)
	GetTemplateDetails(ctx context.Context, templateID string) (*TemplateDetails, error)
	GetTemplateStatus(ctx context.Context, templateID string) (*TemplateStatus, error)
	GetMessageCount(ctx context.Context, params *MessageCountParams) (*MessageCount, error)
}

// BatchAPI is the interface implemented by BatchService.
// Use it in your application code to enable easy mocking in tests.
type BatchAPI interface {
	Create(ctx context.Context, req *CreateBatchRequest) (*Batch, error)
	UpsertContacts(ctx context.Context, batchID string, contacts []BatchContact) (*UpsertContactsResponse, error)
	Send(ctx context.Context, batchID string, req *SendBatchRequest) (*SendBatchResponse, error)
	GetStatus(ctx context.Context, batchID string) (*BatchStatusResponse, error)
	GetDetails(ctx context.Context, batchID string) (*Batch, error)
	Delete(ctx context.Context, batchID string) error
}

// Compile-time assertions that the concrete services satisfy their interfaces.
var (
	_ MessagingAPI = (*MessagingService)(nil)
	_ BatchAPI     = (*BatchService)(nil)
)
