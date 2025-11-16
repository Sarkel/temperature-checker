alter table temp_checker.location drop column location_sid;

alter table temp_checker.location_sensor alter column sensor_id type varchar(20) using sensor_id::varchar(20);

alter table temp_checker.location_sensor rename column sensor_sid to sensor_id;