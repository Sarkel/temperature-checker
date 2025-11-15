-- name: CreateTemperatureData :many
insert into temp_checker.temperature_data(location_sensor_id, value, timestamp)
select unnest(sqlc.arg(location_sensor_ids)::int[]),
       unnest(sqlc.arg(values)::float[]),
       unnest(sqlc.arg(timestamps)::timestamptz[])
returning temperature_data_id;

-- name: GetAPILocationSensors :many
select ls.location_sensor_id,
       ls.sensor_id,
       l.location_name,
       l.latitude,
       l.longitude,
       l.location_id
from temp_checker.location_sensor AS ls
         join temp_checker.location AS l on ls.location_id = l.location_id
where ls.type = 'api';

