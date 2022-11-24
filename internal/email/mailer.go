//go:generate mockgen -source mailer.go -destination mock/mailer.go -package mock
package email

import (
	"context"
	"rmq_service/internal/models"
)

// Mailer interface
type Mailer interface {
	Send(context.Context, *models.Email) error
}
