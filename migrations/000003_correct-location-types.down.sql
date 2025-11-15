alter table temp_checker.location alter column longitude type numeric using longitude::numeric;
alter table temp_checker.location alter column latitude type numeric using latitude::numeric;