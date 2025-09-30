CREATE TABLE IF NOT EXISTS "job" (
    "id" bigint PRIMARY KEY,
    "status" job_status default "inactive" NOT NULL
);

CREATE TYPE job_status AS ENUM ('todo', 'doing', 'done');