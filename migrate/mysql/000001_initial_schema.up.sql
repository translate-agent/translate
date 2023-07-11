CREATE TABLE service (
  id BINARY(16) PRIMARY KEY,
  name VARCHAR(255)
);

CREATE TABLE message (
  id BINARY(16) PRIMARY KEY,
  service_id BINARY(16) NOT NULL,
  language VARCHAR(20) NOT NULL,
  original BOOLEAN NOT NULL DEFAULT false,

  FOREIGN KEY (service_id) REFERENCES service (id)
);

CREATE TABLE message_message (
  message_id BINARY(16) NOT NULL,
  id TEXT NOT NULL,
  message TEXT NOT NULL,
  description TEXT,
  fuzzy TINYINT(1),

  UNIQUE ((SHA1(id)), message_id),
  FOREIGN KEY (message_id) REFERENCES message (id)
);
