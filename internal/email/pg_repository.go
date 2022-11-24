//go:generate mockgen -source pg_repository.go -destination mock/pg_repository.go -package mock

package email

import (
	"context"
	"rmq_service/internal/models"
	"rmq_service/pkg/utils"

	"github.com/google/uuid"
)

// Repository interface
type EmailsRepository interface {
	CreateEmail(context.Context, *models.Email) (*models.Email, error)
	FindEmailById(context.Context, uuid.UUID) (*models.Email, error)
	FindEmailsByReceiver(context.Context, string, *utils.PaginationQuery) (*models.EmailsList, error)
}
