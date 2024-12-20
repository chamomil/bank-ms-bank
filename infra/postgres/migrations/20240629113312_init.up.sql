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
    "atmId"  BIGINT UNIQUE REFERENCES "atms" ("id") ON DELETE CASCADE,
    CHECK (("userId" IS NULL AND "atmId" IS NOT NULL)
        OR ("userId" IS NOT NULL AND "atmId" IS NULL))
);

CREATE TABLE "accounts"
(
    "id"           BIGSERIAL PRIMARY KEY,
    "balanceCents" BIGINT         NOT NULL CHECK ( "balanceCents" >= 0 ) DEFAULT 0,
    "ownerId"      BIGINT         NOT NULL REFERENCES "accountOwners" ("id") ON DELETE CASCADE,
    "status"       status_account NOT NULL                               DEFAULT 'ACTIVE'
);

CREATE TABLE "transactions"
(
    "id"          BIGSERIAL          NOT NULL PRIMARY KEY,
    "senderId"    BIGINT             NOT NULL REFERENCES "accounts" ("id"),
    "receiverId"  BIGINT             NOT NULL REFERENCES "accounts" ("id"),
    "status"      status_transaction NOT NULL DEFAULT 'BLOCKED',
    "createdAt"   TIMESTAMP          NOT NULL DEFAULT current_timestamp,
    "amountCents" BIGINT             NOT NULL,
    "description" TEXT
);

CREATE TABLE "cashOperations"
(
    "id"            BIGSERIAL NOT NULL PRIMARY KEY,
    "atmAccountId"  BIGINT    NOT NULL REFERENCES "atms" ("id"),
    "userAccountId" BIGINT,
    "amountCents"   BIGINT    NOT NULL CHECK ( "amountCents" != 0 ),
    "createdAt"     TIMESTAMP NOT NULL DEFAULT current_timestamp
);

CREATE INDEX "transactions_senderId_index" ON "transactions" ("senderId");
CREATE INDEX "transactions_receiverId_index" ON "transactions" ("receiverId");
CREATE INDEX "accounts_ownerId_index" ON "accounts" ("ownerId");

