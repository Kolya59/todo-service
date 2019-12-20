create table users
(
    uuid     uuid not null
        constraint users_pk
            primary key,
    login    text not null,
    password bytea not null,
    salt     bytea not null
);

alter table users
    owner to kolya59;

create unique index users_login_uindex
    on users (login);

create unique index users_uuid_uindex
    on users (uuid);

create table tasks
(
    uuid        uuid    not null
        constraint tasks_pk
            primary key,
    value       text,
    author_uuid uuid    not null,
    is_resolved boolean not null
);

alter table tasks
    owner to kolya59;

create unique index tasks_uuid_uindex
    on tasks (uuid);

