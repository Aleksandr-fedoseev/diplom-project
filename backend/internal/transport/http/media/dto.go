package media

import (
	"time"

	"heritage-backend/internal/domain/media"
)

// FileResponse — формат медиафайла в JSON. Экспортирован: его же выводят карточка объекта
// и карточка заявки. Ключ в хранилище и автор загрузки наружу не отдаются.
type FileResponse struct {
	ID           int64     `json:"id"`
	ObjectID     *int64    `json:"objectId,omitempty"`
	SubmissionID *int64    `json:"submissionId,omitempty"`
	Kind         string    `json:"kind"`
	OriginalName string    `json:"originalName"`
	ContentType  string    `json:"contentType"`
	SizeBytes    int64     `json:"sizeBytes"`
	Caption      string    `json:"caption"`
	YearTaken    string    `json:"yearTaken"`
	SourceNote   string    `json:"sourceNote"` // откуда материал: архив, книга
	Source       string    `json:"source"`     // photo или ai_reconstruction («реконструкция ИИ»)
	Projection   *string   `json:"projection,omitempty"`
	HAngle       *int      `json:"hAngle,omitempty"`
	Author       string    `json:"author"`
	License      string    `json:"license"`
	URL          string    `json:"url"`
	CreatedAt    time.Time `json:"createdAt"`
}

func ToFileResponse(f media.File) FileResponse {
	var projection *string
	if f.Projection != nil {
		p := string(*f.Projection)
		projection = &p
	}
	return FileResponse{
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
		Source:       string(f.Source),
		Projection:   projection,
		HAngle:       f.HAngle,
		Author:       f.Author,
		License:      f.License,
		URL:          f.URL(),
		CreatedAt:    f.CreatedAt,
	}
}

// ToFileResponses — список файлов; пустой список выводится как [], а не null.
func ToFileResponses(files []media.File) []FileResponse {
	out := make([]FileResponse, len(files))
	for i, f := range files {
		out[i] = ToFileResponse(f)
	}
	return out
}
