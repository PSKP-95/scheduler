CREATE TYPE "status" AS ENUM (
  'pending',
  'running',
  'success',
  'failure'
);

CREATE TABLE IF NOT EXISTS "schedules" (
  "id" uuid PRIMARY KEY,
  "cron" varchar NOT NULL,
  "hook" varchar NOT NULL,
  "owner" varchar DEFAULT '' NOT NULL,
  "active" bool NOT NULL DEFAULT true,
  "till" timestamptz DEFAULT ('0001-01-01 00:00:00Z') NOT NULL,
  "created_at" timestamptz DEFAULT (now()) NOT NULL,
  "last_modified" timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS "next_occurence" (
  "schedule" uuid,
  "worker" uuid,
  "status" status DEFAULT 'pending',
  "occurence" timestamptz NOT NULL,
  "last_updated" timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS "history" (
  "schedule" uuid NOT NULL,
  "status" status NOT NULL,
  "details" text NOT NULL,
  "scehduled_at" timestamptz NOT NULL,
  "started_at" timestamptz NOT NULL,
  "completed_at" timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS "punch_card" (
  "id" uuid PRIMARY KEY,
  "last_punch" timestamptz NOT NULL,
  "created_at" timestamptz DEFAULT (now())
);

CREATE INDEX ON "schedules" ("owner");

CREATE INDEX ON "next_occurence" ("worker");

CREATE INDEX ON "next_occurence" ("occurence");

CREATE INDEX ON "next_occurence" ("schedule");

CREATE INDEX ON "history" ("schedule");

COMMENT ON COLUMN "schedules"."till" IS 'till what timestamp this schedule will run';

ALTER TABLE "next_occurence" ADD FOREIGN KEY ("schedule") REFERENCES "schedules" ("id");

ALTER TABLE "history" ADD FOREIGN KEY ("schedule") REFERENCES "schedules" ("id");

ALTER TABLE "next_occurence" ADD FOREIGN KEY ("worker") REFERENCES "punch_card" ("id") ON DELETE SET NULL;
