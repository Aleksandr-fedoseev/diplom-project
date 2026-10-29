package media

import "heritage-backend/internal/domain/errs"

// Ошибки модуля. Сравниваются через errors.Is; transport переводит их род в HTTP-код.
var (
	// Файл, который запрашивающему не положено видеть, тоже «не найден»:
	// так по ответу нельзя узнать, что он существует.
	ErrFileNotFound   = errs.New(errs.NotFound, "файл не найден")
	ErrObjectNotFound = errs.New(errs.NotFound, "объект не найден")

	ErrUnknownKind          = errs.New(errs.Invalid, "kind должен быть одним из: photo, archive_photo, audio, model3d, panorama")
	ErrInvalidSource        = errs.New(errs.Invalid, "source: photo или ai_reconstruction")
	ErrProjectionRequired   = errs.New(errs.Invalid, "для панорамы укажите projection: full_360 или partial")
	ErrProjectionNotAllowed = errs.New(errs.Invalid, "projection указывается только для панорамы")
	ErrInvalidHAngle        = errs.New(errs.Invalid, "hAngle — угол обзора от 1 до 360 градусов, только для частичной панорамы")
)
