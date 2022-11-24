//go:generate mockgen -source usecase.go -destination mock/usecase.go -package mock

package email

import (
	"context"
	"rmq_service/internal/models"
	"rmq_service/pkg/utils"

	"github.com/google/uuid"
)

// Email useCase interface
type EmailsUseCase interface {
	SendEmails(ctx context.Context, deliveryBody []byte) error
	PublishEmailToQueue(ctx context.Context, email *models.Email) error
	FindEmailById(ctx context.Context, mailId uuid.UUID) (*models.Email, error)
	FindEmailsByReceiver(ctx context.Context, mailTo string, query *utils.PaginationQuery) (*models.EmailsList, error)
}
