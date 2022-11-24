package email

import (
	"context"
	"rmq_service/internal/models"
)

// Email useCase interface
type EmailsUseCase interface {
	SendEmail(ctx context.Context, deliveryBody []byte) error
	PublishEmailToQueue(ctx context.Context, email *models.Email) error
}
