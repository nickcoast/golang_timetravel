CREATE TABLE "insured_addresses" (
	"id"	INTEGER NOT NULL,
	"address"	TEXT NOT NULL,
	"insured_id"	INTEGER NOT NULL,
	"record_timestamp"	INTEGER NOT NULL,
	CONSTRAINT "insured_addresses_insured_id_address_unq" UNIQUE("insured_id","address"),
	FOREIGN KEY(insured_id) REFERENCES insured(id)
	PRIMARY KEY("id" AUTOINCREMENT)
);


INSERT INTO insured_addresses
(address, insured_id, record_timestamp)
VALUES
('123 Fake Street, Springfield, Oregon', 1, CAST(strftime('%s','1984-10-31 12:00:00') AS INT)),
('123 REAL Street, Springfield, Oregon', 1, CAST(strftime('%s','1984-11-15 12:00:01') AS INT)),
('Flavortown', 1, CAST(strftime('%s','1996-01-02 11:59:49') AS INT)),
('Mars', 1, CAST(strftime('%s','1997-01-02 12:00:01') AS INT))
