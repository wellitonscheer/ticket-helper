CREATE TABLE IF NOT EXISTS ticket_chunks (
    id bigserial PRIMARY KEY NOT NULL,
    type character varying,
    ticket_id integer,
    subject character varying,
    ordem integer,
    poster character varying,
    chunk character varying,
    embedding vector(1024)
);
