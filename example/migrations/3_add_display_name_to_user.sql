ALTER TABLE application_user ADD COLUMN display_name TEXT;

UPDATE application_user SET display_name = name;

ALTER TABLE application_user ALTER COLUMN display_name SET NOT NULL;