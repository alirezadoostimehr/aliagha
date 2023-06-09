DROP DATABASE IF EXISTS aliagha;

CREATE DATABASE aliagha;

CREATE TABLE aliagha.user
(
    id int,
    name varchar(255),
    password varchar(255),
    mobile varchar(16),
    email varchar(255),
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY(id)
);

CREATE TABLE aliagha.passenger (
    id int,
    national_code int(10) UNIQUE,
    name varchar(255),
    birthdate date,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY(id)
);

CREATE TABLE aliagha.user_passenger (
    u_id int,
    p_id int,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (u_id, p_id),
    FOREIGN KEY (u_id) REFERENCES user(id),
    FOREIGN KEY (p_id) REFERENCES passenger(id)
);

CREATE TABLE aliagha.city (
    id int,
    name varchar(255) UNIQUE,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY(id)
);

CREATE TABLE aliagha.airplane (
    id int,
    name varchar(255) UNIQUE,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id)
);

CREATE TABLE aliagha.canceling_situation (
    id int,
    description varchar(255),
    data varchar(255),
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY(id)
);

CREATE TABLE aliagha.flight (
    id int,
    dep_city_id int,
    arr_city_id int,
    dep_time datetime,
    arr_time datetime,
    airplane_id int,
    price int,
    cxl_sit_id int,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (dep_city_id) REFERENCES city(id),
    FOREIGN KEY (arr_city_id) REFERENCES city(id),
    FOREIGN KEY (airplane_id) REFERENCES airplane(id),
    FOREIGN KEY (cxl_sit_id) REFERENCES canceling_situation(id)
);

CREATE TABLE aliagha.ticket (
    id int,
    u_id int,
    p_id int,
    f_id int,
    status int,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (u_id) REFERENCES user(id),
    FOREIGN KEY (p_id) REFERENCES passenger(id),
    FOREIGN KEY (f_id) REFERENCES flight(id)
);

CREATE TABLE aliagha.cancelling (
    id int,
    t_id int,
    description varchar(255),
    cost int,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (t_id) REFERENCES ticket(id)
);

# CREATE TABLE payment (
#     id int,
#     u_id int,
#     type bool,
#     usage_id int,
#     created_at datetime DEFAULT NOW(),
#     updated_at datetime DEFAULT NOW() ON UPDATE NOW(),
#
#     PRIMARY KEY (id),
#     FOREIGN KEY (u_id) REFERENCES user(id),
#     FOREIGN KEY (usage_id) REFERENCES
#     (CASE type
#       WHEN false THEN ticket(id)
#       WHEN true THEN cancelling(id)
#     END)
# );