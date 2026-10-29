// Package migrations хранит SQL-миграции goose и вшивает их в бинарник,
// поэтому файлы не нужно копировать на сервер отдельно.
// Эти же файлы sqlc читает как описание схемы БД (см. sqlc.yaml).
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
