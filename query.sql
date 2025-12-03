-- name: CreateTemplate :one
INSERT INTO templates (name, empty_yaml, file)
VALUES (?, ?, ?)
RETURNING id;

-- name: GetTemplateByName :one
SELECT id, name, empty_yaml, file
FROM templates
WHERE name = ?;

-- name: GetAllTemplates :many
SELECT id, name, empty_yaml, file
FROM templates;

-- name: InsertCustomField :exec
INSERT INTO custom_fields (template_id, key, desc)
VALUES (?, ?, ?);

-- name: GetCustomFieldsByTemplateName :many
SELECT cf.id, cf.template_id, cf.key, cf.desc
FROM custom_fields cf
JOIN templates t ON cf.template_id = t.id
WHERE t.name = ?;

-- name: InsertTabDescSchema :exec
INSERT INTO tab_desc_schema (template_id, value)
VALUES (?, ?);

-- name: GetTabDescriptionsByTemplateID :many
SELECT id, template_id, value
FROM tab_desc_schema
WHERE template_id = ?;

-- name: InsertPdfNameSchema :exec
INSERT INTO pdf_name_schema (template_id, value)
VALUES (?, ?);

-- name: GetPdfNamingByTemplateID :many
SELECT id, template_id, value
FROM pdf_name_schema
WHERE template_id = ?;

-- name: InsertEntry :exec
INSERT INTO entries (template_id, data, path, yaml, date)
VALUES (?, ?, ?, ?, ?);

-- name: GetEntryByPath :one
SELECT id, template_id, data, path, yaml, date
FROM entries
WHERE path = ?;

-- name: GetAllEntries :many
SELECT id, template_id, data, path, yaml, date
FROM entries;

-- name: GetAllEntriesPlusTemplateName :many
SELECT
    entries.id,
    entries.data,
    entries.path,
    entries.yaml,
    entries.date,
    templates.name AS template_name
FROM entries
JOIN templates ON entries.template_id = templates.id
ORDER BY entries.date DESC;

-- name: UpdateYamlByPath :exec
UPDATE entries
SET yaml = ?
WHERE path = ?;

-- name: UpdateDataById :exec
UPDATE entries
SET data = ?
WHERE id = ?;

-- name: UpdateYamlById :exec
UPDATE entries
SET yaml = ?
WHERE id = ?;

-- name: DeleteEntryByPath :exec
DELETE FROM entries
WHERE path = ?;

-- name: DeleteCustomFieldsByTemplateID :exec
DELETE FROM custom_fields
WHERE template_id = ?;

-- name: DeleteTabDescSchemaByTemplateID :exec
DELETE FROM tab_desc_schema
WHERE template_id = ?;

-- name: DeletePdfNameSchemaByTemplateID :exec
DELETE FROM pdf_name_schema
WHERE template_id = ?;

-- name: DeleteEntriesByTemplateID :exec
DELETE FROM entries
WHERE template_id = ?;

-- name: DeleteTemplateByID :exec
DELETE FROM templates
WHERE id = ?;

-- name: DoesPathExist :one
SELECT path
FROM entries
WHERE path = ?;
