// Package catalog — объекты наследия и категории: список с фильтрами и геопоиском, карточка, редактирование.
// Модуль — владелец таблиц heritage_objects и categories: другие модули получают
// сведения об объектах через методы Service, а не читают эти таблицы сами.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/catalog/store"
	"heritage-backend/internal/httpx"
)

type Category struct {
	ID   int32  `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Object struct {
	ID                 int64       `json:"id"`
	Slug               string      `json:"slug"`
	Title              string      `json:"title"`
	Category           string      `json:"category"` // slug категории
	ProtectionStatus   string      `json:"protectionStatus"`
	RegistryNumber     *string     `json:"registryNumber,omitempty"`
	YearBuilt          *string     `json:"yearBuilt,omitempty"`
	Architect          *string     `json:"architect,omitempty"`
	ArchitecturalStyle *string     `json:"architecturalStyle,omitempty"`
	Address            string      `json:"address"`
	Coordinates        Coordinates `json:"coordinates"`
	ShortDescription   string      `json:"shortDescription"`
	FullHistory        string      `json:"fullHistory,omitempty"` // в списке не заполняется, есть в карточке
	Tags               []string    `json:"tags"`
	IsPublished        bool        `json:"isPublished"`
	// DistanceM заполняется, только если искали объекты рядом с точкой.
	DistanceM *float64  `json:"distanceM,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Place — минимум сведений об объекте для других модулей (остановка маршрута).
type Place struct {
	ID    int64
	Slug  string
	Title string
	Lat   float64
	Lng   float64
}

type Service struct {
	q *store.Queries // запросы из queries.sql, код сгенерирован sqlc
}

func NewService(db *pgxpool.Pool) *Service { return &Service{q: store.New(db)} }

func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := s.q.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	cats := make([]Category, len(rows))
	for i, c := range rows {
		cats[i] = Category{ID: c.ID, Slug: c.Slug, Name: c.Name}
	}
	return cats, nil
}

// Filter — параметры списка объектов. Все поля необязательны и комбинируются через И.
type Filter struct {
	Category string
	Status   string
	Query    string // поиск по названию и адресу

	// Объекты в радиусе RadiusM метров от точки (Near).
	Near    *Coordinates
	RadiusM float64

	// Объекты в прямоугольнике видимой области карты: [minLng, minLat, maxLng, maxLat].
	BBox *[4]float64

	IncludeUnpublished bool
}

func (s *Service) List(ctx context.Context, f Filter) ([]Object, error) {
	// Пустая строка в фильтре означает «не фильтровать» — в запрос уходит NULL.
	optional := func(v string) *string {
		if v == "" {
			return nil
		}
		return &v
	}
	var minLng, minLat, maxLng, maxLat *float64
	if b := f.BBox; b != nil {
		minLng, minLat, maxLng, maxLat = &b[0], &b[1], &b[2], &b[3]
	}

	if f.Near == nil {
		rows, err := s.q.ListObjects(ctx, store.ListObjectsParams{
			IncludeUnpublished: f.IncludeUnpublished,
			Category:           optional(f.Category),
			Status:             optional(f.Status),
			Query:              optional(f.Query),
			MinLng:             minLng, MinLat: minLat, MaxLng: maxLng, MaxLat: maxLat,
		})
		if err != nil {
			return nil, err
		}
		list := make([]Object, len(rows))
		for i, r := range rows {
			// Строки списка — те же поля, что у карточки, только без полной истории.
			list[i] = objectFromRow(store.GetObjectBySlugRow{
				ID: r.ID, Slug: r.Slug, Title: r.Title, Category: r.Category, ProtectionStatus: r.ProtectionStatus,
				RegistryNumber: r.RegistryNumber, YearBuilt: r.YearBuilt, Architect: r.Architect,
				ArchitecturalStyle: r.ArchitecturalStyle, Address: r.Address, Lat: r.Lat, Lng: r.Lng,
				ShortDescription: r.ShortDescription, Tags: r.Tags, IsPublished: r.IsPublished, UpdatedAt: r.UpdatedAt,
			})
		}
		return list, nil
	}

	rows, err := s.q.ListObjectsNear(ctx, store.ListObjectsNearParams{
		Lng:                f.Near.Lng,
		Lat:                f.Near.Lat,
		RadiusM:            f.RadiusM,
		IncludeUnpublished: f.IncludeUnpublished,
		Category:           optional(f.Category),
		Status:             optional(f.Status),
		Query:              optional(f.Query),
		MinLng:             minLng, MinLat: minLat, MaxLng: maxLng, MaxLat: maxLat,
	})
	if err != nil {
		return nil, err
	}
	list := make([]Object, len(rows))
	for i, r := range rows {
		list[i] = objectFromRow(store.GetObjectBySlugRow{
			ID: r.ID, Slug: r.Slug, Title: r.Title, Category: r.Category, ProtectionStatus: r.ProtectionStatus,
			RegistryNumber: r.RegistryNumber, YearBuilt: r.YearBuilt, Architect: r.Architect,
			ArchitecturalStyle: r.ArchitecturalStyle, Address: r.Address, Lat: r.Lat, Lng: r.Lng,
			ShortDescription: r.ShortDescription, Tags: r.Tags, IsPublished: r.IsPublished, UpdatedAt: r.UpdatedAt,
		})
		list[i].DistanceM = &r.DistanceM
	}
	return list, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Object, error) {
	row, err := s.q.GetObjectBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return Object{}, httpx.NotFound("объект не найден")
	}
	if err != nil {
		return Object{}, err
	}
	return objectFromRow(row), nil
}

func objectFromRow(r store.GetObjectBySlugRow) Object {
	return Object{
		ID:                 r.ID,
		Slug:               r.Slug,
		Title:              r.Title,
		Category:           r.Category,
		ProtectionStatus:   r.ProtectionStatus,
		RegistryNumber:     r.RegistryNumber,
		YearBuilt:          r.YearBuilt,
		Architect:          r.Architect,
		ArchitecturalStyle: r.ArchitecturalStyle,
		Address:            r.Address,
		Coordinates:        Coordinates{Lat: r.Lat, Lng: r.Lng},
		ShortDescription:   r.ShortDescription,
		FullHistory:        r.FullHistory,
		Tags:               r.Tags,
		IsPublished:        r.IsPublished,
		UpdatedAt:          r.UpdatedAt,
	}
}

// ObjectInput — данные для создания и изменения объекта модератором.
type ObjectInput struct {
	Slug               string      `json:"slug"`
	Title              string      `json:"title"`
	Category           string      `json:"category"`
	ProtectionStatus   string      `json:"protectionStatus"`
	RegistryNumber     *string     `json:"registryNumber"`
	YearBuilt          *string     `json:"yearBuilt"`
	Architect          *string     `json:"architect"`
	ArchitecturalStyle *string     `json:"architecturalStyle"`
	Address            string      `json:"address"`
	Coordinates        Coordinates `json:"coordinates"`
	ShortDescription   string      `json:"shortDescription"`
	FullHistory        string      `json:"fullHistory"`
	Tags               []string    `json:"tags"`
	IsPublished        *bool       `json:"isPublished"`
}

var validStatuses = map[string]bool{"federal": true, "regional": true, "identified": true, "lost": true}

func (in *ObjectInput) validate() error {
	in.Slug = strings.TrimSpace(strings.ToLower(in.Slug))
	switch {
	case in.Slug == "" || strings.ContainsAny(in.Slug, " /?#"):
		return httpx.BadRequest("slug обязателен и не должен содержать пробелы и / ? #")
	case strings.TrimSpace(in.Title) == "":
		return httpx.BadRequest("укажите название")
	case strings.TrimSpace(in.Address) == "":
		return httpx.BadRequest("укажите адрес (для утраченного объекта — исторический)")
	case !validStatuses[in.ProtectionStatus]:
		return httpx.BadRequest("protectionStatus: federal, regional, identified или lost")
	case in.Coordinates.Lat < -90 || in.Coordinates.Lat > 90 || in.Coordinates.Lng < -180 || in.Coordinates.Lng > 180:
		return httpx.BadRequest("некорректные координаты")
	}
	if in.Tags == nil {
		in.Tags = []string{}
	}
	if in.IsPublished == nil {
		t := true
		in.IsPublished = &t
	}
	return nil
}

func (s *Service) Create(ctx context.Context, in ObjectInput) (Object, error) {
	return create(ctx, s.q, in)
}

// CreateTx создаёт объект в транзакции вызывающего модуля
// (при одобрении заявки объект, её файлы и решение сохраняются вместе).
func (s *Service) CreateTx(ctx context.Context, tx pgx.Tx, in ObjectInput) (Object, error) {
	return create(ctx, s.q.WithTx(tx), in)
}

func create(ctx context.Context, q *store.Queries, in ObjectInput) (Object, error) {
	if err := in.validate(); err != nil {
		return Object{}, err
	}
	err := q.CreateObject(ctx, store.CreateObjectParams{
		Slug:               in.Slug,
		Title:              in.Title,
		Category:           in.Category,
		ProtectionStatus:   in.ProtectionStatus,
		RegistryNumber:     in.RegistryNumber,
		YearBuilt:          in.YearBuilt,
		Architect:          in.Architect,
		ArchitecturalStyle: in.ArchitecturalStyle,
		Address:            in.Address,
		Lng:                in.Coordinates.Lng,
		Lat:                in.Coordinates.Lat,
		ShortDescription:   in.ShortDescription,
		FullHistory:        in.FullHistory,
		Tags:               in.Tags,
		IsPublished:        *in.IsPublished,
	})
	if err := mapWriteError(err); err != nil {
		return Object{}, err
	}
	row, err := q.GetObjectBySlug(ctx, in.Slug)
	if err != nil {
		return Object{}, err
	}
	return objectFromRow(row), nil
}

func (s *Service) Update(ctx context.Context, id int64, in ObjectInput) (Object, error) {
	if err := in.validate(); err != nil {
		return Object{}, err
	}
	n, err := s.q.UpdateObject(ctx, store.UpdateObjectParams{
		ID:                 id,
		Slug:               in.Slug,
		Title:              in.Title,
		Category:           in.Category,
		ProtectionStatus:   in.ProtectionStatus,
		RegistryNumber:     in.RegistryNumber,
		YearBuilt:          in.YearBuilt,
		Architect:          in.Architect,
		ArchitecturalStyle: in.ArchitecturalStyle,
		Address:            in.Address,
		Lng:                in.Coordinates.Lng,
		Lat:                in.Coordinates.Lat,
		ShortDescription:   in.ShortDescription,
		FullHistory:        in.FullHistory,
		Tags:               in.Tags,
		IsPublished:        *in.IsPublished,
	})
	if err := mapWriteError(err); err != nil {
		return Object{}, err
	}
	if n == 0 {
		return Object{}, httpx.NotFound("объект не найден")
	}
	return s.GetBySlug(ctx, in.Slug)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	n, err := s.q.DeleteObject(ctx, id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
		return httpx.Conflict("объект входит в маршрут — сначала уберите его из маршрутов")
	}
	if err != nil {
		return err
	}
	if n == 0 {
		return httpx.NotFound("объект не найден")
	}
	return nil
}

// mapWriteError превращает ошибки ограничений БД в понятные клиенту ответы.
func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case "23505":
		return httpx.Conflict("объект с таким slug уже есть")
	case "23502": // not_null_violation — сюда попадаем, если категория не нашлась
		if pgErr.ColumnName == "category_id" {
			return httpx.BadRequest("неизвестная категория")
		}
	}
	return err
}

// Методы для других модулей.

// Exists сообщает, есть ли объект с таким id (нужно медиафайлам перед загрузкой).
func (s *Service) Exists(ctx context.Context, id int64) (bool, error) {
	return s.q.ObjectExists(ctx, id)
}

// PublishedPlaces возвращает опубликованные объекты по slug, порядок не гарантируется.
func (s *Service) PublishedPlaces(ctx context.Context, slugs []string) ([]Place, error) {
	rows, err := s.q.PublishedPlacesBySlugs(ctx, slugs)
	if err != nil {
		return nil, fmt.Errorf("объекты по slug: %w", err)
	}
	places := make([]Place, len(rows))
	for i, r := range rows {
		places[i] = Place(r)
	}
	return places, nil
}

// PlacesByIDs возвращает объекты по id, порядок не гарантируется.
func (s *Service) PlacesByIDs(ctx context.Context, ids []int64) ([]Place, error) {
	rows, err := s.q.PlacesByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("объекты по id: %w", err)
	}
	places := make([]Place, len(rows))
	for i, r := range rows {
		places[i] = Place(r)
	}
	return places, nil
}

// Управление категориями (администратор).

type CategoryInput struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func (in *CategoryInput) validate() error {
	in.Slug = strings.TrimSpace(strings.ToLower(in.Slug))
	in.Name = strings.TrimSpace(in.Name)
	if in.Slug == "" || strings.ContainsAny(in.Slug, " /?#") {
		return httpx.BadRequest("slug обязателен и не должен содержать пробелы и / ? #")
	}
	if in.Name == "" {
		return httpx.BadRequest("укажите название категории")
	}
	return nil
}

func (s *Service) CreateCategory(ctx context.Context, in CategoryInput) (Category, error) {
	if err := in.validate(); err != nil {
		return Category{}, err
	}
	c, err := s.q.CreateCategory(ctx, store.CreateCategoryParams{Slug: in.Slug, Name: in.Name})
	if err := mapCategoryError(err); err != nil {
		return Category{}, err
	}
	return Category(c), nil
}

func (s *Service) UpdateCategory(ctx context.Context, id int32, in CategoryInput) (Category, error) {
	if err := in.validate(); err != nil {
		return Category{}, err
	}
	c, err := s.q.UpdateCategory(ctx, store.UpdateCategoryParams{ID: id, Slug: in.Slug, Name: in.Name})
	if errors.Is(err, pgx.ErrNoRows) {
		return Category{}, httpx.NotFound("категория не найдена")
	}
	if err := mapCategoryError(err); err != nil {
		return Category{}, err
	}
	return Category(c), nil
}

func (s *Service) DeleteCategory(ctx context.Context, id int32) error {
	n, err := s.q.DeleteCategory(ctx, id)
	if err := mapCategoryError(err); err != nil {
		return err
	}
	if n == 0 {
		return httpx.NotFound("категория не найдена")
	}
	return nil
}

func mapCategoryError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return httpx.Conflict("категория с таким slug уже есть")
		case "23503":
			return httpx.Conflict("в категории есть объекты — сначала перенесите их в другую")
		}
	}
	return err
}
