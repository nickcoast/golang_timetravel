CREATE TABLE IF NOT EXISTS "insured" (
	"id"	INTEGER NOT NULL UNIQUE,
	"name"	TEXT NOT NULL,
	"policy_number"	INTEGER NOT NULL UNIQUE,
	"record_timestamp"	INTEGER NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT),
	UNIQUE("policy_number")
);

CREATE TABLE IF NOT EXISTS "employees" ( /* could add "created" field here. And maybe a foreign key "last record"? */
	"id" INTEGER NOT NULL,
	"insured_id" INTEGER NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY("insured_id") REFERENCES "insured"("id") ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS "employees_records" (
	"id"	INTEGER NOT NULL UNIQUE, /* *record* id */
	"employee_id" INTEGER NOT NULL,
	"name"	TEXT NOT NULL,
	"start_date"	TEXT NOT NULL,
	"end_date"	TEXT NOT NULL DEFAULT '0001-01-01', /* Cannot be null and have UNIQUE constraint */	
	"record_timestamp"	INTEGER NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT),
	UNIQUE("employee_id","record_timestamp"),
	FOREIGN KEY("employee_id") REFERENCES "employees"("id") ON DELETE CASCADE ON UPDATE CASCADE
	
);


INSERT INTO insured (name, policy_number, record_timestamp)
VALUES
('Jimmy Temelpa', 1000, CAST(strftime('%s','1984-10-31 12:00:00') AS INT)),
('John Smith', 1001, CAST(strftime('%s','1999-12-31 23:59:59') AS INT));

INSERT INTO employees (insured_id)
VALUES
(1), (1), /* 2 employees for insured_id 1 */
(2), (2), (2), (2); /* 3 employees for insured_id 2 */

INSERT INTO employees_records (employee_id, name, start_date, end_date, record_timestamp)
VALUES
/* 0001-01-01 instead of NULL so UNIQUE constraint works */
(1,'Jimmy Temelpa', '1984-10-01', '0001-01-01', CAST(strftime('%s','1984-10-31 12:00:00') AS INT)), 
(2, 'Mister Bungle', '1984-11-10', '0001-01-01', CAST(strftime('%s','1984-11-15 12:00:00') AS INT)),
(2, 'Mister Bungle', '1984-11-10', '1996-01-02', CAST(strftime('%s','1996-01-02 12:00:00') AS INT)), /* TIMETRAVEL */
(2, 'Mister Bungle', '1984-11-10', '1996-06-01', CAST(strftime('%s','1997-01-02 12:00:00') AS INT)), /* TIMETRAVEL */
(3, 'John Smith', '1985-05-15', '1999-12-25', CAST(strftime('%s','1999-12-31 23:59:59') AS INT)),
(4, 'Jane Doe','1985-05-15', '1999-12-25', CAST(strftime('%s','2000-04-01 12:00:00') AS INT)),
(5, 'Grant Tombly','1985-05-15', '1999-12-25', CAST(strftime('%s','2000-04-01 12:00:00') AS INT));


/* JOIN table not needed for 1-to-1 relationship between insured and their address (assuming 1 address) */
/*
CREATE TABLE IF NOT EXISTS "insured_addresses" (
	"id" INTEGER NOT NULL,
	"insured_id" INTEGER NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT),
	FOREIGN KEY("insured_id") REFERENCES "insured" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);
 */

 /* Do not need UNIQUE constraint. Insured could move back and forth. Just need to ensure address changed from last record */
CREATE TABLE "insured_addresses_records" (
	"id"	INTEGER NOT NULL,
	/* addresses_id only needed if allow insured >1 address */
	"address"	TEXT NOT NULL,
	"insured_id"	INTEGER NOT NULL,
	"record_timestamp"	INTEGER NOT NULL,
	/* CONSTRAINT "insured_addresses_insured_id_address_unq" UNIQUE("insured_id","address"),*/
	FOREIGN KEY(insured_id) REFERENCES insured(id) ON DELETE CASCADE ON UPDATE CASCADE
	PRIMARY KEY("id" AUTOINCREMENT)
);


INSERT INTO insured_addresses_records
(address, insured_id, record_timestamp)
VALUES
('123 Fake Street, Springfield, Oregon', 1, CAST(strftime('%s','1984-10-31 12:00:00') AS INT)),
('123 REAL Street, Springfield, Oregon', 1, CAST(strftime('%s','1984-11-15 12:00:01') AS INT)),
('Flavortown', 1, CAST(strftime('%s','1996-01-02 11:59:49') AS INT)),
('Mars', 1, CAST(strftime('%s','1997-01-02 12:00:01') AS INT));
