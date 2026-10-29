-- +goose Up
-- Медиафайлы: панорамы и сведения о происхождении изображения.

-- Новый вид медиа — панорама (сферическая 360° или частичная).
ALTER TABLE media_files DROP CONSTRAINT media_files_kind_check;
ALTER TABLE media_files ADD CONSTRAINT media_files_kind_check
    CHECK (kind IN ('photo', 'archive_photo', 'model3d', 'audio', 'panorama'));

-- Прежний свободный текст «откуда материал» (архив, книга) переименовываем:
-- имя source теперь означает, что это за изображение — фото или реконструкция ИИ.
ALTER TABLE media_files RENAME COLUMN source TO source_note;

ALTER TABLE media_files
    -- Проекция панорамы: full_360 — полная сфера, partial — часть круга.
    ADD COLUMN projection TEXT CHECK (projection IN ('full_360', 'partial')),
    -- Горизонтальный угол обзора частичной панорамы в градусах.
    ADD COLUMN h_angle    INT CHECK (h_angle BETWEEN 1 AND 360),
    -- photo — настоящая фотография, ai_reconstruction — изображение, созданное нейросетью;
    -- сайт подписывает такие изображения «реконструкция ИИ».
    ADD COLUMN source     TEXT NOT NULL DEFAULT 'photo' CHECK (source IN ('photo', 'ai_reconstruction')),
    ADD COLUMN author     TEXT NOT NULL DEFAULT '',
    ADD COLUMN license    TEXT NOT NULL DEFAULT '',
    -- У панорамы проекция обязательна, у остальных видов её нет.
    ADD CONSTRAINT media_files_panorama_projection_check
        CHECK ((kind = 'panorama') = (projection IS NOT NULL)),
    -- Угол обзора бывает только у частичной панорамы.
    ADD CONSTRAINT media_files_h_angle_partial_check
        CHECK (h_angle IS NULL OR projection = 'partial');

-- +goose Down
-- Панорамы в старой схеме не допускаются, поэтому их сведения удаляются
-- (сами файлы в хранилище остаются).
DELETE FROM media_files WHERE kind = 'panorama';
ALTER TABLE media_files
    DROP CONSTRAINT media_files_h_angle_partial_check,
    DROP CONSTRAINT media_files_panorama_projection_check,
    DROP COLUMN license,
    DROP COLUMN author,
    DROP COLUMN source,
    DROP COLUMN h_angle,
    DROP COLUMN projection;
ALTER TABLE media_files RENAME COLUMN source_note TO source;
ALTER TABLE media_files DROP CONSTRAINT media_files_kind_check;
ALTER TABLE media_files ADD CONSTRAINT media_files_kind_check
    CHECK (kind IN ('photo', 'archive_photo', 'model3d', 'audio'));
