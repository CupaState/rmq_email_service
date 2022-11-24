package usecase

import (
	"context"
	"encoding/json"
	"rmq_service/config"
	"rmq_service/internal/email"
	"rmq_service/internal/models"
	"rmq_service/pkg/logger"
	"rmq_service/pkg/utils"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

// Email usecase struct
type EmailUseCase struct {
	mailer 				email.Mailer
	emailsRepo    email.EmailsRepository
	logger 				logger.Logger
	cfg 					*config.Config
	publisher 		email.EmailsPublisher
}

// EmailUseCase constructor
func NewEmailUseCase(
	mailer email.Mailer,
	emailsRepo email.EmailsRepository,
	logger logger.Logger,
	cfg *config.Config,
	publisher email.EmailsPublisher) *EmailUseCase {
		return &EmailUseCase{ mailer: mailer,emailsRepo: emailsRepo,logger: logger,cfg: cfg,publisher: publisher }
}

// Send Email
func (e *EmailUseCase) SendEmails(ctx context.Context, deliveryBody []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.SendEmails")
	defer span.Finish()

	mail := &models.Email{}
	if err := json.Unmarshal(deliveryBody, mail); err != nil {
		return errors.Wrap(err, "json.Unmarshal")
	}

	mail.Body = utils.SanitizeString(mail.Body)
	mail.From = e.cfg.Smtp.User

	if err := utils.ValidateStruct(ctx, mail); err != nil {
		return errors.Wrap(err, "ValidateStruct")
	}

	if err := e.mailer.Send(ctx, mail); err != nil {
		return errors.Wrap(err, "mailer.Send")
	}

	createdEmail, err := e.emailsRepo.CreateEmail(ctx, mail)
	if err != nil {
		return errors.Wrap(err, "emailRepo.CreateEmail")
	}

	span.LogFields(log.String("emailID", createdEmail.EmailID.String()))
	e.logger.Infof("Success sent email: %v", createdEmail.EmailID)
	return nil
}

func (e *EmailUseCase) PublishEmailToQueue(ctx context.Context, email *models.Email) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "EmailUseCase.PublishEmailToQueue")
	defer span.Finish()

	mailBytes, err := json.Marshal(email)
	if err != nil {
		return errors.Wrap(err, "json.Marshall")
	}

	return e.publisher.Publish(mailBytes, email.ContentType)
}

// Find email by uuid
func (e *EmailUseCase) FindEmailById(ctx context.Context, emailID uuid.UUID) (*models.Email, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailsUseCase.FindEmailById")
	defer span.Finish()

	return e.emailsRepo.FindEmailById(ctx, emailID)
}

// Find emails by receiver
func (e *EmailUseCase) FindEmailsByReceiver(
		ctx context.Context,
		mailTo string,
		query *utils.PaginationQuery,
	) (*models.EmailsList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailsUseCase.FindEmailsByReceiver")
	defer span.Finish()

	return e.emailsRepo.FindEmailsByReceiver(ctx, mailTo, query)
}
