CREATE TABLE "insured_addresses" (
	"id"	INTEGER NOT NULL,
	"address"	TEXT NOT NULL,
	"insured_id"	INTEGER NOT NULL,
	"record_timesetamp"	INTEGER NOT NULL,
	CONSTRAINT "insured_addresses_insured_id_address_unq" UNIQUE("insured_id","address"),
	PRIMARY KEY("id" AUTOINCREMENT)
);


INSERT INTO insured_addresses
(address, insured_id, record_timestamp)
VALUES
('123 Fake Street, Springfield, Oregon', 1, 1673687536),
('123 REAL Street, Springfield, Oregon', 1, 1673687537),
('Flavortown', 1, 1673834152)