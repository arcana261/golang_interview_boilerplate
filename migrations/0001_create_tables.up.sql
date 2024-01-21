
create table if not exists articles (
    id serial primary key not null,
    title varchar(255) not null,
    description varchar(255) not null
)
