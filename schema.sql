CREATE TABLE IF NOT EXISTS templates (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  empty_yaml TEXT,
  file TEXT
);
CREATE TABLE IF NOT EXISTS custom_fields (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  template_id INTEGER NOT NULL,
  key TEXT NOT NULL,
  desc TEXT NOT NULL,
  FOREIGN KEY (template_id)
    REFERENCES templates (id)
);
CREATE TABLE IF NOT EXISTS tab_desc_schema (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  template_id INTEGER NOT NULL,
  value TEXT NOT NULL,
  FOREIGN KEY (template_id)
    REFERENCES templates (id)
);
CREATE TABLE IF NOT EXISTS pdf_name_schema (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  template_id INTEGER NOT NULL,
  value TEXT NOT NULL,
  FOREIGN KEY (template_id)
    REFERENCES templates (id)
);
CREATE TABLE IF NOT EXISTS entries (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  template_id INTEGER NOT NULL,
  data TEXT NOT NULL,
  path TEXT NOT NULL UNIQUE,
  yaml TEXT,
  date INT,
  FOREIGN KEY (template_id)
    REFERENCES templates (id)
);
