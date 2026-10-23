-- MUSA's tables. Ids are UUIDs made on the device, money is whole cents, times are Unix milliseconds
-- and rows are soft-deleted (deleted_at), so a future sync can tell a deletion from a missing row.

CREATE TABLE IF NOT EXISTS product (
	id TEXT NOT NULL PRIMARY KEY,
	kind TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	deleted_at INTEGER
);

CREATE TABLE IF NOT EXISTS digital_product (
	product_id TEXT NOT NULL PRIMARY KEY,
	delivery_format TEXT NOT NULL,
	turnaround_days INTEGER
);

CREATE TABLE IF NOT EXISTS physical_product (
	product_id TEXT NOT NULL PRIMARY KEY,
	material TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS product_variant (
	id TEXT NOT NULL PRIMARY KEY,
	product_id TEXT NOT NULL,
	name TEXT NOT NULL,
	price_cents INTEGER NOT NULL,
	cost_cents INTEGER NOT NULL,
	position INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	deleted_at INTEGER
);


CREATE INDEX IF NOT EXISTS product_variant_product ON product_variant(product_id);
CREATE TABLE IF NOT EXISTS physical_variant (
	variant_id TEXT NOT NULL PRIMARY KEY,
	sku TEXT,
	stock INTEGER NOT NULL,
	min_stock INTEGER,
	weight_grams INTEGER,
	width_mm INTEGER,
	height_mm INTEGER,
	depth_mm INTEGER
);


CREATE INDEX IF NOT EXISTS physical_variant_sku ON physical_variant(sku);
CREATE TABLE IF NOT EXISTS tag (
	id TEXT NOT NULL PRIMARY KEY,
	name TEXT NOT NULL,
	normalized_name TEXT NOT NULL UNIQUE,
	color TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	deleted_at INTEGER
);

CREATE TABLE IF NOT EXISTS product_tag (
	product_id TEXT NOT NULL,
	tag_id TEXT NOT NULL,
	position INTEGER NOT NULL,
	PRIMARY KEY (product_id, tag_id)
);

CREATE TABLE IF NOT EXISTS event (
	id TEXT NOT NULL PRIMARY KEY,
	name TEXT NOT NULL,
	location TEXT NOT NULL,
	start_date TEXT NOT NULL,
	end_date TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	deleted_at INTEGER
);

CREATE TABLE IF NOT EXISTS event_stock (
	event_id TEXT NOT NULL,
	variant_id TEXT NOT NULL,
	taken INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	PRIMARY KEY (event_id, variant_id)
);

CREATE TABLE IF NOT EXISTS event_expense (
	id TEXT NOT NULL PRIMARY KEY,
	event_id TEXT NOT NULL,
	concept TEXT NOT NULL,
	amount_cents INTEGER NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	deleted_at INTEGER
);


CREATE INDEX IF NOT EXISTS event_expense_event ON event_expense(event_id);
CREATE TABLE IF NOT EXISTS sale (
	id TEXT NOT NULL PRIMARY KEY,
	channel TEXT NOT NULL,
	event_id TEXT,
	customer TEXT NOT NULL,
	note TEXT NOT NULL,
	shipping_charged_cents INTEGER NOT NULL,
	shipping_cost_cents INTEGER NOT NULL,
	sold_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	deleted_at INTEGER
);


CREATE INDEX IF NOT EXISTS sale_event ON sale(event_id);
CREATE TABLE IF NOT EXISTS sale_line (
	id TEXT NOT NULL PRIMARY KEY,
	sale_id TEXT NOT NULL,
	variant_id TEXT NOT NULL,
	product_name TEXT NOT NULL,
	variant_name TEXT NOT NULL,
	quantity INTEGER NOT NULL,
	unit_price_cents INTEGER NOT NULL,
	unit_cost_cents INTEGER NOT NULL,
	position INTEGER NOT NULL
);


CREATE INDEX IF NOT EXISTS sale_line_sale ON sale_line(sale_id);
