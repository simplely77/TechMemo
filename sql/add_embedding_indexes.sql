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
