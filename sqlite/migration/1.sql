CREATE TABLE IF NOT EXISTS "insured" (
	"id"	INTEGER NOT NULL UNIQUE,
	"name"	TEXT NOT NULL,
	"policy_number"	INTEGER NOT NULL UNIQUE,
	"record_timestamp"	INTEGER NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT),
	UNIQUE("policy_number")
);


CREATE TABLE IF NOT EXISTS "employees" (
	"id"	INTEGER NOT NULL UNIQUE,
	"name"	TEXT NOT NULL,
	"start_date"	TEXT NOT NULL,
	"end_date"	TEXT NOT NULL DEFAULT '0001-01-01', /* Cannot be null and have UNIQUE constraint */
	"insured_id"	INTEGER NOT NULL,
	"record_timestamp"	INTEGER NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT),
	UNIQUE("insured_id","name","start_date","end_date"),
	FOREIGN KEY("insured_id") REFERENCES "insured"("id") ON DELETE CASCADE ON UPDATE CASCADE
	
);

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
('Mister Bungle', '1984-11-10', '1996-06-01', 1, CAST(strftime('%s','1997-01-02 12:00:00') AS INT)), /* TIMETRAVEL */
('John Smith', '1985-05-15', '1999-12-25', 2, CAST(strftime('%s','1999-12-31 23:59:59') AS INT)),
('Jane Doe','1985-05-15', '1999-12-25',2, CAST(strftime('%s','2000-04-01 12:00:00') AS INT)),
('Grant Tombly','1985-05-15', '1999-12-25',2, CAST(strftime('%s','2000-04-01 12:00:00') AS INT));


CREATE TABLE "insured_addresses" (
	"id"	INTEGER NOT NULL,
	"address"	TEXT NOT NULL,
	"insured_id"	INTEGER NOT NULL,
	"record_timestamp"	INTEGER NOT NULL,
	CONSTRAINT "insured_addresses_insured_id_address_unq" UNIQUE("insured_id","address"),
	FOREIGN KEY(insured_id) REFERENCES insured(id) ON DELETE CASCADE ON UPDATE CASCADE
	PRIMARY KEY("id" AUTOINCREMENT)
);


INSERT INTO insured_addresses
(address, insured_id, record_timestamp)
VALUES
('123 Fake Street, Springfield, Oregon', 1, CAST(strftime('%s','1984-10-31 12:00:00') AS INT)),
('123 REAL Street, Springfield, Oregon', 1, CAST(strftime('%s','1984-11-15 12:00:01') AS INT)),
('Flavortown', 1, CAST(strftime('%s','1996-01-02 11:59:49') AS INT)),
('Mars', 1, CAST(strftime('%s','1997-01-02 12:00:01') AS INT))
