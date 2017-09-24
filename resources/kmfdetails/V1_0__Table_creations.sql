-- noinspection SqlDialectInspectionForFile
CREATE TABLE dairy (
  id       BIGSERIAL   NOT NULL,
  dairy_id VARCHAR(50) NOT NULL,
  CONSTRAINT unique_key UNIQUE (id),
  CONSTRAINT unique_dairy_id UNIQUE (dairy_id)
);

CREATE TABLE persons (
  id           BIGSERIAL                    NOT NULL,
  dairy_ref    BIGINT REFERENCES dairy (id) NOT NULL,
  person_id    VARCHAR(50)                  NOT NULL,
  first_name   VARCHAR(25)                  NOT NULL,
  last_name    VARCHAR(25)                  NOT NULL,
  last_updated TIMESTAMP                    NOT NULL,
  CONSTRAINT unique_id UNIQUE (id),
  CONSTRAINT unique_dairy_id_person_id UNIQUE (dairy_ref, person_id)
);

CREATE TABLE address (
  id           BIGSERIAL                      NOT NULL,
  person_ref   BIGINT REFERENCES persons (id) NOT NULL,
  phone_number BIGINT                         NOT NULL,
  full_address VARCHAR(500)                   NOT NULL,
  CONSTRAINT unique_per_person UNIQUE (person_ref)
);

CREATE TABLE transactions (
  id                      BIGSERIAL                                    NOT NULL,
  dairy_ref               BIGINT REFERENCES dairy (id)                 NOT NULL,
  person_ref              BIGINT REFERENCES persons (id)               NOT NULL,
  number_of_liters        INT                                          NOT NULL,
  amount                  INT                                          NOT NULL,
  remaining_total         INT                                          NOT NULL,
  day                     TIMESTAMP                                    NOT NULL,
  person_name             VARCHAR(25)                                  NOT NULL,
  transaction_type        VARCHAR(10)                                  NOT NULL,
  CONSTRAINT unique_dairy_id_person_id_transaction UNIQUE (dairy_ref, person_ref,id)
);

CREATE TABLE total_balance (
  id           BIGSERIAL                      NOT NULL,
  dairy_ref    BIGINT REFERENCES dairy (id)   NOT NULL,
  person_ref   BIGINT REFERENCES persons (id) NOT NULL,
  amount       BIGINT                         NOT NULL,
  last_updated TIMESTAMP                      NOT NULL,
  CONSTRAINT unique_dairy_id_person_id_balance UNIQUE (dairy_ref, person_ref)
);
