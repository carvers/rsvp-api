CREATE TABLE parties (
	id VARCHAR(36) PRIMARY KEY,
	lead_id VARCHAR(36) NOT NULL,
	name VARCHAR(128) NOT NULL,
	sort_value VARCHAR(128) NOT NULL,
	address TEXT NOT NULL,
	magic_word VARCHAR(128) NOT NULL
);

CREATE TABLE people (
	id VARCHAR(36) PRIMARY KEY,
	party_id VARCHAR(36) NOT NULL,
	name VARCHAR(128) NOT NULL,
	email VARCHAR(128) NOT NULL,
	gets_plus_one BOOLEAN NOT NULL,
	plus_one_id VARCHAR(36) NOT NULL,
	is_plus_one BOOLEAN NOT NULL,
	is_plus_one_of_id VARCHAR(36) NOT NULL,
	replied BOOLEAN NOT NULL,
	reply BOOLEAN NOT NULL,
	dietary_restrictions TEXT NOT NULL,
	song_request TEXT NOT NULL,
	is_child BOOLEAN NOT NULL,
	will_accompany_id VARCHAR(36) NOT NULL,

	hiking BOOLEAN NOT NULL,
	kayaking BOOLEAN NOT NULL,
	jetski BOOLEAN NOT NULL,
	fishing BOOLEAN NOT NULL,
	hanford BOOLEAN NOT NULL,
	ligo BOOLEAN NOT NULL,
	reach BOOLEAN NOT NULL,
	bechtel BOOLEAN NOT NULL,
	wine BOOLEAN NOT NULL,
	escape_room BOOLEAN NOT NULL
);
