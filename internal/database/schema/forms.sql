create type form_status as enum ('draft', 'published');

create table forms (
    id serial primary key,
    form_id text unique,
    name varchar(100) not null,
    description text,
    user_id integer not null,
    status form_status default 'draft',
    schema jsonb not null default '[]',
    settings jsonb not null default '{}', -- UI themes, logic rules
    share_url text unique,
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp
);