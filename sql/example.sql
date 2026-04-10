-- -- 用户
-- INSERT INTO "user" (id, username, password_hash) VALUES
-- (1, 'liyi', 'hashed_password');

-- -- 分类
-- INSERT INTO category (id, name, user_id) VALUES
-- (1, '后端', 1),
-- (2, '前端', 1),
-- (3, '数据库', 1),
-- (4, 'AI', 1);

-- -- 标签
-- INSERT INTO tag (id, name, user_id) VALUES
-- (1, 'Go', 1),
-- (2, 'Java', 1),
-- (3, 'React', 1),
-- (4, 'PostgreSQL', 1),
-- (5, '向量数据库', 1),
-- (6, '机器学习', 1);

-- -- 笔记 1（Go）
-- INSERT INTO note (id, user_id, title, content_md, category_id)
-- VALUES (
-- 1, 1, 'Go 协程与并发模型',
-- 'Go 的并发模型基于 goroutine 和 channel。

-- goroutine 是轻量级线程，由 Go runtime 管理。
-- channel 用于 goroutine 之间通信，避免共享内存带来的问题。

-- 核心思想是：不要通过共享内存通信，而是通过通信共享内存。',
-- 1
-- );

-- -- 笔记 2（Java）
-- INSERT INTO note (id, user_id, title, content_md, category_id)
-- VALUES (
-- 2, 1, 'Java 注解机制',
-- 'Java 注解（Annotation）是一种元数据，用于描述代码。

-- 常见用途包括：
-- 1. 编译时检查
-- 2. 运行时反射处理
-- 3. 框架配置（如 Spring）

-- 例如 @Override、@Autowired。',
-- 1
-- );

-- -- 笔记 3（React）
-- INSERT INTO note (id, user_id, title, content_md, category_id)
-- VALUES (
-- 3, 1, 'React Hooks 原理',
-- 'React Hooks 允许在函数组件中使用状态和生命周期。

-- useState 用于状态管理。
-- useEffect 用于副作用处理。

-- Hooks 的本质是通过链表结构保存状态。',
-- 2
-- );

-- -- 笔记 4（PostgreSQL）
-- INSERT INTO note (id, user_id, title, content_md, category_id)
-- VALUES (
-- 4, 1, 'PostgreSQL 索引类型',
-- 'PostgreSQL 支持多种索引类型：

-- 1. B-Tree：默认索引，适合范围查询
-- 2. Hash：适合等值查询
-- 3. GIN：适合全文搜索
-- 4. GiST：适合地理数据

-- 合理使用索引可以显著提高查询性能。',
-- 3
-- );

-- -- 笔记 5（向量数据库）
-- INSERT INTO note (id, user_id, title, content_md, category_id)
-- VALUES (
-- 5, 1, '向量数据库原理',
-- '向量数据库用于存储 embedding 向量。

-- 核心能力：
-- 1. 相似度搜索（余弦相似度）
-- 2. 高维向量索引（HNSW）

-- 常见应用：
-- - 语义搜索
-- - 推荐系统
-- - RAG 系统',
-- 4
-- );

-- -- 笔记 6（机器学习）
-- INSERT INTO note (id, user_id, title, content_md, category_id)
-- VALUES (
-- 6, 1, '机器学习基本概念',
-- '机器学习分为三种：

-- 1. 监督学习：有标签数据
-- 2. 无监督学习：无标签数据
-- 3. 强化学习：基于奖励机制

-- 常见算法包括：
-- - 线性回归
-- - 决策树
-- - 神经网络',
-- 4
-- );

-- INSERT INTO note_tag (note_id, tag_id) VALUES
-- (1, 1),
-- (2, 2),
-- (3, 3),
-- (4, 4),
-- (5, 5),
-- (6, 6);