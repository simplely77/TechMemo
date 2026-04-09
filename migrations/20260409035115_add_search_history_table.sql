-- Add migration script here
CREATE TABLE search_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    keyword VARCHAR(255) NOT NULL,

    search_type VARCHAR(32) DEFAULT 'keyword',

    last_searched_at TIMESTAMP NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uk_user_keyword UNIQUE (user_id, keyword)
);

-- 索引单独创建
CREATE INDEX idx_user_time 
ON search_history (user_id, last_searched_at DESC);