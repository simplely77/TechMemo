-- Add migration script here
ALTER TABLE search_history ADD COLUMN target_type VARCHAR(32) NOT NULL;