create table users
(
    user_id text UNIQUE,
    insert_date timestamp with time zone,
    update_date timestamp with time zone
);
