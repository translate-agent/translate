CREATE TABLE service (
  id BINARY(16) PRIMARY KEY,
  name VARCHAR(255)
);

CREATE TABLE translation (
  id BINARY(16) PRIMARY KEY,
  service_id BINARY(16) NOT NULL,
  language VARCHAR(20) NOT NULL,
  original BOOLEAN NOT NULL DEFAULT false,

  FOREIGN KEY (service_id) REFERENCES service (id)
);

CREATE TABLE message (
  translation_id BINARY(16) NOT NULL,
  id TEXT NOT NULL,
  message TEXT NOT NULL,
  description TEXT,
  status ENUM('UNTRANSLATED', 'FUZZY', 'TRANSLATED') NOT NULL,
  positions JSON,

  UNIQUE ((SHA1(id)), translation_id),
  FOREIGN KEY (translation_id) REFERENCES translation (id)
);
