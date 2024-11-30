CREATE TYPE status_account AS ENUM ('ACTIVE', 'BLOCKED');
CREATE TYPE status_transaction AS ENUM ('BLOCKED', 'CONFIRMED', 'CANCELLED');

CREATE TABLE "atms"
(
    "id"        BIGSERIAL PRIMARY KEY,
    "cashCents" BIGINT             NOT NULL CHECK ( "cashCents" >= 0 ),
    "login"     VARCHAR(32) UNIQUE NOT NULL CHECK ( login ~ '^[a-z0-9_-]+$'),
    "password"  BYTEA              NOT NULL CHECK ( length(password) <= 60 )
);

CREATE TABLE "accountOwners"
(
    "id"     BIGSERIAL PRIMARY KEY,
    "userId" BIGINT UNIQUE,
    "atmId"  BIGINT UNIQUE REFERENCES "atms" ("id"),
    CHECK (("userId" IS NULL AND "atmId" IS NOT NULL)
        OR ("userId" IS NOT NULL AND "atmId" IS NULL))
);

CREATE TABLE "accounts"
(
    "id"           BIGSERIAL PRIMARY KEY,
    "balanceCents" BIGINT         NOT NULL CHECK ( "balanceCents" >= 0 ) DEFAULT 0,
    "ownerId"      BIGINT         NOT NULL REFERENCES "accountOwners" ("id"),
    "status"       status_account NOT NULL                               DEFAULT 'ACTIVE'
);

CREATE TABLE "transactions"
(
    "id"          BIGSERIAL             NOT NULL PRIMARY KEY,
    "senderId"    BIGINT             NOT NULL REFERENCES "accounts" ("id"),
    "receiverId"  BIGINT             NOT NULL REFERENCES "accounts" ("id"),
    "status"      status_transaction NOT NULL DEFAULT 'BLOCKED',
    "createdAt"   TIMESTAMP          NOT NULL DEFAULT current_timestamp,
    "amountCents" BIGINT             NOT NULL,
    "description" TEXT
);

CREATE TABLE "cashOperations"
(
    "id"          BIGSERIAL NOT NULL PRIMARY KEY,
    "atmAccountId"       BIGINT NOT NULL REFERENCES "atms" ("id"),
    "userAccountId"      BIGINT,
    "amountCents" BIGINT NOT NULL CHECK ( "amountCents" != 0 ),
    "createdAt"   TIMESTAMP          NOT NULL DEFAULT current_timestamp
);

CREATE INDEX "transactions_senderId_index" ON "transactions" ("senderId");
CREATE INDEX "transactions_receiverId_index" ON "transactions" ("receiverId");
CREATE INDEX "accounts_ownerId_index" ON "accounts" ("ownerId");

INSERT INTO "atms" ("cashCents", "login", "password")
VALUES (5000000, 'atm001', '$2a$10$GHg/65CqcSqdLAeRcHUtxOYEAiHCKWF8I7WWJPLPv0mF54BkBBbh.'), -- password: atm
       (7500000, 'atm002', '$2a$10$qYNh0MLqcFiDBtYAL28GEOlLVa.sZWs.UhtSORr6iTGG5CuFNmW8S'); -- password: atm2


INSERT INTO "accountOwners" ("userId", "atmId")
VALUES (NULL, 1),
       (NULL, 2),
       (1, NULL);

INSERT INTO "accounts" ("balanceCents", "ownerId", "status")
VALUES (100000, 1, 'ACTIVE'),
       (200000, 2, 'ACTIVE'),
       (300000, 3, 'ACTIVE'),
       (400000, 3, 'BLOCKED');

INSERT INTO "transactions" ("senderId", "receiverId", "status", "amountCents", "description")
VALUES (1, 2, 'CONFIRMED', 50000, 'Payment for services'),
       (2, 3, 'CANCELLED', 100000, 'Refund'),
       (3, 1, 'BLOCKED', 20000, 'Transfer'),
       (1, 3, 'CONFIRMED', 30000, 'Gift');

SELECT *
FROM transactions;
SELECT *
FROM accounts;
SELECT *
FROM "accountOwners";
SELECT *
FROM atms;
SELECT * FROM "cashOperations";
