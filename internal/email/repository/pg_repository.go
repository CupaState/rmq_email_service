package repository

import (
	"context"
	"log"
	"rmq_service/internal/models"
	"rmq_service/pkg/utils"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/internal/errors"
)

// Emails Repository
type EmailsRepository struct {
	db *sqlx.DB
}

// Images AWS repository constructor
func NewEmailsRepository(db *sqlx.DB) *EmailsRepository {
	return &EmailsRepository{db: db}
}

// Create email
func (r *EmailsRepository) CreateEmail(ctx context.Context, email *models.Email) (*models.Email, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailsRepository.CreateEmail")
	defer span.Finish()

	var id uuid.UUID
	if err := r.db.QueryRowContext(
		ctx,
		createEmailQuery,
		email.GetToString(),
		email.From,
		email.Subject,
		email.Body,
		email.ContentType,
	).Scan(&id); err != nil {
		log.Fatalf("repository::QueryRowContext(): %v", err)
		return nil, err
	}

	email.EmailID = id
	return email, nil
}

// FindEmailById
func (r *EmailsRepository) FindEmailById(ctx context.Context, id uuid.UUID) (*models.Email, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailsRepository.FindEmailById")
	defer span.Finish()

	var to string
	email := &models.Email{}

	if err := r.db.QueryRowContext(ctx, findEmailByIdQuery, id).Scan(
		&email.EmailID,
		&to,
		&email.From,
		&email.Subject,
		&email.Body,
		&email.ContentType,
		&email.CreatedAt,
	); err != nil {
		log.Fatalf("repository::FindEmailById: %v", err)
		return nil, err
	}

	email.SetToFromString(to)

	return email, nil
}

// FindEmailsByReceiver
func (r *EmailsRepository) FindEmailsByReceiver(
	ctx context.Context, 
	to string, 
	query *utils.PaginationQuery) (list *models.EmailsList, err error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "EmailsRepository.FindEmailsByReceiver")
		defer span.Finish()

		var totalCount uint64
		if err := r.db.QueryRowContext(ctx ,totalCountQuery, to).Scan(&totalCount); err != nil {
			log.Fatalf("EmailsRepository.FindEmailsByReceiver QueryRowContext(): %v", err)
			return nil, err
		}

		if totalCount == 0 {
			return &models.EmailsList{Emails: []*models.Email{}}, nil
		}

		rows, err := r.db.QueryxContext(ctx, findEmailByReceiverQuery, to, query.GetOffset(), query.GetLimit())
		if err != nil {
			log.Fatalf("EmailsRepository.FindEmailsByReceiver QueryxContext(): %v", err)
			return nil, err
		}
		defer func() {
			if closeErr := rows.Close(); closeErr != nil {
				err = errors.Wrap(closeErr, "rows.Close")

			}
		}()

		if err := rows.Err(); err != nil {
			return nil, errors.Wrap(err, "rows.Err")
		}

		emails := make([]*models.Email, 0, query.GetSize())
		for rows.Next() {
			var mailTo string
			email := &models.Email{}

			if err := rows.Scan(
				&email.EmailID,
				&mailTo,
				&email.From,
				&email.Subject,
				&email.Body,
				&email.ContentType,
				&email.CreatedAt,
			); err != nil {
				return nil, errors.Wrap(err, "rows.Scan")
			}

			email.SetToFromString(mailTo)
			emails = append(emails, email)
		}

		return &models.EmailsList{
			TotalCount: totalCount,
			TotalPages: utils.GetTotalPages(totalCount, query.GetSize()),
			Page: 			query.Page,
			Size: 			query.Size,
			HasMore: 		utils.GetHasMore(query.GetPage(), totalCount, query.GetSize()),
			Emails:			emails,
		}, err
	}
