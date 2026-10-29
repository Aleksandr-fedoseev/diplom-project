package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/domain/catalog"
	"heritage-backend/internal/repository/postgres/sqlc"
)

// Objects — объекты наследия. Координаты хранятся в PostGIS (geography),
// поиск рядом с точкой и в области карты идёт по GIST-индексу.
type Objects struct {
	pool *pgxpool.Pool
}

var _ catalog.ObjectRepository = (*Objects)(nil)

func NewObjects(pool *pgxpool.Pool) *Objects { return &Objects{pool: pool} }

// objectRow — поля строки списка объектов. Совпадает по полям с sqlc.ListObjectsRow,
// поэтому строку списка можно просто привести к нему: objectRow(row).
// sqlc создаёт на каждый запрос свой тип, а сущность домена собирается одинаково.
type objectRow struct {
	ID                 int64
	Slug               string
	Title              string
	Category           string
	ProtectionStatus   string
	RegistryNumber     *string
	YearBuilt          *string
	Architect          *string
	ArchitecturalStyle *string
	Address            string
	Lat, Lng           float64
	ShortDescription   string
	Tags               []string
	IsPublished        bool
	UpdatedAt          time.Time
}

func (r objectRow) toObject() catalog.Object {
	return catalog.Object{
		ID:                 r.ID,
		Slug:               r.Slug,
		Title:              r.Title,
		Category:           r.Category,
		ProtectionStatus:   catalog.ProtectionStatus(r.ProtectionStatus),
		RegistryNumber:     r.RegistryNumber,
		YearBuilt:          r.YearBuilt,
		Architect:          r.Architect,
		ArchitecturalStyle: r.ArchitecturalStyle,
		Address:            r.Address,
		Coordinates:        catalog.Coordinates{Lat: r.Lat, Lng: r.Lng},
		ShortDescription:   r.ShortDescription,
		Tags:               r.Tags,
		IsPublished:        r.IsPublished,
		UpdatedAt:          r.UpdatedAt,
	}
}

// optional: пустая строка в фильтре означает «не фильтровать» — в запрос уходит NULL.
func optional(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func (r *Objects) List(ctx context.Context, f catalog.Filter) ([]catalog.Object, error) {
	var minLng, minLat, maxLng, maxLat *float64
	if b := f.BBox; b != nil {
		minLng, minLat, maxLng, maxLat = &b[0], &b[1], &b[2], &b[3]
	}
	q := queries(ctx, r.pool)

	if f.Near == nil {
		rows, err := q.ListObjects(ctx, sqlc.ListObjectsParams{
			IncludeUnpublished: f.IncludeUnpublished,
			Category:           optional(f.Category),
			Status:             optional(f.Status),
			Query:              optional(f.Query),
			MinLng:             minLng, MinLat: minLat, MaxLng: maxLng, MaxLat: maxLat,
		})
		if err != nil {
			return nil, err
		}
		list := make([]catalog.Object, len(rows))
		for i, row := range rows {
			list[i] = objectRow(row).toObject()
		}
		return list, nil
	}

	// Поиск рядом с точкой: только объекты в радиусе, ближайшие первыми.
	rows, err := q.ListObjectsNear(ctx, sqlc.ListObjectsNearParams{
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
	list := make([]catalog.Object, len(rows))
	for i, row := range rows {
		list[i] = objectRow{
			ID: row.ID, Slug: row.Slug, Title: row.Title, Category: row.Category, ProtectionStatus: row.ProtectionStatus,
			RegistryNumber: row.RegistryNumber, YearBuilt: row.YearBuilt, Architect: row.Architect,
			ArchitecturalStyle: row.ArchitecturalStyle, Address: row.Address, Lat: row.Lat, Lng: row.Lng,
			ShortDescription: row.ShortDescription, Tags: row.Tags, IsPublished: row.IsPublished, UpdatedAt: row.UpdatedAt,
		}.toObject()
		list[i].DistanceM = &row.DistanceM
	}
	return list, nil
}

func (r *Objects) BySlug(ctx context.Context, slug string) (catalog.Object, error) {
	row, err := queries(ctx, r.pool).GetObjectBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return catalog.Object{}, catalog.ErrObjectNotFound
	}
	if err != nil {
		return catalog.Object{}, err
	}
	o := objectRow{
		ID: row.ID, Slug: row.Slug, Title: row.Title, Category: row.Category, ProtectionStatus: row.ProtectionStatus,
		RegistryNumber: row.RegistryNumber, YearBuilt: row.YearBuilt, Architect: row.Architect,
		ArchitecturalStyle: row.ArchitecturalStyle, Address: row.Address, Lat: row.Lat, Lng: row.Lng,
		ShortDescription: row.ShortDescription, Tags: row.Tags, IsPublished: row.IsPublished, UpdatedAt: row.UpdatedAt,
	}.toObject()
	o.FullHistory = row.FullHistory
	return o, nil
}

// Create вставляет объект и перечитывает его через то же соединение из ctx:
// внутри транзакции вызывающего новый объект виден только ей.
func (r *Objects) Create(ctx context.Context, in catalog.ObjectInput) (catalog.Object, error) {
	err := queries(ctx, r.pool).CreateObject(ctx, sqlc.CreateObjectParams{
		Slug:               in.Slug,
		Title:              in.Title,
		Category:           in.Category,
		ProtectionStatus:   string(in.ProtectionStatus),
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
		IsPublished:        published(in),
	})
	if err != nil {
		return catalog.Object{}, objectWriteError(err)
	}
	return r.BySlug(ctx, in.Slug)
}

func (r *Objects) Update(ctx context.Context, id int64, in catalog.ObjectInput) (catalog.Object, error) {
	n, err := queries(ctx, r.pool).UpdateObject(ctx, sqlc.UpdateObjectParams{
		ID:                 id,
		Slug:               in.Slug,
		Title:              in.Title,
		Category:           in.Category,
		ProtectionStatus:   string(in.ProtectionStatus),
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
		IsPublished:        published(in),
	})
	if err != nil {
		return catalog.Object{}, objectWriteError(err)
	}
	if n == 0 {
		return catalog.Object{}, catalog.ErrObjectNotFound
	}
	return r.BySlug(ctx, in.Slug)
}

// published: nil в IsPublished означает «опубликовать» (сервис уже подставил значение, это страховка).
func published(in catalog.ObjectInput) bool { return in.IsPublished == nil || *in.IsPublished }

// objectWriteError переводит нарушения ограничений БД в ошибки домена.
func objectWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case codeUniqueViolation:
		return catalog.ErrSlugTaken
	case codeNotNullViolation:
		// Категория ищется подзапросом по slug; не нашлась — в category_id попадает NULL.
		if pgErr.ColumnName == "category_id" {
			return catalog.ErrUnknownCategory
		}
	}
	return err
}

func (r *Objects) Delete(ctx context.Context, id int64) error {
	n, err := queries(ctx, r.pool).DeleteObject(ctx, id)
	if pgCode(err) == codeForeignKeyViolation { // на объект ссылается остановка маршрута
		return catalog.ErrObjectInRoute
	}
	if err != nil {
		return err
	}
	if n == 0 {
		return catalog.ErrObjectNotFound
	}
	return nil
}

func (r *Objects) Exists(ctx context.Context, id int64) (bool, error) {
	return queries(ctx, r.pool).ObjectExists(ctx, id)
}

func (r *Objects) PublishedPlaces(ctx context.Context, slugs []string) ([]catalog.Place, error) {
	rows, err := queries(ctx, r.pool).ListPublishedPlacesBySlugs(ctx, slugs)
	if err != nil {
		return nil, err
	}
	places := make([]catalog.Place, len(rows))
	for i, p := range rows {
		places[i] = catalog.Place{ID: p.ID, Slug: p.Slug, Title: p.Title, Lat: p.Lat, Lng: p.Lng}
	}
	return places, nil
}

func (r *Objects) PlacesByIDs(ctx context.Context, ids []int64) ([]catalog.Place, error) {
	rows, err := queries(ctx, r.pool).ListPlacesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	places := make([]catalog.Place, len(rows))
	for i, p := range rows {
		places[i] = catalog.Place{ID: p.ID, Slug: p.Slug, Title: p.Title, Lat: p.Lat, Lng: p.Lng}
	}
	return places, nil
}

// Categories — категории объектов.
type Categories struct {
	pool *pgxpool.Pool
}

var _ catalog.CategoryRepository = (*Categories)(nil)

func NewCategories(pool *pgxpool.Pool) *Categories { return &Categories{pool: pool} }

func toCategory(c sqlc.Category) catalog.Category {
	return catalog.Category{ID: c.ID, Slug: c.Slug, Name: c.Name}
}

// categoryError переводит ошибки БД: нет строки, дубликат slug, в категории есть объекты.
func categoryError(err error) error {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return catalog.ErrCategoryNotFound
	case pgCode(err) == codeUniqueViolation:
		return catalog.ErrCategorySlugTaken
	case pgCode(err) == codeForeignKeyViolation:
		return catalog.ErrCategoryNotEmpty
	}
	return err
}

func (r *Categories) List(ctx context.Context) ([]catalog.Category, error) {
	rows, err := queries(ctx, r.pool).ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	cats := make([]catalog.Category, len(rows))
	for i, c := range rows {
		cats[i] = toCategory(c)
	}
	return cats, nil
}

func (r *Categories) Create(ctx context.Context, in catalog.CategoryInput) (catalog.Category, error) {
	c, err := queries(ctx, r.pool).CreateCategory(ctx, sqlc.CreateCategoryParams{Slug: in.Slug, Name: in.Name})
	if err != nil {
		return catalog.Category{}, categoryError(err)
	}
	return toCategory(c), nil
}

func (r *Categories) Update(ctx context.Context, id int32, in catalog.CategoryInput) (catalog.Category, error) {
	c, err := queries(ctx, r.pool).UpdateCategory(ctx, sqlc.UpdateCategoryParams{ID: id, Slug: in.Slug, Name: in.Name})
	if err != nil {
		return catalog.Category{}, categoryError(err)
	}
	return toCategory(c), nil
}

func (r *Categories) Delete(ctx context.Context, id int32) error {
	n, err := queries(ctx, r.pool).DeleteCategory(ctx, id)
	if err != nil {
		return categoryError(err)
	}
	if n == 0 {
		return catalog.ErrCategoryNotFound
	}
	return nil
}
