CREATE TABLE responses (
    id serial primary key,
    form_id integer not null,
    data jsonb not null,
    meta jsonb not null,
    created_at timestamp with time zone default current_timestamp
);