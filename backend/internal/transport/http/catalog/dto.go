package catalog

import (
	"time"

	"heritage-backend/internal/domain/catalog"
	mediahttp "heritage-backend/internal/transport/http/media"
)

// DTO — формат JSON в API; совпадает с прежним, на него опирается фронтенд.

type categoryResponse struct {
	ID   int32  `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func toCategoryResponse(c catalog.Category) categoryResponse {
	return categoryResponse{ID: c.ID, Slug: c.Slug, Name: c.Name}
}

type categoryRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type objectResponse struct {
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
	Coordinates        coordinates `json:"coordinates"`
	ShortDescription   string      `json:"shortDescription"`
	FullHistory        string      `json:"fullHistory,omitempty"` // в списке пусто, есть в карточке
	Tags               []string    `json:"tags"`
	IsPublished        bool        `json:"isPublished"`
	DistanceM          *float64    `json:"distanceM,omitempty"` // только при поиске рядом с точкой
	UpdatedAt          time.Time   `json:"updatedAt"`
}

func toObjectResponse(o catalog.Object) objectResponse {
	tags := o.Tags
	if tags == nil {
		tags = []string{}
	}
	return objectResponse{
		ID:                 o.ID,
		Slug:               o.Slug,
		Title:              o.Title,
		Category:           o.Category,
		ProtectionStatus:   string(o.ProtectionStatus),
		RegistryNumber:     o.RegistryNumber,
		YearBuilt:          o.YearBuilt,
		Architect:          o.Architect,
		ArchitecturalStyle: o.ArchitecturalStyle,
		Address:            o.Address,
		Coordinates:        coordinates{Lat: o.Coordinates.Lat, Lng: o.Coordinates.Lng},
		ShortDescription:   o.ShortDescription,
		FullHistory:        o.FullHistory,
		Tags:               tags,
		IsPublished:        o.IsPublished,
		DistanceM:          o.DistanceM,
		UpdatedAt:          o.UpdatedAt,
	}
}

// cardResponse — карточка объекта: сам объект и его медиафайлы.
type cardResponse struct {
	Object objectResponse           `json:"object"`
	Media  []mediahttp.FileResponse `json:"media"`
}

type objectRequest struct {
	Slug               string      `json:"slug"`
	Title              string      `json:"title"`
	Category           string      `json:"category"`
	ProtectionStatus   string      `json:"protectionStatus"`
	RegistryNumber     *string     `json:"registryNumber"`
	YearBuilt          *string     `json:"yearBuilt"`
	Architect          *string     `json:"architect"`
	ArchitecturalStyle *string     `json:"architecturalStyle"`
	Address            string      `json:"address"`
	Coordinates        coordinates `json:"coordinates"`
	ShortDescription   string      `json:"shortDescription"`
	FullHistory        string      `json:"fullHistory"`
	Tags               []string    `json:"tags"`
	IsPublished        *bool       `json:"isPublished"`
}

func (in objectRequest) toInput() catalog.ObjectInput {
	return catalog.ObjectInput{
		Slug:               in.Slug,
		Title:              in.Title,
		Category:           in.Category,
		ProtectionStatus:   catalog.ProtectionStatus(in.ProtectionStatus),
		RegistryNumber:     in.RegistryNumber,
		YearBuilt:          in.YearBuilt,
		Architect:          in.Architect,
		ArchitecturalStyle: in.ArchitecturalStyle,
		Address:            in.Address,
		Coordinates:        catalog.Coordinates{Lat: in.Coordinates.Lat, Lng: in.Coordinates.Lng},
		ShortDescription:   in.ShortDescription,
		FullHistory:        in.FullHistory,
		Tags:               in.Tags,
		IsPublished:        in.IsPublished,
	}
}
