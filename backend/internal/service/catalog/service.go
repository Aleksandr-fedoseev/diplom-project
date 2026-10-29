// Package catalog — сценарии каталога: категории, список объектов с фильтрами и геопоиском,
// карточка, создание и изменение объектов. Реализует catalog.Service из домена.
package catalog

import (
	"context"

	"heritage-backend/internal/domain/catalog"
)

type Service struct {
	objects    catalog.ObjectRepository
	categories catalog.CategoryRepository
}

// Проверка при сборке: Service реализует интерфейс домена.
var _ catalog.Service = (*Service)(nil)

func NewService(objects catalog.ObjectRepository, categories catalog.CategoryRepository) *Service {
	return &Service{objects: objects, categories: categories}
}

// Категории.

func (s *Service) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	return s.categories.List(ctx)
}

func (s *Service) CreateCategory(ctx context.Context, in catalog.CategoryInput) (catalog.Category, error) {
	if err := in.Normalize(); err != nil {
		return catalog.Category{}, err
	}
	return s.categories.Create(ctx, in)
}

func (s *Service) UpdateCategory(ctx context.Context, id int32, in catalog.CategoryInput) (catalog.Category, error) {
	if err := in.Normalize(); err != nil {
		return catalog.Category{}, err
	}
	return s.categories.Update(ctx, id, in)
}

func (s *Service) DeleteCategory(ctx context.Context, id int32) error {
	return s.categories.Delete(ctx, id)
}

// Объекты.

func (s *Service) List(ctx context.Context, f catalog.Filter, v catalog.Viewer) ([]catalog.Object, error) {
	return s.objects.List(ctx, f.ForViewer(v))
}

func (s *Service) GetVisible(ctx context.Context, slug string, v catalog.Viewer) (catalog.Object, error) {
	o, err := s.objects.BySlug(ctx, slug)
	if err != nil {
		return catalog.Object{}, err
	}
	// Скрытый объект для посторонних «не существует»: так по ответу нельзя узнать о его наличии.
	if !o.VisibleTo(v) {
		return catalog.Object{}, catalog.ErrObjectNotFound
	}
	return o, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (catalog.Object, error) {
	return s.objects.BySlug(ctx, slug)
}

// Create вызывается и модулем заявок внутри его транзакции: репозиторий берёт
// транзакцию из ctx, поэтому объект и решение по заявке сохраняются вместе.
func (s *Service) Create(ctx context.Context, in catalog.ObjectInput) (catalog.Object, error) {
	if err := in.Normalize(); err != nil {
		return catalog.Object{}, err
	}
	return s.objects.Create(ctx, in)
}

func (s *Service) Update(ctx context.Context, id int64, in catalog.ObjectInput) (catalog.Object, error) {
	if err := in.Normalize(); err != nil {
		return catalog.Object{}, err
	}
	return s.objects.Update(ctx, id, in)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.objects.Delete(ctx, id)
}

// Методы для других модулей.

func (s *Service) Exists(ctx context.Context, id int64) (bool, error) {
	return s.objects.Exists(ctx, id)
}

func (s *Service) PublishedPlaces(ctx context.Context, slugs []string) ([]catalog.Place, error) {
	return s.objects.PublishedPlaces(ctx, slugs)
}

func (s *Service) PlacesByIDs(ctx context.Context, ids []int64) ([]catalog.Place, error) {
	return s.objects.PlacesByIDs(ctx, ids)
}
