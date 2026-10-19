// Package catalog — объекты наследия и категории: список с фильтрами и геопоиском, карточка, редактирование.
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

	"heritage-backend/internal/httpx"
)

type Category struct {
	ID   int    `json:"id"`
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
	FullHistory        string      `json:"fullHistory,omitempty"`
	Tags               []string    `json:"tags"`
	IsPublished        bool        `json:"isPublished"`
	// DistanceM заполняется, только если искали объекты рядом с точкой.
	DistanceM *float64  `json:"distanceM,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := r.db.Query(ctx, `SELECT id, slug, name FROM categories ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Category])
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

// В SELECT координаты достаём из geography: ST_Y — широта, ST_X — долгота.
const objectColumns = `
	o.id, o.slug, o.title, c.slug, o.protection_status, o.registry_number, o.year_built,
	o.architect, o.architectural_style, o.address,
	ST_Y(o.location::geometry), ST_X(o.location::geometry),
	o.short_description, o.full_history, o.tags, o.is_published, o.updated_at`

func (r *Repository) List(ctx context.Context, f Filter) ([]Object, error) {
	// Запрос собирается из условий: в args копятся значения, $N — их номера.
	var (
		where []string
		args  []any
	)
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	distance := "NULL::float8"
	order := "o.title"

	if !f.IncludeUnpublished {
		where = append(where, "o.is_published")
	}
	if f.Category != "" {
		where = append(where, "c.slug = "+arg(f.Category))
	}
	if f.Status != "" {
		where = append(where, "o.protection_status = "+arg(f.Status))
	}
	if f.Query != "" {
		p := arg("%" + f.Query + "%")
		where = append(where, fmt.Sprintf("(o.title ILIKE %s OR o.address ILIKE %s)", p, p))
	}
	if f.Near != nil {
		point := fmt.Sprintf("ST_SetSRID(ST_MakePoint(%s, %s), 4326)::geography", arg(f.Near.Lng), arg(f.Near.Lat))
		// ST_DWithin использует GIST-индекс, поэтому не перебирает все объекты.
		where = append(where, fmt.Sprintf("ST_DWithin(o.location, %s, %s)", point, arg(f.RadiusM)))
		distance = fmt.Sprintf("ST_Distance(o.location, %s)", point)
		order = "distance_m"
	}
	if f.BBox != nil {
		b := f.BBox
		where = append(where, fmt.Sprintf("o.location && ST_MakeEnvelope(%s, %s, %s, %s, 4326)::geography",
			arg(b[0]), arg(b[1]), arg(b[2]), arg(b[3])))
	}

	sql := "SELECT " + objectColumns + ", " + distance + " AS distance_m" +
		" FROM heritage_objects o JOIN categories c ON c.id = o.category_id"
	if len(where) > 0 {
		sql += " WHERE " + strings.Join(where, " AND ")
	}
	sql += " ORDER BY " + order + " LIMIT 500"

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Object, error) {
		var o Object
		err := row.Scan(objectDest(&o, &o.DistanceM)...)
		o.FullHistory = "" // в списке полная история не нужна, она есть в карточке
		return o, err
	})
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (Object, error) {
	var o Object
	var dist *float64
	err := r.db.QueryRow(ctx,
		"SELECT "+objectColumns+", NULL::float8 FROM heritage_objects o JOIN categories c ON c.id = o.category_id WHERE o.slug = $1",
		slug).Scan(objectDest(&o, &dist)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return Object{}, httpx.NotFound("объект не найден")
	}
	return o, err
}

func objectDest(o *Object, dist **float64) []any {
	return []any{
		&o.ID, &o.Slug, &o.Title, &o.Category, &o.ProtectionStatus, &o.RegistryNumber, &o.YearBuilt,
		&o.Architect, &o.ArchitecturalStyle, &o.Address, &o.Coordinates.Lat, &o.Coordinates.Lng,
		&o.ShortDescription, &o.FullHistory, &o.Tags, &o.IsPublished, &o.UpdatedAt, dist,
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

func (r *Repository) Create(ctx context.Context, in ObjectInput) (Object, error) {
	if err := in.validate(); err != nil {
		return Object{}, err
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO heritage_objects
			(slug, title, category_id, protection_status, registry_number, year_built, architect,
			 architectural_style, address, location, short_description, full_history, tags, is_published)
		VALUES ($1, $2, (SELECT id FROM categories WHERE slug = $3), $4, $5, $6, $7, $8, $9,
			ST_SetSRID(ST_MakePoint($10, $11), 4326)::geography, $12, $13, $14, $15)`,
		in.Slug, in.Title, in.Category, in.ProtectionStatus, in.RegistryNumber, in.YearBuilt, in.Architect,
		in.ArchitecturalStyle, in.Address, in.Coordinates.Lng, in.Coordinates.Lat,
		in.ShortDescription, in.FullHistory, in.Tags, *in.IsPublished)
	if err := mapWriteError(err); err != nil {
		return Object{}, err
	}
	return r.GetBySlug(ctx, in.Slug)
}

func (r *Repository) Update(ctx context.Context, id int64, in ObjectInput) (Object, error) {
	if err := in.validate(); err != nil {
		return Object{}, err
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE heritage_objects SET
			slug = $2, title = $3, category_id = (SELECT id FROM categories WHERE slug = $4),
			protection_status = $5, registry_number = $6, year_built = $7, architect = $8,
			architectural_style = $9, address = $10,
			location = ST_SetSRID(ST_MakePoint($11, $12), 4326)::geography,
			short_description = $13, full_history = $14, tags = $15, is_published = $16,
			updated_at = now()
		WHERE id = $1`,
		id, in.Slug, in.Title, in.Category, in.ProtectionStatus, in.RegistryNumber, in.YearBuilt, in.Architect,
		in.ArchitecturalStyle, in.Address, in.Coordinates.Lng, in.Coordinates.Lat,
		in.ShortDescription, in.FullHistory, in.Tags, *in.IsPublished)
	if err := mapWriteError(err); err != nil {
		return Object{}, err
	}
	if tag.RowsAffected() == 0 {
		return Object{}, httpx.NotFound("объект не найден")
	}
	return r.GetBySlug(ctx, in.Slug)
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM heritage_objects WHERE id = $1`, id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
		return httpx.Conflict("объект входит в маршрут — сначала уберите его из маршрутов")
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
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
