CREATE TABLE service (
  id BINARY(16) PRIMARY KEY,
  name VARCHAR(255)
);

CREATE TABLE messages (
  id BINARY(16) PRIMARY KEY,
  service_id BINARY(16) NOT NULL,
  language VARCHAR(20),
  FOREIGN KEY (service_id) REFERENCES service (id)
);

CREATE TABLE message (
  service_id BINARY(16) NOT NULL,
  id VARCHAR(255),
  message TEXT NOT NULL,
  description TEXT NOT NULL,
  fuzzy TINYINT(1) NOT NULL,
  PRIMARY KEY (service_id, id),
  FOREIGN KEY (service_id) REFERENCES messages (service_id)
);
