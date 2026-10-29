package submission

import (
	"time"

	"heritage-backend/internal/domain/submission"
)

// DTO — формат JSON в API (такой же, как в старом модуле submissions).

type coordinatesDTO struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (c *coordinatesDTO) toDomain() *submission.Coordinates {
	if c == nil {
		return nil
	}
	return &submission.Coordinates{Lat: c.Lat, Lng: c.Lng}
}

// createRequest — заявка пользователя.
type createRequest struct {
	Kind        string          `json:"kind"` // new_object или addition
	ObjectSlug  string          `json:"objectSlug"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Address     string          `json:"address"`
	Coordinates *coordinatesDTO `json:"coordinates"`
	Period      string          `json:"period"`
	Source      string          `json:"source"`
}

func (r createRequest) toDomain() submission.NewSubmission {
	return submission.NewSubmission{
		Kind: submission.Kind(r.Kind), ObjectSlug: r.ObjectSlug, Title: r.Title, Description: r.Description,
		Address: r.Address, Coordinates: r.Coordinates.toDomain(), Period: r.Period, Source: r.Source,
	}
}

// approveRequest — поля объекта каталога (как в POST /api/objects); незаполненные берутся из заявки.
type approveRequest struct {
	Slug               string          `json:"slug"`
	Title              string          `json:"title"`
	Category           string          `json:"category"`
	ProtectionStatus   string          `json:"protectionStatus"`
	RegistryNumber     *string         `json:"registryNumber"`
	YearBuilt          *string         `json:"yearBuilt"`
	Architect          *string         `json:"architect"`
	ArchitecturalStyle *string         `json:"architecturalStyle"`
	Address            string          `json:"address"`
	Coordinates        *coordinatesDTO `json:"coordinates"`
	ShortDescription   string          `json:"shortDescription"`
	FullHistory        string          `json:"fullHistory"`
	Tags               []string        `json:"tags"`
	IsPublished        *bool           `json:"isPublished"`
}

func (r approveRequest) toDomain() submission.ApproveInput {
	return submission.ApproveInput{
		Slug: r.Slug, Title: r.Title, Category: r.Category, ProtectionStatus: r.ProtectionStatus,
		RegistryNumber: r.RegistryNumber, YearBuilt: r.YearBuilt, Architect: r.Architect,
		ArchitecturalStyle: r.ArchitecturalStyle, Address: r.Address, Coordinates: r.Coordinates.toDomain(),
		ShortDescription: r.ShortDescription, FullHistory: r.FullHistory, Tags: r.Tags, IsPublished: r.IsPublished,
	}
}

type rejectRequest struct {
	Reason      string `json:"reason"`
	BlockAuthor bool   `json:"blockAuthor"`
}

type decisionResponse struct {
	Decision    string    `json:"decision"` // approved или rejected
	Reason      string    `json:"reason,omitempty"`
	ModeratorID *int64    `json:"moderatorId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// authorResponse — автор заявки; показывается модератору в карточке.
type authorResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	IsBlocked bool   `json:"isBlocked"`
}

type fileResponse struct {
	ID           int64     `json:"id"`
	ObjectID     *int64    `json:"objectId,omitempty"`
	SubmissionID *int64    `json:"submissionId,omitempty"`
	Kind         string    `json:"kind"`
	OriginalName string    `json:"originalName"`
	ContentType  string    `json:"contentType"`
	SizeBytes    int64     `json:"sizeBytes"`
	Caption      string    `json:"caption"`
	YearTaken    string    `json:"yearTaken"`  // год съёмки
	SourceNote   string    `json:"sourceNote"` // источник: архив, книга, личный архив
	URL          string    `json:"url"`
	CreatedAt    time.Time `json:"createdAt"`
}

func toFileResponse(f submission.File) fileResponse {
	return fileResponse{
		ID: f.ID, ObjectID: f.ObjectID, SubmissionID: f.SubmissionID, Kind: f.Kind, OriginalName: f.OriginalName,
		ContentType: f.ContentType, SizeBytes: f.SizeBytes, Caption: f.Caption, YearTaken: f.YearTaken,
		SourceNote: f.SourceNote, URL: f.URL, CreatedAt: f.CreatedAt,
	}
}

type submissionResponse struct {
	ID          int64             `json:"id"`
	AuthorID    int64             `json:"authorId"`
	Author      *authorResponse   `json:"author,omitempty"` // только в карточке для модератора
	Kind        string            `json:"kind"`
	ObjectID    *int64            `json:"objectId,omitempty"`
	ObjectSlug  string            `json:"objectSlug,omitempty"`
	ObjectTitle string            `json:"objectTitle,omitempty"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Address     string            `json:"address"`
	Coordinates *coordinatesDTO   `json:"coordinates,omitempty"`
	Period      string            `json:"period"`
	Source      string            `json:"source"`
	Status      string            `json:"status"`
	Decision    *decisionResponse `json:"decision,omitempty"`
	Media       []fileResponse    `json:"media,omitempty"` // только в карточке заявки
	CreatedAt   time.Time         `json:"createdAt"`
}

func toSubmissionResponse(s submission.Submission) submissionResponse {
	out := submissionResponse{
		ID: s.ID, AuthorID: s.AuthorID, Kind: string(s.Kind), ObjectID: s.ObjectID, ObjectSlug: s.ObjectSlug,
		ObjectTitle: s.ObjectTitle, Title: s.Title, Description: s.Description, Address: s.Address,
		Period: s.Period, Source: s.Source, Status: string(s.Status), CreatedAt: s.CreatedAt,
	}
	if a := s.Author; a != nil {
		out.Author = &authorResponse{ID: a.ID, Name: a.Name, Email: a.Email, IsBlocked: a.IsBlocked}
	}
	if c := s.Coordinates; c != nil {
		out.Coordinates = &coordinatesDTO{Lat: c.Lat, Lng: c.Lng}
	}
	if d := s.Decision; d != nil {
		out.Decision = &decisionResponse{Decision: string(d.Decision), Reason: d.Reason, ModeratorID: d.ModeratorID, CreatedAt: d.CreatedAt}
	}
	for _, f := range s.Media {
		out.Media = append(out.Media, toFileResponse(f))
	}
	return out
}

// toListResponse — пустой список уходит клиенту как [], а не null.
func toListResponse(list []submission.Submission) []submissionResponse {
	out := make([]submissionResponse, len(list))
	for i, s := range list {
		out[i] = toSubmissionResponse(s)
	}
	return out
}
