CREATE TABLE bed_type (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 name text NOT NULL UNIQUE CHECK(length(name) BETWEEN 1 AND 80),
 description text NOT NULL DEFAULT '' CHECK(length(description)<=2000),
 active boolean NOT NULL DEFAULT true,
 version integer NOT NULL DEFAULT 1 CHECK(version>0)
);
INSERT INTO bed_type(name) SELECT DISTINCT bed_type FROM hospital_bed;
ALTER TABLE hospital_bed ADD COLUMN type_id uuid REFERENCES bed_type(id);
UPDATE hospital_bed b SET type_id=t.id FROM bed_type t WHERE b.bed_type=t.name;
ALTER TABLE hospital_bed ALTER COLUMN type_id SET NOT NULL;
