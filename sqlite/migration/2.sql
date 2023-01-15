INSERT INTO insured (name, policy_number, record_timestamp)
VALUES
('Jimmy Temelpa', 1000, CAST(strftime('%s','1984-10-31 12:00:00') AS INT)),
('John Smith', 1001, CAST(strftime('%s','1999-12-31 23:59:59') AS INT));


INSERT INTO employees (name, start_date, end_date, insured_id, record_timestamp)
VALUES
/* 0001-01-01 instead of NULL so UNIQUE constraint works */
('Jimmy Temelpa', '1984-10-01', '0001-01-01', 1, CAST(strftime('%s','1984-10-31 12:00:00') AS INT)), 
('Mister Bungle', '1984-11-10', '0001-01-01', 1, CAST(strftime('%s','1984-11-15 12:00:00') AS INT)),
('Mister Bungle', '1984-11-10', '1996-01-02', 1, CAST(strftime('%s','1996-01-02 12:00:00') AS INT)), /* TIMETRAVEL */
('John Smith', '1985-05-15', '1999-12-25', 2, CAST(strftime('%s','1999-12-31 23:59:59') AS INT)),
('Jane Doe','1985-05-15', '1999-12-25',2, CAST(strftime('%s','2000-04-01 12:00:00') AS INT)),
('Grant Tombly','1985-05-15', '1999-12-25',2, CAST(strftime('%s','2000-04-01 12:00:00') AS INT));


/*
Tried this. Works, but makes it hard to test Time Travel in unit tests.
overwrites any inserted timestamp
Must come after the above inserts
*/
/* CREATE TRIGGER employees_timestamp
AFTER INSERT
ON employees
BEGIN
UPDATE employees SET record_timestamp = strftime('%s') WHERE id = NEW.id;
END; */