DROP DATABASE IF EXISTS aliagha;

CREATE DATABASE aliagha;

CREATE TABLE aliagha.user
(
    id int PRIMARY KEY AUTO_INCREMENT,
    name varchar(255) NOT NULL ,
    password varchar(255) NOT NULL ,
    mobile varchar(16) NOT NULL ,
    email varchar(255) NOT NULL ,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW()
);

CREATE TABLE aliagha.passenger (
    id int PRIMARY KEY AUTO_INCREMENT,
    national_code int(10) UNIQUE NOT NULL ,
    name varchar(255) NOT NULL ,
    birthdate date NOT NULL ,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW()
);

CREATE TABLE aliagha.user_passenger (
    u_id int NOT NULL ,
    p_id int NOT NULL ,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (u_id, p_id),
    FOREIGN KEY (u_id) REFERENCES user(id),
    FOREIGN KEY (p_id) REFERENCES passenger(id)
);

CREATE TABLE aliagha.city (
    id int PRIMARY KEY AUTO_INCREMENT,
    name varchar(255) UNIQUE NOT NULL ,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW()
);

CREATE TABLE aliagha.airplane (
    id int PRIMARY KEY AUTO_INCREMENT,
    name varchar(255) UNIQUE NOT NULL ,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW()
);

CREATE TABLE aliagha.canceling_situation (
    id int PRIMARY KEY AUTO_INCREMENT,
    description varchar(255) NOT NULL ,
    data varchar(255) NOT NULL ,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW()
);

CREATE TABLE aliagha.flight (
    id int PRIMARY KEY AUTO_INCREMENT,
    dep_city_id int NOT NULL ,
    arr_city_id int NOT NULL ,
    dep_time datetime NOT NULL ,
    arr_time datetime NOT NULL ,
    airplane_id int NOT NULL ,
    price int NOT NULL ,
    cxl_sit_id int NOT NULL ,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    FOREIGN KEY (dep_city_id) REFERENCES city(id),
    FOREIGN KEY (arr_city_id) REFERENCES city(id),
    FOREIGN KEY (airplane_id) REFERENCES airplane(id),
    FOREIGN KEY (cxl_sit_id) REFERENCES canceling_situation(id)
);

CREATE TABLE aliagha.ticket (
    id int PRIMARY KEY AUTO_INCREMENT,
    u_id int NOT NULL ,
    p_id int NOT NULL ,
    f_id int NOT NULL ,
    status int NOT NULL ,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    FOREIGN KEY (u_id) REFERENCES user(id),
    FOREIGN KEY (p_id) REFERENCES passenger(id),
    FOREIGN KEY (f_id) REFERENCES flight(id)
);

CREATE TABLE aliagha.cancelling (
    id int PRIMARY KEY AUTO_INCREMENT,
    t_id int NOT NULL ,
    description varchar(255) NOT NULL ,
    cost int NOT NULL ,
    created_at datetime DEFAULT NOW(),
    updated_at datetime DEFAULT NOW() ON UPDATE NOW(),

    FOREIGN KEY (t_id) REFERENCES ticket(id)
);

# Considering the scarcity of commands available in MySQL,
# I would advise using PostgresSQL instead in this situation.
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