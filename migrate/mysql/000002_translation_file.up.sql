CREATE TABLE translate_file (
  id BINARY(16) PRIMARY KEY,
  service_id BINARY(16),
  language VARCHAR(20),
  messages JSON,
  FOREIGN KEY (service_id) REFERENCES service (id),
  UNIQUE INDEX idx_service_language (service_id, language)
);
