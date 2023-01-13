INSERT INTO insured (name, policy_number, record_timestamp)
VALUES
('Jimmy Temelpa', 1000, CAST(strftime('%s','1984-10-31 12:00:00') AS INT)),
('John Smith', 1001, CAST(strftime('%s','1999-12-31 23:59:59') AS INT));


INSERT INTO employees (name, start_date, end_date, insured_id, record_timestamp)
VALUES
('Jimmy Temelpa', '1984-10-01', NULL, 1, CAST(strftime('%s','1984-10-31 12:00:00') AS INT)),
('Mister Bungle', '1984-11-10', NULL, 1, CAST(strftime('%s','1984-11-15 12:00:00') AS INT)),
('John Smith', '1985-05-15', '1999-12-25', 2, CAST(strftime('%s','1999-12-31 23:59:59') AS INT)),
('Jane Doe','1985-05-15', '1999-12-25',2, CAST(strftime('%s','2000-04-01 12:00:00') AS INT)),
('Grant Tombly','1985-05-15', '1999-12-25',2, CAST(strftime('%s','2000-04-01 12:00:00') AS INT));
