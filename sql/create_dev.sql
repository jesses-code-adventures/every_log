\c everylog
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_roles
        WHERE rolname = 'dev'
    ) THEN
        CREATE ROLE dev WITH LOGIN PASSWORD '2jd78sj2hd7wkaqsk237ksf';
    END IF;
END
$$;
