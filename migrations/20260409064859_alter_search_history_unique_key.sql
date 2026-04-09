-- Add migration script here
ALTER TABLE search_history
DROP CONSTRAINT uk_user_keyword;

ALTER TABLE search_history
ADD CONSTRAINT uk_user_keyword_type_target
UNIQUE (user_id, keyword, search_type, target_type);