create schema temp_checker;

create table temp_checker.location
(
    location_id   int primary key generated always as identity,
    location_name varchar(255) unique not null,
    latitude      numeric(9, 6)       not null,
    longitude     numeric(9, 6)       not null
);

create type temp_checker.sensor_type as enum('local', 'api');

create table temp_checker.location_sensor
(
    location_sensor_id int primary key generated always as identity,
    location_id        int references temp_checker.location (location_id) not null,
    sensor_id          varchar(20)                                        not null,
    type               temp_checker.sensor_type                           not null,
    constraint unique_location_sensor unique (location_id, sensor_id)
);

create table temp_checker.temperature_data
(
    temperature_data_id int primary key generated always as identity,
    location_sensor_id  int references temp_checker.location_sensor (location_sensor_id) not null,
    temperature         numeric                                                          not null,
    timestamp           timestamptz                                                      not null
);
