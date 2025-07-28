CREATE TABLE IF NOT EXISTS black_entries (
    id bigserial PRIMARY KEY NOT NULL,
    content character varying,
    embedding vector(1024)
);
