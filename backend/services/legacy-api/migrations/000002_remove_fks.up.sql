DO $$
DECLARE
  r RECORD;
BEGIN
  -- Drop all foreign keys that reference `users(id)`.
  -- This allows other services to treat `user_id` as an opaque cross-service identifier
  -- without requiring legacy DB referential integrity.
  FOR r IN
    SELECT
      n.nspname AS schema_name,
      t.relname AS table_name,
      c.conname AS constraint_name
    FROM pg_constraint c
    JOIN pg_class t ON t.oid = c.conrelid
    JOIN pg_class u ON u.oid = c.confrelid
    JOIN pg_namespace n ON n.oid = t.relnamespace
    WHERE u.relname = 'users'
      AND c.contype = 'f'
  LOOP
    EXECUTE format(
      'ALTER TABLE %I.%I DROP CONSTRAINT %I',
      r.schema_name, r.table_name, r.constraint_name
    );
  END LOOP;
END $$;

