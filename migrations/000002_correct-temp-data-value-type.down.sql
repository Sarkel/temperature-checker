alter table temp_checker.temperature_data rename column value to temperature;

alter table temp_checker.temperature_data alter column temperature type numeric using value::numeric;