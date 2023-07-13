# Database Documentation

## Tables

### users

- id: int (Primary Key, Auto Increment)
- name: varchar(255) (Not Null)
- password: varchar(255) (Not Null)
- cellphone: varchar(16) (Not Null)
- email: varchar(255) (Not Null, Unique)
- created_at: datetime (Default: Current Timestamp)
- updated_at: datetime (Default: Current Timestamp, On Update: Current Timestamp)

### passengers

- id: int (Primary Key, Auto Increment)
- u_id: int (Not Null, Foreign Key: users.id)
- national_code: int (Not Null)
- name: varchar(255) (Not Null)
- birthdate: date (Not Null)
- created_at: datetime (Default: Current Timestamp)
- updated_at: datetime (Default: Current Timestamp, On Update: Current Timestamp)

### cities

- id: int (Primary Key, Auto Increment)
- name: varchar(255) (Unique, Not Null)
- created_at: datetime (Default: Current Timestamp)
- updated_at: datetime (Default: Current Timestamp, On Update: Current Timestamp)

### airplanes

- id: int (Primary Key, Auto Increment)
- name: varchar(255) (Unique, Not Null)
- created_at: datetime (Default: Current Timestamp)
- updated_at: datetime (Default: Current Timestamp, On Update: Current Timestamp)

### canceling_situations

- id: int (Primary Key, Auto Increment)
- description: varchar(255) (Not Null)
- data: varchar(255) (Not Null)
- created_at: datetime (Default: Current Timestamp)
- updated_at: datetime (Default: Current Timestamp, On Update: Current Timestamp)

### flights

- id: int (Primary Key, Auto Increment)
- dep_city_id: int (Not Null, Foreign Key: cities.id)
- arr_city_id: int (Not Null, Foreign Key: cities.id)
- dep_time: datetime (Not Null)
- arr_time: datetime (Not Null)
- airplane_id: int (Not Null, Foreign Key: airplanes.id)
- airline: varchar(255) (Not Null)
- price: int (Not Null)
- cxl_sit_id: int (Not Null, Foreign Key: canceling_situations.id)
- created_at: datetime (Default: Current Timestamp)
- updated_at: datetime (Default: Current Timestamp, On Update: Current Timestamp)

### tickets

- id: int (Primary Key, Auto Increment)
- u_id: int (Not Null, Foreign Key: users.id)
- p_id: int (Not Null, Foreign Key: passengers.id)
- f_id: int (Not Null, Foreign Key: flights.id)
- status: text (Not Null)
- created_at: datetime (Default: Current Timestamp)
- updated_at: datetime (Default: Current Timestamp, On Update: Current Timestamp)

### payments

- id: int (Primary Key, Auto Increment)
- u_id: int (Not Null, Foreign Key: users.id)
- type: text (Not Null)
- ticket_id: int (Not Null, Foreign Key: tickets.id)
- created_at: datetime (Default: Current Timestamp)
- updated_at: datetime (Default: Current Timestamp, On Update: Current Timestamp)

## Relations

- The `users` table has a one-to-many relationship with the `passengers` table through the `u_id` foreign key.
- The `users` table has a one-to-many relationship with the `tickets` table through the `u_id` foreign key.
- The `passengers` table has a one-to-many relationship with the `tickets` table through the `p_id` foreign key.
- The `flights` table has a one-to-many relationship with the `tickets` table through the `f_id` foreign key.
- The `users` table has a one-to-many relationship with the `payments` table through the `u_id` foreign key.
- The `tickets` table has a one-to-many relationship with the `payments` table through the `ticket_id` foreign key.
- The `cities` table is referenced by the `dep_city_id` and `arr_city_id` foreign keys in the `flights` table.
- The `airplanes` table is referenced by the `airplane_id` foreign key in the `flights` table.
- The `canceling_situations` table is referenced by the `cxl_sit_id` foreign key in the `flights` table.

## Database Schematics

```
users
+----+------+----------+-----------+--------+------------+-------------+
| id | name | password | cellphone | email  | created_at | updated_at  |
+----+------+----------+-----------+--------+------------+-------------+
|    |      |          |           |                     |             |
+----+------+----------+-----------+---------------------+-------------+

passengers
+----+------+--------------+------+------------+------------+------------+
| id | u_id | national_code| name | birthdate  | created_at | updated_at |
+----+------+--------------+------+------------+------------+------------+
|    |      |              |      |            |            |            |
+----+------+--------------+------+------------+------------+------------+

cities
+----+------+
| id | name |
+----+------+
|    |      |
+----+------+

airplanes
+----+------+
| id | name |
+----+------+
|    |      |
+----+------+

canceling_situations
+----+--------------+------+
| id | description  | data |
+----+--------------+------+
|    |              |      |
+----+--------------+------+

flights
+----+--------------+--------------+-----------+----------+--------------+---------+-------+------------+------------+
| id | dep_city_id  | arr_city_id  | dep_time  | arr_time | airplane_id  | airline | price | created_at | updated_at |
+----+--------------+--------------+-----------+----------+--------------+---------+-------+------------+------------+
|    |              |              |           |          |              |         |       |            |            |
+----+--------------+--------------+-----------+----------+--------------+---------+-------+------------+------------+

tickets
+----+------+------+------+--------+-------------+---------------+
| id | u_id | p_id | f_id | status | created_at  | updated_at    |
+----+------+------+------+--------+-------------+---------------+
|    |      |      |      |        |             |               |
+----+------+------+------+--------+-------------+---------------+

payments
+----+------+-------+-----------+-------------+--------------+
| id | u_id | type  | ticket_id | created_at  | updated_at   |
+----+------+-------+-----------+-------------+--------------+
|    |      |       |           |             |              |
+----+------+-------+-----------+-------------+--------------+
```

This diagram represents the tables and their relationships in the database schema. The primary keys are denoted by the `id` column, and foreign keys are indicated by the references to the primary keys of other tables. The `created_at` and `updated_at` columns are used to track the creation and modification timestamps of the records.
