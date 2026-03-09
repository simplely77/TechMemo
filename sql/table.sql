-- =========================================
-- TechMemo 数据库表
-- =========================================

-- 1. 用户表
CREATE TABLE "user" (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. 分类表
CREATE TABLE category (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    user_id BIGINT NOT NULL
);

-- 3. 标签表
CREATE TABLE tag (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    user_id BIGINT NOT NULL
);

-- 4. 笔记表
CREATE TABLE note (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content_md TEXT NOT NULL,
    note_type VARCHAR(20) DEFAULT 'unknown',
    category_id BIGINT,
    status VARCHAR(20) DEFAULT 'normal',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. 笔记版本表
CREATE TABLE note_version (
    id BIGSERIAL PRIMARY KEY,
    note_id BIGINT NOT NULL,
    content_md TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. 笔记-标签关联表
CREATE TABLE note_tag (
    note_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    PRIMARY KEY(note_id, tag_id)
);

-- 7. 知识点表
CREATE TABLE knowledge_point (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    source_note_id BIGINT,
    importance_score FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 8. 知识点关系表
CREATE TABLE knowledge_relation (
    id BIGSERIAL PRIMARY KEY,
    from_knowledge_id BIGINT NOT NULL,
    to_knowledge_id BIGINT NOT NULL,
    relation_type VARCHAR(20)
);

-- 9. AI 处理日志表
CREATE TABLE ai_process_log (
    id BIGSERIAL PRIMARY KEY,
    source_note_id BIGINT NOT NULL,
    task_id VARCHAR(64) NOT NULL,
    target_type VARCHAR(20) NOT NULL,   -- note / knowledge
    target_id BIGINT NOT NULL,
    process_type VARCHAR(20) NOT NULL,  -- extract / embedding
    model_name VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'success',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 10. 向量表 (pgvector)
CREATE TABLE embedding (
    id BIGSERIAL PRIMARY KEY,
    target_type VARCHAR(20) NOT NULL,  -- note / knowledge
    target_id BIGINT NOT NULL,
    vector vector(1536),               -- 假设使用 1536 维度的 embedding
    model_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
