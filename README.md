# TechMemo —— 个人技术知识演化系统

## 一、项目概述

**TechMemo** 是一个基于大语言模型（LLM）的个人技术知识管理与演化系统，旨在帮助用户通过智能整理和关联技术笔记，提升知识积累、复习和应用效率。

系统通过 AI 自动化分析用户的技术笔记内容，将零散的信息转化为**结构化知识**，并在长期学习过程中实现知识的**持续演化与强化**。同时，系统可以基于**多个笔记生成思维导图**，帮助用户可视化知识结构。

---

## 二、用户需求分析

### 2.1 用户角色

#### 普通用户
- 学生、程序员、技术从业者
- 核心需求：
  - 记录技术学习笔记
  - 使用 AI 自动整理和总结内容
  - 通过思维导图和复习机制提升学习效率

#### 管理员（可选）
- 负责系统配置、监控和维护
- 系统早期阶段可暂不考虑，重点放在用户端功能实现

---

## 三、核心使用场景与功能分析

### 场景一：记录和管理技术笔记
- **用户目标**：高效记录、管理技术学习笔记
- **用户行为**：
  - 创建、编辑、删除笔记
  - Markdown 支持文本、代码块、图片
  - 添加分类与标签（如 Python、算法等）
  - 自动生成版本历史
- **期望功能**：
  - 自动保存
  - 历史版本回溯
  - 支持关键词和标签搜索

### 场景二：AI 自动整理笔记
- **用户目标**：将零散笔记自动整理为结构化知识
- **用户行为**：
  - 系统在笔记创建或编辑时自动分析内容
  - 用户查看 AI 自动生成的摘要和结构化内容
- **期望功能**：
  - 自动提取要点
  - 识别知识点层级关系
  - 补充相关概念和背景
  - 错误检测与修正建议

### 场景三：跨笔记生成思维导图
- **用户目标**：通过可视化方式理解整个知识体系
- **用户行为**：
  - 浏览跨笔记生成的思维导图
  - 点击节点查看对应笔记
  - 可选手动调整节点和分支
- **期望功能**：
  - 自动生成跨笔记思维导图
  - 展示笔记之间的知识关联
  - 导图实时同步笔记更新
  - 节点交互支持查看详细内容

### 场景四：知识复习与智能推荐
- **用户目标**：系统性复习已学知识，提高学习效率
- **用户行为**：
  - 查看系统推荐的复习内容
  - 接收定期复习提醒
  - 查看复习历史
- **期望功能**：
  - 基于学习历史智能推荐复习内容
  - 自定义复习周期和提醒
  - 记录复习结果（已掌握 / 未掌握）

### 场景五：基于笔记的问答系统
- **用户目标**：通过自然语言提问，快速获取基于笔记的答案
- **用户行为**：
  - 输入技术问题
  - 查看系统生成答案
  - 查看答案引用的笔记
- **期望功能**：
  - 基于 NLP 的智能问答
  - 优先使用用户已有笔记生成答案
  - 支持查看答案来源

### 场景六：多设备同步与云备份
- **用户目标**：跨设备无缝使用系统，确保数据安全
- **用户行为**：
  - 在手机、电脑等多设备访问系统
  - 自动同步笔记和学习记录
- **期望功能**：
  - 跨平台数据同步
  - 云端自动备份
  - 防止数据丢失

---

# TechMemo 数据库表结构设计说明

TechMemo —— 个人技术知识演化系统的核心数据库表结构及各字段设计说明。  
系统采用“笔记 → 知识点 → AI 演化”的数据组织方式，兼顾系统可实现性与后续扩展能力。  

**数据库选型及原因：**  
推荐使用 **PostgreSQL + pgvector**：
- PostgreSQL 稳定、开源、支持事务和复杂关联查询；
- pgvector 扩展提供向量数据类型和相似度搜索能力，方便实现 RAG 检索；
- 后期可直接在 PostgreSQL 内完成向量存储和近似最近邻查询，无需额外向量数据库。

---

## 1. 用户表（user）

用于存储系统用户的基础信息。TechMemo 以个人使用为主，用户表设计保持简洁。

| 字段名 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 用户唯一标识，主键 |
| username | VARCHAR | 用户名 |
| password_hash | VARCHAR | 用户密码加密存储 |
| created_at | DATETIME | 用户创建时间 |

---

## 2. 分类表（category）

用于对技术笔记进行**粗粒度结构化分类**，一个笔记对应一个主要分类。

| 字段名 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 分类唯一标识 |
| name | VARCHAR | 分类名称（如：Python、数据结构、后端） |
| user_id | BIGINT | 关联用户 ｜

---

## 3. 笔记表（note）

用于存储用户的原始学习笔记，是系统的“事实层”。  
笔记内容以 Markdown 格式保存，AI 处理不会直接修改原始内容。

| 字段名 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 笔记唯一标识 |
| user_id | BIGINT | 笔记所属用户 |
| title | VARCHAR | 笔记标题 |
| content_md | TEXT | 笔记正文（Markdown 格式） |
| note_type | VARCHAR | 笔记类型（tech / daily / unknown），由 AI 判断 |
| category_id | BIGINT | 笔记所属分类（外键，单选） |
| status | VARCHAR | 笔记状态（normal / deleted） |
| created_at | DATETIME | 笔记创建时间 |
| updated_at | DATETIME | 笔记更新时间 |

---

## 4. 笔记版本表（note_version）

用于记录笔记的历史版本，支持知识演化过程中的内容回溯。

| 字段名 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 版本记录唯一标识 |
| note_id | BIGINT | 对应的笔记 ID |
| content_md | TEXT | 该版本的笔记内容 |
| created_at | DATETIME | 版本生成时间 |

---

## 5. 标签表（tag）

用于对笔记进行**细粒度语义标注**，支持多标签。

| 字段名 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 标签唯一标识 |
| name | VARCHAR | 标签名称（如：装饰器、BFS、缓存） |
| user_id | BIGINT | 关联用户 ｜

---

## 6. 笔记-标签关联表（note_tag）

用于建立笔记与标签之间的多对多关系。

| 字段名 | 类型 | 说明 |
|------|------|------|
| note_id | BIGINT | 笔记 ID |
| tag_id | BIGINT | 标签 ID |

---

## 7. 知识点表（knowledge_point）

知识点是系统中由 AI 从笔记中抽取的**核心知识单元**，是知识关联、思维导图和智能复习的基础。

| 字段名 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 知识点唯一标识 |
| user_id | BIGINT | 知识点所属用户 |
| name | VARCHAR | 知识点名称 |
| description | TEXT | 知识点说明，由 AI 生成 |
| source_note_id | BIGINT | 来源笔记 ID |
| importance_score | FLOAT | 知识点重要程度评分（AI 生成） |
| created_at | DATETIME | 知识点创建时间 |

---

## 8. 知识点关系表（knowledge_relation）

用于描述知识点之间的关联关系，构建个人技术知识网络。

| 字段名 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 关系记录唯一标识 |
| from_knowledge_id | BIGINT | 起始知识点 ID |
| to_knowledge_id | BIGINT | 目标知识点 ID |
| relation_type | VARCHAR | 关系类型（prerequisite / related / extension） |

---

## 9. AI 处理日志表（ai_process_log）

用于记录系统中每一次大语言模型参与的处理过程，提升系统的可追踪性与可维护性。

| 字段名 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 日志记录唯一标识 |
| target_type | VARCHAR | 处理对象类型（note / knowledge） |
| target_id | BIGINT | 处理对象 ID |
| process_type | VARCHAR | 处理类型（summarize / extract / embedding） |
| model_name | VARCHAR | 使用的大模型名称 |
| status | VARCHAR | 处理状态（success / failed） |
| created_at | DATETIME | 处理时间 |

> AI 处理日志用于记录每一次模型调用，包括摘要生成、知识点抽取和向量生成。通过日志可追踪失败任务，分析处理结果，并提供系统优化和重试的依据。

---

## 10. 向量表（embedding）

用于存储笔记或知识点的向量表示，支持语义检索和基于 RAG 的智能问答。

| 字段名 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 向量记录唯一标识 |
| target_type | VARCHAR | 向量对应对象类型（note / knowledge） |
| target_id | BIGINT | 向量对应对象 ID |
| vector | vector(1536) | 向量数据（embedding），使用 pgvector 类型存储 1536 维浮点数 |
| model_name | VARCHAR | 使用的向量模型名称 |
| created_at | DATETIME | 向量生成时间 |

> 两种向量类型代表不同粒度：`knowledge` 类型为细粒度知识点向量，`note` 类型为整篇笔记向量。系统检索优先使用 `knowledge` 类型，当未检索到相关内容时，再回退到 `note` 类型，以实现“粗到细”的语义检索策略。

---

## 11. 数据库设计说明总结

- 笔记表存储用户原始学习内容，是系统基础数据来源；  
- 知识点表存储 AI 抽取的核心知识单元，用于知识关联、思维导图和智能复习；  
- 分类表实现粗粒度笔记管理，标签表用于细粒度语义标注；  
- 向量表与 AI 处理日志表支撑系统的语义检索、RAG 问答及可追踪性；  
- PostgreSQL + pgvector 支持向量存储与检索，使系统既可快速开发，也便于后期扩展。

---

# API 接口文档

TechMemo RESTful API 接口设计文档，基于 Gin 框架实现，后续通过 gin-swagger 自动生成在线文档。

**基础信息**
- Base URL: `http://localhost:8080/api/v1`
- 数据格式: JSON
- 字符编码: UTF-8
- 认证方式: JWT Token (Bearer Token)

**统一响应格式**
```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

**状态码说明**
- 200: 请求成功
- 400: 请求参数错误
- 401: 未授权/Token 无效
- 403: 禁止访问
- 404: 资源不存在
- 500: 服务器内部错误

---

## 1. 用户管理模块

### 1.1 用户注册
```
POST /auth/register
```

**请求参数**
```json
{
  "username": "testuser",
  "password": "password123"
}
```

**响应示例**
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": 1,
    "username": "testuser",
    "created_at": "2026-01-13T10:00:00Z"
  }
}
```

### 1.2 用户登录
```
POST /auth/login
```

**请求参数**
```json
{
  "username": "testuser",
  "password": "password123"
}
```

**响应示例**
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user_id": 1,
    "username": "testuser"
  }
}
```

### 1.3 获取当前用户信息
```
GET /auth/profile
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "testuser",
    "created_at": "2026-01-13T10:00:00Z"
  }
}
```

---

## 2. 分类管理模块

### 2.1 获取分类列表
```
GET /categories
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "categories": [
      {
        "id": 1,
        "name": "Python"
      },
      {
        "id": 2,
        "name": "算法"
      }
    ]
  }
}
```

### 2.2 创建分类
```
POST /categories
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "name": "数据库"
}
```

**响应示例**
```json
{
  "code": 200,
  "message": "分类创建成功",
  "data": {
    "id": 3,
    "name": "数据库"
  }
}
```

### 2.3 更新分类
```
PUT /categories/:id
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "name": "数据库设计"
}
```

### 2.4 删除分类
```
DELETE /categories/:id
Authorization: Bearer {token}
```

---

## 3. 标签管理模块

### 3.1 获取标签列表
```
GET /tags
Authorization: Bearer {token}
```

**查询参数**
- `keyword` (可选): 标签名关键词搜索
- `page` (可选): 页码，默认 1
- `page_size` (可选): 每页数量，默认 20

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "tags": [
      {
        "id": 1,
        "name": "装饰器"
      },
      {
        "id": 2,
        "name": "BFS"
      }
    ],
    "total": 10,
    "page": 1,
    "page_size": 20
  }
}
```

### 3.2 创建标签
```
POST /tags
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "name": "缓存"
}
```

### 3.5 更新标签
```
PUT /tags/:id
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "name": "缓存"
}
```

### 3.4 删除标签
```
DELETE /tags/:id
Authorization: Bearer {token}
```

---

## 4. 笔记管理模块

### 4.1 创建笔记
```
POST /notes
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "title": "Python 装饰器学习笔记",
  "content_md": "# 装饰器\n装饰器是 Python 中的一种设计模式...",
  "category_id": 1,
  "tag_ids": [1, 2, 3]
}
```

**响应示例**
```json
{
  "code": 200,
  "message": "笔记创建成功",
  "data": {
    "id": 1,
    "title": "Python 装饰器学习笔记",
    "content_md": "# 装饰器\n装饰器是 Python 中的一种设计模式...",
    "note_type": "tech",
    "category_id": 1,
    "status": "normal",
    "created_at": "2026-01-13T10:30:00Z",
    "updated_at": "2026-01-13T10:30:00Z"
  }
}
```

### 4.2 获取笔记列表
```
GET /notes
Authorization: Bearer {token}
```

**查询参数**
- `category_id` (可选): 分类 ID
- `tag_ids` (可选): 标签 ID 列表，逗号分隔
- `keyword` (可选): 关键词搜索（标题和内容）
- `note_type` (可选): 笔记类型（tech/daily/unknown）
- `page` (可选): 页码，默认 1
- `page_size` (可选): 每页数量，默认 10
- `sort` (可选): 排序方式（created_at_desc/updated_at_desc），默认 created_at_desc

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "notes": [
      {
        "id": 1,
        "title": "Python 装饰器学习笔记",
        "note_type": "tech",
        "category": {
          "id": 1,
          "name": "Python"
        },
        "tags": [
          {"id": 1, "name": "装饰器"},
          {"id": 2, "name": "高级特性"}
        ],
        "created_at": "2026-01-13T10:30:00Z",
        "updated_at": "2026-01-13T10:30:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 10
  }
}
```

### 4.3 获取笔记详情
```
GET /notes/:id
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "title": "Python 装饰器学习笔记",
    "content_md": "# 装饰器\n装饰器是 Python 中的一种设计模式...",
    "note_type": "tech",
    "category": {
      "id": 1,
      "name": "Python"
    },
    "tags": [
      {"id": 1, "name": "装饰器"}
    ],
    "status": "normal",
    "created_at": "2026-01-13T10:30:00Z",
    "updated_at": "2026-01-13T10:30:00Z"
  }
}
```

### 4.4 更新笔记
```
PUT /notes/:id
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "title": "Python 装饰器完整笔记",
  "content_md": "# 装饰器\n更新后的内容...",
  "category_id": 1,
  "tag_ids": [1, 2]
}
```

### 4.5 更新笔记标签
```
PUT /notes/:id/tags
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "tag_ids": [1, 2]
}
```

### 4.6 删除笔记（软删除）
```
DELETE /notes/:id
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "笔记删除成功",
  "data": null
}
```

### 4.7 获取笔记版本历史
```
GET /notes/:id/versions
Authorization: Bearer {token}
```

**查询参数**
- `sort` (可选): 排序方式（created_at_desc/updated_at_desc），默认 created_at_desc

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "versions": [
      {
        "id": 1,
        "note_id": 1,
        "content_md": "旧版本内容...",
        "created_at": "2026-01-12T10:00:00Z"
      },
      {
        "id": 2,
        "note_id": 1,
        "content_md": "新版本内容...",
        "created_at": "2026-01-13T10:30:00Z"
      }
    ]
  }
}
```

### 4.8 恢复笔记到指定版本
```
POST /notes/:id/versions/:version_id/restore
Authorization: Bearer {token}
```

---

## 5. AI 处理模块

### 5.1 触发笔记 AI 处理
```
POST /ai/process/note/:id
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "process_types": ["summarize", "extract", "embedding"]
}
```

**响应示例**
```json
{
  "code": 200,
  "message": "AI 处理任务已提交",
  "data": {
    "task_id": "task_123456",
    "status": "processing"
  }
}
```

### 5.2 获取笔记 AI 摘要
```
GET /notes/:id/summary
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "note_id": 1,
    "summary": "本笔记介绍了 Python 装饰器的基本概念、使用方法和应用场景...",
    "generated_at": "2026-01-13T11:00:00Z"
  }
}
```

### 5.3 从笔记提取知识点
```
POST /notes/:id/extract-knowledge
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "知识点提取成功",
  "data": {
    "knowledge_points": [
      {
        "id": 1,
        "name": "装饰器定义",
        "description": "装饰器是一种设计模式，可以在不修改原函数的情况下增加功能",
        "importance_score": 0.9
      },
      {
        "id": 2,
        "name": "装饰器语法",
        "description": "使用 @decorator_name 语法糖来应用装饰器",
        "importance_score": 0.85
      }
    ]
  }
}
```

---

## 6. 知识点管理模块

### 6.1 获取知识点列表
```
GET /knowledge-points
Authorization: Bearer {token}
```

**查询参数**
- `source_note_id` (可选): 来源笔记 ID
- `keyword` (可选): 关键词搜索
- `min_importance` (可选): 最低重要程度
- `page` (可选): 页码，默认 1
- `page_size` (可选): 每页数量，默认 20

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "knowledge_points": [
      {
        "id": 1,
        "name": "装饰器定义",
        "description": "装饰器是一种设计模式...",
        "source_note_id": 1,
        "source_note_title": "Python 装饰器学习笔记",
        "importance_score": 0.9,
        "created_at": "2026-01-13T11:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

### 6.2 获取知识点详情
```
GET /knowledge-points/:id
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "装饰器定义",
    "description": "装饰器是一种设计模式，可以在不修改原函数的情况下增加功能",
    "source_note_id": 1,
    "source_note_title": "Python 装饰器学习笔记",
    "importance_score": 0.9,
    "related_knowledge": [
      {
        "id": 2,
        "name": "闭包",
        "relation_type": "prerequisite"
      },
      {
        "id": 3,
        "name": "高阶函数",
        "relation_type": "related"
      }
    ],
    "created_at": "2026-01-13T11:00:00Z"
  }
}
```

### 6.3 更新知识点
```
PUT /knowledge-points/:id
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "name": "装饰器定义（更新）",
  "description": "更新后的描述",
  "importance_score": 0.95
}
```

### 6.4 删除知识点
```
DELETE /knowledge-points/:id
Authorization: Bearer {token}
```

### 6.5 创建知识点关系
```
POST /knowledge-relations
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "from_knowledge_id": 1,
  "to_knowledge_id": 2,
  "relation_type": "prerequisite"
}
```

**关系类型说明**
- `prerequisite`: 前置依赖（需要先学习 from 才能理解 to）
- `related`: 相关知识（两个知识点有关联）
- `extension`: 扩展知识（to 是 from 的深入延伸）

---

## 7. 思维导图模块

### 7.1 生成思维导图
```
POST /mindmap/generate
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "note_ids": [1, 2, 3, 5],
  "include_relations": true
}
```

**响应示例**
```json
{
  "code": 200,
  "message": "思维导图生成成功",
  "data": {
    "nodes": [
      {
        "id": "k1",
        "type": "knowledge",
        "name": "装饰器定义",
        "importance_score": 0.9,
        "source_note_id": 1
      },
      {
        "id": "k2",
        "type": "knowledge",
        "name": "闭包",
        "importance_score": 0.85,
        "source_note_id": 2
      }
    ],
    "edges": [
      {
        "from": "k2",
        "to": "k1",
        "relation_type": "prerequisite",
        "label": "前置依赖"
      }
    ]
  }
}
```

### 7.2 获取全局知识图谱
```
GET /mindmap/global
Authorization: Bearer {token}
```

**查询参数**
- `category_id` (可选): 限定分类
- `min_importance` (可选): 最低重要程度阈值

---

## 8. 智能搜索与问答模块

### 8.1 语义搜索笔记
```
POST /search/semantic
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "query": "Python 中如何实现装饰器",
  "search_type": "knowledge",
  "top_k": 5
}
```

**search_type 说明**
- `knowledge`: 搜索知识点
- `note`: 搜索笔记
- `both`: 同时搜索知识点和笔记

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "results": [
      {
        "type": "knowledge",
        "id": 1,
        "name": "装饰器定义",
        "description": "装饰器是一种设计模式...",
        "similarity_score": 0.92,
        "source_note_id": 1
      },
      {
        "type": "note",
        "id": 1,
        "title": "Python 装饰器学习笔记",
        "summary": "本笔记介绍了 Python 装饰器...",
        "similarity_score": 0.88
      }
    ]
  }
}
```

### 8.2 基于笔记的智能问答
```
POST /qa/ask
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "question": "装饰器和闭包有什么关系？",
  "use_rag": true,
  "top_k": 3
}
```

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "answer": "装饰器和闭包密切相关。装饰器本质上是一个返回函数的高阶函数，而闭包是装饰器实现的基础...",
    "sources": [
      {
        "type": "knowledge",
        "id": 1,
        "name": "装饰器定义",
        "source_note_id": 1
      },
      {
        "type": "knowledge",
        "id": 5,
        "name": "闭包概念",
        "source_note_id": 2
      }
    ],
    "model": "gpt-4",
    "generated_at": "2026-01-13T12:00:00Z"
  }
}
```

### 8.3 关键词搜索
```
GET /search/keyword
Authorization: Bearer {token}
```

**查询参数**
- `q`: 搜索关键词
- `search_in`: 搜索范围（title/content/both），默认 both
- `category_id` (可选): 限定分类
- `tag_ids` (可选): 限定标签，逗号分隔
- `page` (可选): 页码
- `page_size` (可选): 每页数量

---

## 9. 复习推荐模块

### 9.1 获取复习推荐列表
```
GET /review/recommendations
Authorization: Bearer {token}
```

**查询参数**
- `count` (可选): 推荐数量，默认 10

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "recommendations": [
      {
        "knowledge_id": 1,
        "knowledge_name": "装饰器定义",
        "importance_score": 0.9,
        "last_review_at": "2026-01-10T10:00:00Z",
        "review_count": 3,
        "reason": "距离上次复习已 3 天，建议复习"
      }
    ]
  }
}
```

### 9.2 记录复习结果
```
POST /review/record
Authorization: Bearer {token}
```

**请求参数**
```json
{
  "knowledge_id": 1,
  "mastered": true
}
```

---

## 10. AI 处理日志模块

### 10.1 获取 AI 处理日志
```
GET /ai/logs
Authorization: Bearer {token}
```

**查询参数**
- `target_type` (可选): note/knowledge
- `target_id` (可选): 目标对象 ID
- `process_type` (可选): 处理类型
- `status` (可选): success/failed
- `page` (可选): 页码
- `page_size` (可选): 每页数量

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "logs": [
      {
        "id": 1,
        "target_type": "note",
        "target_id": 1,
        "process_type": "summarize",
        "model_name": "gpt-4",
        "status": "success",
        "created_at": "2026-01-13T11:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 11. 统计分析模块

### 11.1 获取用户统计数据
```
GET /stats/overview
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total_notes": 100,
    "total_knowledge_points": 250,
    "total_categories": 10,
    "total_tags": 50,
    "recent_notes_count": 5,
    "ai_process_count": 150,
    "last_note_at": "2026-01-13T10:30:00Z"
  }
}
```

### 11.2 获取分类统计
```
GET /stats/categories
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "categories": [
      {
        "category_id": 1,
        "category_name": "Python",
        "note_count": 35,
        "knowledge_count": 80
      },
      {
        "category_id": 2,
        "category_name": "算法",
        "note_count": 40,
        "knowledge_count": 120
      }
    ]
  }
}
```

---

## API 开发说明

### Gin + Swagger 集成步骤

1. **安装依赖**
```bash
go get -u github.com/gin-gonic/gin
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
go get -u github.com/swaggo/swag/cmd/swag
```

2. **API 注释格式示例**
```go
// @Summary 创建笔记
// @Description 创建一条新的技术笔记
// @Tags 笔记管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param note body CreateNoteRequest true "笔记信息"
// @Success 200 {object} Response{data=Note}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /notes [post]
func CreateNote(c *gin.Context) {
    // 实现逻辑
}
```

3. **生成 Swagger 文档**
```bash
swag init
```

4. **访问 API 文档**
```
http://localhost:8080/swagger/index.html
```

### 数据校验

使用 `binding` tag 进行参数校验：
```go
type CreateNoteRequest struct {
    Title      string `json:"title" binding:"required,min=1,max=255"`
    ContentMD  string `json:"content_md" binding:"required"`
    CategoryID int64  `json:"category_id" binding:"required"`
    TagIDs     []int64 `json:"tag_ids"`
}
```

### 中间件

- **JWT 认证中间件**: 验证 Token 有效性
- **日志中间件**: 记录请求日志
- **CORS 中间件**: 处理跨域请求
- **限流中间件**: 防止接口被滥用

---

**文档版本**: v1.0  
**最后更新**: 2026-01-13  
**维护者**: TechMemo Team

