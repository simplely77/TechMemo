-- 常用查询路径上的 B-tree 索引（与 backend/dao 中按 user_id / 外键 / 任务等过滤一致）
-- 执行: sqlx migrate run --database-url "$DATABASE_URL"
-- 说明: 未使用 CONCURRENTLY，若线上表极大需在维护窗口用 CONCURRENTLY 手建等价索引。

-- 登录/注册查重：按 username 等值
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_username
    ON "user" (username);

-- 笔记列表/统计/分类筛选：未删除 + user_id 排序或分组
CREATE INDEX IF NOT EXISTS idx_note_user_id_created_at_desc
    ON note (user_id, created_at DESC)
    WHERE status IS DISTINCT FROM 'deleted';

CREATE INDEX IF NOT EXISTS idx_note_user_id_category_id
    ON note (user_id, category_id)
    WHERE status IS DISTINCT FROM 'deleted';

-- 知识点按来源笔记、按用户
CREATE INDEX IF NOT EXISTS idx_knowledge_point_source_note_id
    ON knowledge_point (source_note_id)
    WHERE source_note_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_knowledge_point_user_id
    ON knowledge_point (user_id);

-- 版本裁剪与按笔记拉历史
CREATE INDEX IF NOT EXISTS idx_note_version_note_id_created_at_asc
    ON note_version (note_id, created_at ASC, id ASC);

-- AI 日志按笔记/任务查
CREATE INDEX IF NOT EXISTS idx_ai_process_log_source_note_id
    ON ai_process_log (source_note_id);

CREATE INDEX IF NOT EXISTS idx_ai_process_log_task_id
    ON ai_process_log (task_id);

-- 关系图/删边
CREATE INDEX IF NOT EXISTS idx_knowledge_relation_from_knowledge_id
    ON knowledge_relation (from_knowledge_id);

CREATE INDEX IF NOT EXISTS idx_knowledge_relation_to_knowledge_id
    ON knowledge_relation (to_knowledge_id);

-- 分类/标签/会话侧栏列表
CREATE INDEX IF NOT EXISTS idx_category_user_id
    ON category (user_id);

CREATE INDEX IF NOT EXISTS idx_tag_user_id
    ON tag (user_id);

CREATE INDEX IF NOT EXISTS idx_chat_session_user_id
    ON chat_session (user_id);

-- 全局脑图/删根节点
CREATE INDEX IF NOT EXISTS idx_note_root_node_note_id
    ON note_root_node (note_id);

-- ILIKE 搜索若数据量大可另建: CREATE EXTENSION IF NOT EXISTS pg_trgm; + GIN 索引，此处不强制开启
