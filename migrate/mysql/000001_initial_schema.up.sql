CREATE TABLE service (
  id BINARY(16) PRIMARY KEY,
  name VARCHAR(255)
);

CREATE TABLE message (
  service_id BINARY(16),
  language VARCHAR(20),
  FOREIGN KEY (service_id) REFERENCES service (id),
  PRIMARY KEY (service_id, language)
);

CREATE TABLE message_message (
  id BINARY(16) PRIMARY KEY,
  message_service_id BINARY(16) NOT NULL,
  message_language VARCHAR(20) NOT NULL,
  message_id TEXT NOT NULL,
  message TEXT NOT NULL,
  description TEXT,
  fuzzy TINYINT(1),
  FOREIGN KEY (message_service_id, message_language) REFERENCES message (
    service_id, language
  ),
  UNIQUE (message_language, message_id (100))
);
