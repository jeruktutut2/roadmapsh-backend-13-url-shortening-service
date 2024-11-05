package services

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"
	"url-shortening-service/helpers"
	modelentities "url-shortening-service/models/entities"
	modelrequests "url-shortening-service/models/requests"
	modelresponses "url-shortening-service/models/responses"
	"url-shortening-service/repositories"
	"url-shortening-service/utils"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ShortenService interface {
	Create(ctx context.Context, createShortenRequest modelrequests.CreateShortenRequest, now int64) (httpCode int, response interface{})
	GetOriginalUrl(ctx context.Context, shortCode string) (httpCode int, url string, response interface{})
	UpdateUrl(ctx context.Context, updateUrlShortenRequest modelrequests.UpdateUrlShortenRequest, shortCode string, now int64) (httpCode int, response interface{})
	Delete(ctx context.Context, shortCode string) (httpCode int, response interface{})
	GetUrlStatistics(ctx context.Context, shortCode string) (httpCode int, response interface{})
}

type ShortenServiceImplementation struct {
	PostgresUtil      utils.PostgresUtil
	Validate          *validator.Validate
	ShortenRepository repositories.ShortenRepository
}

func NewShortenService(postgresUtil utils.PostgresUtil, validate *validator.Validate, shortenRepository repositories.ShortenRepository) ShortenService {
	return &ShortenServiceImplementation{
		PostgresUtil:      postgresUtil,
		Validate:          validate,
		ShortenRepository: shortenRepository,
	}
}

func (service *ShortenServiceImplementation) Create(ctx context.Context, createShortenRequest modelrequests.CreateShortenRequest, now int64) (httpCode int, response interface{}) {
	err := service.Validate.Struct(createShortenRequest)
	if err != nil {
		httpCode = http.StatusBadRequest
		response = helpers.ToResponse(err.Error())
		return
	}

	tx, err := service.PostgresUtil.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	defer func() {
		errCommitOrRollback := service.PostgresUtil.CommitOrRollback(tx, ctx, err)
		if errCommitOrRollback != nil {
			httpCode = http.StatusInternalServerError
			response = helpers.ToResponse(errCommitOrRollback.Error())
		}
	}()

	var shorten modelentities.Shorten
	shorten.Url = pgtype.Text{Valid: true, String: createShortenRequest.Url}
	shorten.ShortCode = pgtype.Text{Valid: true, String: shorteningCode(6)}
	shorten.CreatedAt = pgtype.Int8{Valid: true, Int64: now}
	shorten.UpdatedAt = pgtype.Int8{Valid: true, Int64: now}
	shorten.AccessCount = pgtype.Int4{Valid: false, Int32: 0}
	lastInsertedId, err := service.ShortenRepository.Create(tx, ctx, shorten)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	var createShortenResponse modelresponses.CreateShortenResponse
	createShortenResponse.Id = lastInsertedId
	createShortenResponse.Url = shorten.Url.String
	createShortenResponse.ShortCode = shorten.ShortCode.String
	createShortenResponse.CreatedAt = time.Unix(shorten.CreatedAt.Int64/1000, 0).UTC().Format("2006-01-02T15:04:05Z")
	createShortenResponse.UpdatedAt = time.Unix(shorten.UpdatedAt.Int64/1000, 0).UTC().Format("2006-01-02T15:04:05Z")

	httpCode = http.StatusCreated
	response = createShortenResponse
	return
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func shorteningCode(n int) string {
	result := make([]byte, n)
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range result {
		result[i] = charset[seed.Intn(len(charset))]
	}
	return string(result)
}

func (service *ShortenServiceImplementation) GetOriginalUrl(ctx context.Context, shortCode string) (httpCode int, url string, response interface{}) {
	tx, err := service.PostgresUtil.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	defer func() {
		errCommitOrRollback := service.PostgresUtil.CommitOrRollback(tx, ctx, err)
		if errCommitOrRollback != nil {
			httpCode = http.StatusInternalServerError
			response = helpers.ToResponse(errCommitOrRollback.Error())
		}
	}()

	shorten, err := service.ShortenRepository.GetByShortCodeTx(tx, ctx, shortCode)
	if err == pgx.ErrNoRows {
		httpCode = http.StatusNotFound
		response = helpers.ToResponse(err.Error())
		return
	} else if err != nil && err != pgx.ErrNoRows {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	accessCount := shorten.AccessCount.Int32 + 1
	rowAffected, err := service.ShortenRepository.UpdateAccessCount(tx, ctx, int(accessCount), int(shorten.Id.Int32))
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	if rowAffected != 1 {
		err = errors.New("rows affected not one")
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	var getShortenResponse modelresponses.GetShortenResponse
	getShortenResponse.Id = int(shorten.Id.Int32)
	getShortenResponse.Url = shorten.Url.String
	getShortenResponse.ShortCode = shorten.ShortCode.String
	getShortenResponse.CreatedAt = time.Unix(shorten.CreatedAt.Int64/1000, 0).UTC().Format("2006-01-02T15:04:05Z")
	getShortenResponse.UpdatedAt = time.Unix(shorten.UpdatedAt.Int64/1000, 0).UTC().Format("2006-01-02T15:04:05Z")

	httpCode = http.StatusOK
	url = shorten.Url.String
	response = getShortenResponse
	return
}

func (service *ShortenServiceImplementation) UpdateUrl(ctx context.Context, updateUrlShortenRequest modelrequests.UpdateUrlShortenRequest, shortCode string, now int64) (httpCode int, response interface{}) {
	err := service.Validate.Struct(updateUrlShortenRequest)
	if err != nil {
		httpCode = http.StatusBadRequest
		response = helpers.ToResponse(err.Error())
		return
	}

	tx, err := service.PostgresUtil.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	defer func() {
		errCommitOrRollback := service.PostgresUtil.CommitOrRollback(tx, ctx, err)
		if errCommitOrRollback != nil {
			httpCode = http.StatusInternalServerError
			response = helpers.ToResponse(errCommitOrRollback.Error())
		}
	}()

	shorten, err := service.ShortenRepository.GetByShortCodeTx(tx, ctx, shortCode)
	if err == pgx.ErrNoRows {
		httpCode = http.StatusNotFound
		response = helpers.ToResponse(err.Error())
		return
	} else if err != nil && err != pgx.ErrNoRows {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	shorten.Url = pgtype.Text{Valid: true, String: updateUrlShortenRequest.Url}
	shorten.UpdatedAt = pgtype.Int8{Valid: true, Int64: now}
	rowsAffected, err := service.ShortenRepository.UpdateUrl(tx, ctx, shorten)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	if rowsAffected != 1 {
		err = errors.New("rows affected not one")
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	var updateUrlShortenResponse modelresponses.UpdateUrlShortenResponse
	updateUrlShortenResponse.Id = int(shorten.Id.Int32)
	updateUrlShortenResponse.Url = shorten.Url.String
	updateUrlShortenResponse.ShortCode = shorten.ShortCode.String
	updateUrlShortenResponse.CreatedAt = time.Unix(shorten.CreatedAt.Int64/1000, 0).UTC().Format("2006-01-02T15:04:05Z")
	updateUrlShortenResponse.UpdatedAt = time.Unix(shorten.UpdatedAt.Int64/1000, 0).UTC().Format("2006-01-02T15:04:05Z")

	httpCode = http.StatusOK
	response = updateUrlShortenResponse
	return
}

func (service *ShortenServiceImplementation) Delete(ctx context.Context, shortCode string) (httpCode int, response interface{}) {
	tx, err := service.PostgresUtil.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	defer func() {
		errCommitOrRollback := service.PostgresUtil.CommitOrRollback(tx, ctx, err)
		if errCommitOrRollback != nil {
			httpCode = http.StatusInternalServerError
			response = helpers.ToResponse(errCommitOrRollback.Error())
		}
	}()

	shorten, err := service.ShortenRepository.GetByShortCodeTx(tx, ctx, shortCode)
	if err == pgx.ErrNoRows {
		httpCode = http.StatusNotFound
		response = helpers.ToResponse(err.Error())
		return
	}
	if err != nil && err != pgx.ErrNoRows {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	rowsAffected, err := service.ShortenRepository.DeleteById(tx, ctx, int(shorten.Id.Int32))
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	if rowsAffected != 1 {
		err = errors.New("rows affected not one")
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	httpCode = http.StatusNoContent
	response = helpers.ToResponse("successfully deleted")
	return
}

func (service *ShortenServiceImplementation) GetUrlStatistics(ctx context.Context, shortCode string) (httpCode int, response interface{}) {
	shorten, err := service.ShortenRepository.GetByShortCode(service.PostgresUtil.GetPool(), ctx, shortCode)
	if err == pgx.ErrNoRows {
		httpCode = http.StatusNotFound
		response = helpers.ToResponse(err.Error())
		return
	} else if err != nil && err != pgx.ErrNoRows {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	var getUrlStatisticResponse modelresponses.GetUrlStatisticShortenResponse
	getUrlStatisticResponse.Id = int(shorten.Id.Int32)
	getUrlStatisticResponse.Url = shorten.Url.String
	getUrlStatisticResponse.ShortCode = shorten.ShortCode.String
	getUrlStatisticResponse.CreatedAt = time.Unix(shorten.CreatedAt.Int64/1000, 0).UTC().Format("2006-01-02T15:04:05Z")
	getUrlStatisticResponse.UpdatedAt = time.Unix(shorten.UpdatedAt.Int64/1000, 0).UTC().Format("2006-01-02T15:04:05Z")
	getUrlStatisticResponse.AccessCount = int(shorten.AccessCount.Int32)

	httpCode = http.StatusOK
	response = getUrlStatisticResponse
	return
}
