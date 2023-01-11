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
	"end_date"	TEXT,
	"insured_id"	INTEGER NOT NULL,
	"record_timestamp"	INTEGER NOT NULL,
	UNIQUE("insured_id","name","start_date","end_date"),
	FOREIGN KEY("insured_id") REFERENCES "insured"("id") ON DELETE CASCADE ON UPDATE CASCADE,
	PRIMARY KEY("id" AUTOINCREMENT)
);
CREATE TRIGGER employees_timestamp
AFTER INSERT
ON employees
BEGIN
UPDATE employees SET record_timestamp = strftime('%s') WHERE id = NEW.id;
END;
