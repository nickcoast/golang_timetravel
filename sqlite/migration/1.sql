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


/* Assuming Sqlite will enforce NOT NULL, this won't be needed */
/* CREATE TRIGGER employees_end_date
AFTER INSERT ON employees
BEGIN UPDATE employees
SET end_date = "0001-01-01"
WHERE NEW.end_date IS NULL;
END */