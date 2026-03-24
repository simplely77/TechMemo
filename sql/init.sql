-- =========================================
-- TechMemo 数据库表
-- =========================================

-- 安装插件
CREATE EXTENSION IF NOT EXISTS vector;

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
    target_type VARCHAR(20) NOT NULL,   -- note / knowledge / chat_message
    target_id BIGINT NOT NULL,
    process_type VARCHAR(20) NOT NULL,  -- extract / embedding / classify / chat_embedding
    model_name VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'success',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 10. 向量表 (pgvector)
CREATE TABLE embedding (
    id BIGSERIAL PRIMARY KEY,
    target_type VARCHAR(20) NOT NULL,  -- note / knowledge / chat_message
    target_id BIGINT NOT NULL,
    vector vector(384),               -- 使用 384 维度的 embedding
    model_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 11. 笔记顶节点表（每篇笔记的根知识点，用于全局思维导图）
CREATE TABLE note_root_node (
    id BIGSERIAL PRIMARY KEY,
    note_id BIGINT NOT NULL,
    root_knowledge_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    importance_score FLOAT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 12. 聊天会话表
CREATE TABLE chat_session (
    id BIGSERIAL primary key,   
    user_id BIGINT NOT NULL,
    title VARCHAR(255) DEFAULT '新对话',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 13. 聊天消息表
CREATE TABLE chat_message (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role VARCHAR(20) NOT NULL,  -- user / system / assitant
    content TEXT NOT NULL,
    token_count int DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- 为 embedding 表创建 pgvector 索引以提高语义搜索性能
-- 注意：IVFFlat 索引需要先有一定量的数据（建议至少 1000 条）才能创建
-- 如果数据量较少，可以先不创建索引，等数据量增加后再创建

-- 为 vector 列创建 IVFFlat 索引（使用余弦距离）
-- lists 参数根据数据量调整，一般为 sqrt(总行数)
-- 对于小数据集，可以使用较小的 lists 值（如 10-100）
CREATE INDEX IF NOT EXISTS idx_embedding_vector_cosine
ON embedding USING ivfflat (vector vector_cosine_ops)
WITH (lists = 100);

-- 为常用查询字段创建索引
CREATE INDEX IF NOT EXISTS idx_embedding_target_type
ON embedding(target_type);

CREATE INDEX IF NOT EXISTS idx_embedding_target_id
ON embedding(target_id);

-- 复合索引，用于加速按 target_type 过滤的查询
CREATE INDEX IF NOT EXISTS idx_embedding_target_type_id
ON embedding(target_type, target_id);

CREATE INDEX idx_chat_message_session_id ON chat_message(session_id);
CREATE INDEX idx_chat_message_created_at ON chat_message(created_at DESC);
CREATE INDEX idx_chat_message_session_created ON chat_message(session_id,created_at DESC);