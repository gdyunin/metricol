CREATE TABLE IF NOT EXISTS metrics (
   id SERIAL PRIMARY KEY,
   m_type TEXT NOT NULL,
   m_name TEXT NOT NULL,
   m_value JSONB NOT NULL,
   CONSTRAINT unique_type_name UNIQUE (m_type, m_name)
);

DO $$
    BEGIN
        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE schemaname = 'public' AND indexname = 'idx_metrics_type_name'
        ) THEN
            CREATE INDEX idx_metrics_type_name ON metrics (m_type, m_name);
    END IF;
END $$;