alter table temp_checker.temperature_data rename column temperature to value;

alter table temp_checker.temperature_data alter column value type float using value::float;