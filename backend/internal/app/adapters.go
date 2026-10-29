package app

import (
	"context"
	"errors"
	"fmt"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/catalog"
	"heritage-backend/internal/domain/media"
	"heritage-backend/internal/domain/submission"
	"heritage-backend/internal/infrastructure/osrm"
	"heritage-backend/internal/infrastructure/yookassa"
	routesvc "heritage-backend/internal/service/route"
	shopsvc "heritage-backend/internal/service/shop"
	submissionsvc "heritage-backend/internal/service/submission"
)

// Адаптеры связывают модули между собой. Каждый модуль объявил в своём ports.go
// маленький интерфейс со СВОИМИ типами — так модули не импортируют друг друга.
// Здесь, в единственном месте, которое знает все модули, типы переводятся.

// --- маршруты → каталог и OSRM ---

type routePlaces struct{ catalog catalog.Service }

func toRoutePlaces(list []catalog.Place) []routesvc.Place {
	out := make([]routesvc.Place, len(list))
	for i, p := range list {
		out[i] = routesvc.Place{ID: p.ID, Slug: p.Slug, Title: p.Title, Lat: p.Lat, Lng: p.Lng}
	}
	return out
}

func (a routePlaces) PublishedPlaces(ctx context.Context, slugs []string) ([]routesvc.Place, error) {
	list, err := a.catalog.PublishedPlaces(ctx, slugs)
	return toRoutePlaces(list), err
}

func (a routePlaces) PlacesByIDs(ctx context.Context, ids []int64) ([]routesvc.Place, error) {
	list, err := a.catalog.PlacesByIDs(ctx, ids)
	return toRoutePlaces(list), err
}

type routeRouter struct{ osrm *osrm.Client }

func (a routeRouter) Route(ctx context.Context, points []routesvc.Point) (routesvc.RouteLine, error) {
	pts := make([]osrm.Point, len(points))
	for i, p := range points {
		pts[i] = osrm.Point{Lat: p.Lat, Lng: p.Lng}
	}
	r, err := a.osrm.Route(ctx, pts)
	return routesvc.RouteLine{DistanceM: r.DistanceM, DurationS: r.DurationS, Geometry: r.Geometry}, err
}

// --- заявки → каталог, медиафайлы, пользователи ---

type submissionCatalog struct{ catalog catalog.Service }

func (a submissionCatalog) ObjectBySlug(ctx context.Context, slug string) (submissionsvc.ObjectRef, error) {
	// Дополнять можно только опубликованный объект: смотрим глазами посетителя.
	o, err := a.catalog.GetVisible(ctx, slug, catalog.Viewer{})
	if errors.Is(err, catalog.ErrObjectNotFound) {
		return submissionsvc.ObjectRef{}, submission.ErrObjectNotFound
	}
	if err != nil {
		return submissionsvc.ObjectRef{}, err
	}
	return submissionsvc.ObjectRef{ID: o.ID, Slug: o.Slug, Title: o.Title, Published: o.IsPublished}, nil
}

func (a submissionCatalog) CreateObject(ctx context.Context, in submissionsvc.NewObject) (submissionsvc.ObjectRef, error) {
	published := in.IsPublished
	o, err := a.catalog.Create(ctx, catalog.ObjectInput{
		Slug:               in.Slug,
		Title:              in.Title,
		Category:           in.Category,
		ProtectionStatus:   catalog.ProtectionStatus(in.ProtectionStatus),
		RegistryNumber:     in.RegistryNumber,
		YearBuilt:          in.YearBuilt,
		Architect:          in.Architect,
		ArchitecturalStyle: in.ArchitecturalStyle,
		Address:            in.Address,
		Coordinates:        catalog.Coordinates{Lat: in.Lat, Lng: in.Lng},
		ShortDescription:   in.ShortDescription,
		FullHistory:        in.FullHistory,
		Tags:               in.Tags,
		IsPublished:        &published,
	})
	if err != nil {
		return submissionsvc.ObjectRef{}, err
	}
	return submissionsvc.ObjectRef{ID: o.ID, Slug: o.Slug, Title: o.Title, Published: o.IsPublished}, nil
}

func (a submissionCatalog) ObjectsByIDs(ctx context.Context, ids []int64) ([]submissionsvc.ObjectRef, error) {
	list, err := a.catalog.PlacesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]submissionsvc.ObjectRef, len(list))
	for i, p := range list {
		out[i] = submissionsvc.ObjectRef{ID: p.ID, Slug: p.Slug, Title: p.Title}
	}
	return out, nil
}

type submissionMedia struct{ media media.Service }

func toSubmissionFile(f media.File) submission.File {
	return submission.File{
		ID:           f.ID,
		ObjectID:     f.ObjectID,
		SubmissionID: f.SubmissionID,
		Kind:         string(f.Kind),
		OriginalName: f.OriginalName,
		ContentType:  f.ContentType,
		SizeBytes:    f.SizeBytes,
		Caption:      f.Caption,
		YearTaken:    f.YearTaken,
		SourceNote:   f.SourceNote,
		URL:          fmt.Sprintf("/api/media/%d/file", f.ID),
		CreatedAt:    f.CreatedAt,
	}
}

func (a submissionMedia) Upload(ctx context.Context, in submissionsvc.MediaUpload) (submissionsvc.MediaFile, error) {
	f, err := a.media.Upload(ctx, media.UploadInput{
		SubmissionID: in.SubmissionID,
		Kind:         media.Kind(in.Kind),
		Caption:      in.Caption,
		YearTaken:    in.YearTaken,
		SourceNote:   in.SourceNote,
		OriginalName: in.OriginalName,
		Size:         in.Size,
		Body:         in.Body,
		UploadedBy:   in.UploadedBy,
	})
	if err != nil {
		return submissionsvc.MediaFile{}, err
	}
	return toSubmissionFile(f), nil
}

func (a submissionMedia) ListBySubmission(ctx context.Context, id int64) ([]submissionsvc.MediaFile, error) {
	list, err := a.media.ListBySubmission(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]submissionsvc.MediaFile, len(list))
	for i, f := range list {
		out[i] = toSubmissionFile(f)
	}
	return out, nil
}

func (a submissionMedia) AttachToObject(ctx context.Context, submissionID, objectID int64) error {
	return a.media.AttachToObject(ctx, submissionID, objectID)
}

type submissionUsers struct{ auth auth.Service }

func (a submissionUsers) User(ctx context.Context, id int64) (submissionsvc.Author, error) {
	u, err := a.auth.GetByID(ctx, id)
	if err != nil {
		return submissionsvc.Author{}, err
	}
	return submissionsvc.Author{ID: u.ID, Name: u.Name, Email: u.Email, Role: string(u.Role), IsBlocked: u.IsBlocked}, nil
}

func (a submissionUsers) EnsureActive(ctx context.Context, id int64) error {
	return a.auth.EnsureActive(ctx, id)
}

// Block вызывается внутри транзакции заявок: репозиторий пользователей берёт её из ctx.
func (a submissionUsers) Block(ctx context.Context, actorID int64, actorRole string, userID int64) error {
	_, err := a.auth.SetBlocked(ctx, auth.Principal{UserID: actorID, Role: auth.Role(actorRole)}, userID, true)
	return err
}

// --- уведомления → пользователи ---

type notificationUsers struct{ auth auth.Service }

// Contact передаёт ошибку auth как есть: «не найден» модуль уведомлений распознаёт по роду ошибки.
func (a notificationUsers) Contact(ctx context.Context, id int64) (string, string, error) {
	u, err := a.auth.GetByID(ctx, id)
	return u.Email, u.Name, err
}

// --- магазин → ЮKassa ---

type shopPayments struct{ client *yookassa.Client }

func mapYooKassaErr(err error) error {
	switch {
	case errors.Is(err, yookassa.ErrPaymentsDisabled):
		return shopsvc.ErrGatewayDisabled
	case errors.Is(err, yookassa.ErrPaymentNotFound):
		return shopsvc.ErrGatewayPaymentNotFound
	}
	return err
}

func (a shopPayments) CreatePayment(ctx context.Context, orderID, amount int64, desc, returnURL, key string) (string, string, error) {
	p, err := a.client.CreatePayment(ctx, yookassa.PaymentRequest{
		OrderID: orderID, AmountKopecks: amount, Description: desc, ReturnURL: returnURL, IdempotenceKey: key,
	})
	return p.ID, p.ConfirmationURL, mapYooKassaErr(err)
}

func (a shopPayments) GetPayment(ctx context.Context, id string) (shopsvc.Payment, error) {
	p, err := a.client.GetPayment(ctx, id)
	return shopsvc.Payment{ID: p.ID, Status: p.Status, Amount: p.Amount, OrderID: p.OrderID}, mapYooKassaErr(err)
}
