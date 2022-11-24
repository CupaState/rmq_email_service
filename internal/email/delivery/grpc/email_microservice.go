package grpc

import (
	"context"
	"rmq_service/config"
	"rmq_service/internal/email"
	emailService "rmq_service/internal/email/proto"
	"rmq_service/internal/models"
	"rmq_service/pkg/grpc_errors"
	"rmq_service/pkg/logger"
	"rmq_service/pkg/utils"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Email gRPC microservice
type EmailMicroservice struct {
	emailService.UnimplementedEmailServiceServer
	cfg 			*config.Config
	logger 		logger.Logger
	emailUC 	email.EmailsUseCase
}

// Email gRPC microservice constructor
func NewEmailMicroservice(cfg *config.Config, logger logger.Logger, emailUC email.EmailsUseCase) *EmailMicroservice	{
	return &EmailMicroservice{ cfg: cfg, logger: logger, emailUC: emailUC }
}

// Send Emails
func (e *EmailMicroservice) SendEmails(ctx context.Context, r *emailService.SendEmailsRequest) (*emailService.SendEmailsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailMicroservice.SendEmails")
	defer span.Finish()

	mail := &models.Email{
		From: 		e.cfg.Smtp.User,
		To: 			r.GetTo(),
		Body: 		r.GetBody(),
		Subject: 	r.GetSubject(),
	}

	if err := mail.PrepareAndValidate(ctx); err != nil {
		e.logger.Errorf("PrepareAndValidate: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "emailUC.publishEmailToQueue: %v", err)
	}

	return &emailService.SendEmailsResponse{Status: "Ok"}, nil
}

// Find email by id
func (e *EmailMicroservice) FindEmailById(
	ctx context.Context,
	r *emailService.FindEmailByIdRequest) (*emailService.FindEmailByIdResponse, error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "EmailMicroservice.FindEmailById")
		defer span.Finish()

		emailUUID, err := uuid.Parse(r.GetEmailUuid())
		if err != nil {
			e.logger.Errorf("uuid.Parse: %v", err)
			return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "emailService.FindEmailById: %v", err)
		}

		emailByID, err := e.emailUC.FindEmailById(ctx, emailUUID)
		if err != nil {
			e.logger.Errorf("emailUC.FindEmailById", err)
			return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "emailUC.FindEmailById")
		}

		return &emailService.FindEmailByIdResponse{Email: e.convertEmailToProto(emailByID)}, nil
}

// Find Emails By Receiver
func (e *EmailMicroservice) FindEmailsByReceiver(
	ctx context.Context, 
	r *emailService.FindEmailsByReceiverRequest) (*emailService.FindEmailsByReceiverResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailMicroservice.FindEmailByReceiver")
	defer span.Finish()

	PaginationQuery := &utils.PaginationQuery{
		Size: r.GetSize(),
		Page: r.GetPage(),
	}

	emails, err := e.emailUC.FindEmailsByReceiver(ctx, r.GetReceiverEmail(), PaginationQuery)
	if err != nil {
		e.logger.Errorf("emailUC.FindEmailByReceiver", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "emailUC.FindEmailByReceiver: %v", err)
	}

	return &emailService.FindEmailsByReceiverResponse{
		Emails: e.convertEmailsListToProto(emails.Emails),
		TotalPages: emails.TotalPages,
		TotalCount: emails.TotalCount,
		HasMore: 		emails.HasMore,
		Page:				emails.Page,
		Size: 			emails.Size,
	}, nil
}

func (e *EmailMicroservice) convertEmailToProto(email *models.Email) *emailService.Email {
	return &emailService.Email{
		EmailId: 			email.EmailID.String(),
		To: 					email.To,
		From: 				email.From,
		Body: 				email.Body,
		Subject:  		email.Subject,
		ContentType: 	email.ContentType,
		CreatedAt: 		timestamppb.New(email.CreatedAt),
	}
}

func (e *EmailMicroservice) convertEmailsListToProto(emails []*models.Email) []*emailService.Email {
	protoEmails := make([]*emailService.Email, 0, len(emails))
	for _, m := range emails {
		protoEmails = append(protoEmails, e.convertEmailToProto(m))
	}
	return protoEmails
}
