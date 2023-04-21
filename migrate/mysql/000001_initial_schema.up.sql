CREATE TABLE service (
  id BINARY(16) PRIMARY KEY,
  name VARCHAR(255)
);

CREATE TABLE message (
  service_id BINARY(16) NOT NULL,
  language VARCHAR(20),
  FOREIGN KEY (service_id) REFERENCES service (id),
  PRIMARY KEY (service_id, language)
);

CREATE TABLE message_message (
  message_service_id BINARY(16) NOT NULL,
  message_language VARCHAR(20) NOT NULL,
  id TEXT,
  message TEXT NOT NULL,
  description TEXT NOT NULL,
  fuzzy TINYINT(1) NOT NULL,
  FOREIGN KEY (message_service_id, message_language) REFERENCES message (
    service_id, language
  ),
  PRIMARY KEY (message_language, id (100))
);
