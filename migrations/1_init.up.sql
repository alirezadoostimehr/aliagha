
CREATE TABLE IF NOT EXISTS users
(
    id int PRIMARY KEY AUTO_INCREMENT ,
    name varchar(255) NOT NULL ,
    password varchar(255) NOT NULL ,
    mobile varchar(16) NOT NULL ,
    email varchar(255) NOT NULL UNIQUE ,
    created_at datetime DEFAULT NOW() ,
    updated_at datetime DEFAULT NOW() ON UPDATE NOW()
    );

CREATE TABLE IF NOT EXISTS passengers (
    id int PRIMARY KEY AUTO_INCREMENT ,
    u_id int NOT NULL,
    national_code int NOT NULL ,
    name varchar(255) NOT NULL ,
    birthdate date NOT NULL ,
    created_at datetime DEFAULT NOW() ,
    updated_at datetime DEFAULT NOW() ON UPDATE NOW() ,

    FOREIGN KEY (u_id) REFERENCES users(id) ,
    UNIQUE (u_id, national_code)
    );

CREATE TABLE IF NOT EXISTS cities (
    id int PRIMARY KEY AUTO_INCREMENT ,
    name varchar(255) UNIQUE NOT NULL ,
    created_at datetime DEFAULT NOW() ,
    updated_at datetime DEFAULT NOW() ON UPDATE NOW()
    );

CREATE TABLE IF NOT EXISTS airplanes (
    id int PRIMARY KEY AUTO_INCREMENT ,
    name varchar(255) UNIQUE NOT NULL ,
    created_at datetime DEFAULT NOW() ,
    updated_at datetime DEFAULT NOW() ON UPDATE NOW()
    );

CREATE TABLE IF NOT EXISTS canceling_situations (
    id int PRIMARY KEY AUTO_INCREMENT ,
    description varchar(255) NOT NULL ,
    data varchar(255) NOT NULL ,
    created_at datetime DEFAULT NOW() ,
    updated_at datetime DEFAULT NOW() ON UPDATE NOW()
    );

CREATE TABLE IF NOT EXISTS flights (
    id int PRIMARY KEY AUTO_INCREMENT ,
    dep_city_id int NOT NULL ,
    arr_city_id int NOT NULL ,
    dep_time datetime NOT NULL ,
    arr_time datetime NOT NULL ,
    airplane_id int NOT NULL ,
    airline varchar(255) NOT NULL ,
    cxl_sit_id int NOT NULL ,
    created_at datetime DEFAULT NOW() ,
    updated_at datetime DEFAULT NOW() ON UPDATE NOW() ,

    FOREIGN KEY (dep_city_id) REFERENCES cities(id) ,
    FOREIGN KEY (arr_city_id) REFERENCES cities(id) ,
    FOREIGN KEY (airplane_id) REFERENCES airplanes(id) ,
    FOREIGN KEY (cxl_sit_id) REFERENCES canceling_situations(id)
    );

CREATE TABLE IF NOT EXISTS tickets (
    id int PRIMARY KEY AUTO_INCREMENT ,
    u_id int NOT NULL ,
    p_ids varchar(255) NOT NULL ,
    f_id int NOT NULL ,
    status text NOT NULL ,
    price int NOT NULL ,
    created_at datetime DEFAULT NOW() ,
    updated_at datetime DEFAULT NOW() ON UPDATE NOW() ,

    FOREIGN KEY (u_id) REFERENCES users(id) ,
    FOREIGN KEY (f_id) REFERENCES flights(id)
    );

CREATE TABLE IF NOT EXISTS payments (
    id int PRIMARY KEY AUTO_INCREMENT,
    u_id int NOT NULL ,
    type text NOT NULL ,
    ticket_id int NOT NULL ,
    status varchar(255) NOT NULL ,
    created_at datetime DEFAULT NOW() ,
    updated_at datetime DEFAULT NOW() ON UPDATE NOW() ,

    FOREIGN KEY (u_id) REFERENCES users(id) ,
    FOREIGN KEY (ticket_id) REFERENCES tickets(id)
    );