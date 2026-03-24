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

### 场景三：AI 自动提取知识点
- **用户目标**：从笔记中自动提取结构化知识点
- **用户行为**：
  - 创建或编辑笔记后触发 AI 处理
  - 查看 AI 提取的知识点列表
  - 查看知识点之间的关联关系
- **期望功能**：
  - 自动识别笔记中的核心知识点
  - 生成知识点描述和重要性评分
  - 建立知识点之间的关联关系
  - 支持向量化检索

### 场景四：思维导图可视化
- **用户目标**：通过思维导图可视化知识结构
- **用户行为**：
  - 查看单个笔记的思维导图
  - 生成全局知识图谱
  - 查看任务处理状态
- **期望功能**：
  - 基于知识点生成思维导图
  - 展示知识点之间的层级关系
  - 支持全局知识图谱生成

### 场景五：智能问答系统
- **用户目标**：通过自然语言提问，快速获取基于笔记的答案
- **用户行为**：
  - 输入技术问题
  - 查看系统生成答案
  - 查看答案引用的笔记
- **期望功能**：
  - 基于 RAG 的智能问答
  - 优先使用用户已有笔记生成答案
  - 支持查看答案来源和相关知识点
  - 语义搜索笔记和知识点

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
| vector | vector(384) | 向量数据（embedding），使用 pgvector 类型存储 384 维浮点数 |
| model_name | VARCHAR | 使用的向量模型名称 |
| created_at | DATETIME | 向量生成时间 |

> 两种向量类型代表不同粒度：`knowledge` 类型为细粒度知识点向量，`note` 类型为整篇笔记向量。系统检索优先使用 `knowledge` 类型，当未检索到相关内容时，再回退到 `note` 类型，以实现”粗到细”的语义检索策略。

---

## 11. 笔记根节点表（note_root_node）

用于存储笔记思维导图的根节点信息。

| 字段名 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 根节点唯一标识 |
| note_id | BIGINT | 关联的笔记 ID |
| root_knowledge_id | BIGINT | 根知识点 ID |
| created_at | DATETIME | 创建时间 |

---

## 12. 数据库设计说明总结

- 笔记表存储用户原始学习内容，是系统基础数据来源；
- 知识点表存储 AI 抽取的核心知识单元，用于知识关联、思维导图和智能复习；
- 分类表实现粗粒度笔记管理，标签表用于细粒度语义标注；
- 向量表与 AI 处理日志表支撑系统的语义检索、RAG 问答及可追踪性；
- 笔记根节点表用于存储思维导图的根节点信息；
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
  "note_type": "knowledge"
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
POST /ai/note/:id
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "AI 处理任务已提交",
  "data": {
    "note_id": 1,
    "task_id": "task_123456",
    "status": "pending"
  }
}
```

### 5.2 获取笔记 AI 处理状态
```
GET /ai/note/:id/status
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "note_id": 1,
    "status": "completed",
    "progress": {
      "extract": "completed",
      "embedding": "completed"
    }
  }
}
```

### 5.3 触发全局思维导图生成
```
POST /ai/mindmap/global
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "全局思维导图生成任务已提交",
  "data": {
    "task_id": "task_789012",
    "status": "pending"
  }
}
```

### 5.4 查询任务状态
```
GET /ai/task/:task_id/status
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "task_789012",
    "status": "completed",
    "result": {}
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

---

## 7. 思维导图模块

### 7.1 获取笔记思维导图
```
GET /mindmap?note_id=:id
Authorization: Bearer {token}
```

**查询参数**
- `note_id` (必填): 笔记 ID

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "note_id": 1,
    "root_node": {
      "id": 1,
      "name": "Python 装饰器",
      "children": [
        {
          "id": 2,
          "name": "装饰器定义",
          "relation_type": "prerequisite"
        },
        {
          "id": 3,
          "name": "闭包",
          "relation_type": "related"
        }
      ]
    }
  }
}
```

### 7.2 获取全局知识图谱
```
GET /mindmap/global
Authorization: Bearer {token}
```

**响应示例**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "nodes": [
      {
        "id": 1,
        "name": "Python 装饰器",
        "type": "knowledge_point"
      }
    ],
    "edges": [
      {
        "from": 1,
        "to": 2,
        "relation_type": "prerequisite"
      }
    ]
  }
}
```

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

---

## 9. 统计分析模块（待实现）

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
    // "recent_notes_count": 5,
    // "ai_process_count": 150,
    // "last_note_at": "2026-01-13T10:30:00Z"
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

## 快速开始

### 环境要求
- Go 1.24+
- PostgreSQL 14+ (带 pgvector 扩展)
- Redis 7+
- Docker & Docker Compose (推荐)

### 使用 Docker Compose 启动

```bash
# 启动所有服务（PostgreSQL, Redis, Embedding Service）
docker-compose up -d

# 进入 backend 目录
cd backend

# 安装依赖
go mod tidy

# 运行后端服务
go run main.go
```

服务启动后：
- 后端 API: http://localhost:8080
- Swagger 文档: http://localhost:8080/swagger-ui/index.html
- 健康检查: http://localhost:8080/health

### 开发说明

详细的开发指南请参考 [backend/README.md](backend/README.md)，包括：
- 项目结构说明
- 分层架构设计
- 添加新 API 的步骤
- GORM Gen 代码生成
- Swagger 文档生成
- 中间件使用

---

**文档版本**: v1.0
**最后更新**: 2026-03-13
**维护者**: TechMemo Team


# 第三周

第一天：完成统计模块的后端实现，对智慧问答系统进行深入分析，敲定实现方案

# 一些思考
## 智慧问答模块的思考
智慧问答模块是通过将用户问题抽取成向量去向量库找相关向量然后拿到向量对应数据传给大模型，让大模型生成相关回答，考虑做知识点，笔记入口

一些问题：
首先deepseek的免费模型是不支持长期记忆的，想要打造具有长期记忆的问答系统需要服务端存储历史消息，进而引发第二个问题，大模型的token限制导致不可能将完整的历史消息传递给大模型

两种解决方案：
方案一：
  当记忆达到阈值时，调用大模型，将记忆传递给模型，让它进行压缩，然后传递压缩记忆，与最近的消息给大模型
方案二：
  利用已经建好的向量库，将大模型的每次回答进行向量化，将用户输入向量化之后，在向量库搜寻topk的大模型回复，找到对应消息和最近的消息一同喂给大模型，
  衍生问题：
    方案二使用topk来找到对应的回复还是有可能会超过大模型的token阈值
  解决方案：
    1.还是让大模型自己来压缩达到阈值的记忆，，然后将压缩后的记忆与近期消息传给大模型，还是会有方案一所说的额外调用一次的延迟，只不过提高了精度
    2.动态上下文裁剪，使用topk来动态的控制上下文，当达到阈值后，降低topk的值，只取强相关的回复，可能会丢掉一些有用的上下文

方案二相对于方案一来说精度会有比较大的提升，但是实现相对复杂，同时为避免增加延迟，使用异步队列来抽取大模型回复向量，减少使用大模型压缩自身记忆的一次大模型调用延迟，决定使用方案二，解决方案也选方案二

第二天：
     ╭───────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
     │ RAG 智慧问答模块实现计划                                                                                          │
     │                                                                                                                   │
     │ Context                                                                                                           │
     │                                                                                                                   │
     │ 用户希望实现一个基于 RAG（检索增强生成）的智慧问答系统，解决 DeepSeek 免费模型不支持长期记忆和 token 限制的问题。 │
     │                                                                                                                   │
     │ 选定方案： 方案二 + 动态上下文裁剪                                                                                │
     │ - 将大模型的每次回答向量化存储                                                                                    │
     │ - 用户输入向量化后，在向量库搜寻 topK 的相关回复                                                                  │
     │ - 使用动态上下文裁剪控制 token 数量（激进策略，优先响应速度）                                                     │
     │ - 同时搜索聊天历史和知识库（notes + knowledge_points）                                                            │
     │                                                                                                                   │
     │ 现有基础设施：                                                                                                    │
     │ - ✅ pgvector 向量存储（384维，cosine distance）                                                                  │
     │ - ✅ embedding 服务（all-MiniLM-L6-v2）                                                                           │
     │ - ✅ 语义搜索功能（dao/search.go）                                                                                │
     │ - ✅ 异步队列系统（Redis/Memory）                                                                                 │
     │ - ✅ DeepSeek LLM 集成                                                                                            │
     │                                                                                                                   │
     │ ---                                                                                                               │
     │ 核心算法                                                                                                          │
     │                                                                                                                   │
     │ RAG 上下文构建流程                                                                                                │
     │                                                                                                                   │
     │ 1. 用户发送消息 → 向量化（同步）                                                                                  │
     │ 2. 语义搜索：                                                                                                     │
     │    - 聊天历史中的相似消息（topK=5）                                                                               │
     │    - 知识库中的相关笔记（topK=5）                                                                                 │
     │    - 知识库中的相关知识点（topK=5）                                                                               │
     │ 3. 获取最近 6 条对话历史                                                                                          │
     │ 4. 合并上下文：                                                                                                   │
     │    - 知识库上下文（system role）                                                                                  │
     │    - 语义相关的历史消息（去重最近消息）                                                                           │
     │    - 最近消息（时间顺序）                                                                                         │
     │    - 当前用户查询                                                                                                 │
     │ 5. 动态裁剪（token 阈值 4000）：                                                                                  │
     │    - 如果超限，逐步减少 topK（最小值 2）                                                                          │
     │    - 仍超限则截断知识库上下文                                                                                     │
     │ 6. 调用 LLM 生成回复                                                                                              │
     │ 7. 立即返回响应                                                                                                   │
     │ 8. 异步：将 AI 回复入队，向量化并存储                                                                             │
     │                                                                                                                   │
     │ Token 估算策略                                                                                                    │
     │                                                                                                                   │
     │ - 官方估算：1 个英文字符 ≈ 0.3 token，1 个中文字符 ≈ 0.6 token                                                    │
     │ - 实现：逐字符判断是否为中文（Unicode 范围），分别累加估算值                                                      │
     │                                                                                                                   │
     │ ---                                                                                                               │
     │ 数据库设计                                                                                                        │
     │                                                                                                                   │
     │ 新增表                                                                                                            │
     │                                                                                                                   │
     │ -- 聊天会话表                                                                                                     │
     │ CREATE TABLE chat_session (                                                                                       │
     │     id BIGSERIAL PRIMARY KEY,                                                                                     │
     │     user_id BIGINT NOT NULL,                                                                                      │
     │     title VARCHAR(255) DEFAULT '新对话',                                                                          │
     │     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,                                                               │
     │     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP                                                                │
     │ );                                                                                                                │
     │                                                                                                                   │
     │ -- 聊天消息表                                                                                                     │
     │ CREATE TABLE chat_message (                                                                                       │
     │     id BIGSERIAL PRIMARY KEY,                                                                                     │
     │     session_id BIGINT NOT NULL,                                                                                   │
     │     user_id BIGINT NOT NULL,                                                                                      │
     │     role VARCHAR(20) NOT NULL,  -- 'user' / 'assistant'                                                           │
     │     content TEXT NOT NULL,                                                                                        │
     │     token_count INT DEFAULT 0,                                                                                    │
     │     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,                                                               │
     │     FOREIGN KEY (session_id) REFERENCES chat_session(id) ON DELETE CASCADE                                        │
     │ );                                                                                                                │
     │                                                                                                                   │
     │ CREATE INDEX idx_chat_message_session_id ON chat_message(session_id);                                             │
     │ CREATE INDEX idx_chat_message_created_at ON chat_message(created_at DESC);                                        │
     │ CREATE INDEX idx_chat_message_session_created ON chat_message(session_id, created_at DESC);                       │
     │                                                                                                                   │
     │ 扩展现有表                                                                                                        │
     │                                                                                                                   │
     │ - embedding 表：支持 target_type='chat_message'（无需修改表结构）                                                 │
     │ - ai_process_log 表：支持 process_type='chat_embedding'（无需修改表结构）                                         │
     │                                                                                                                   │
     │ ---                                                                                                               │
     │ 文件结构                                                                                                          │
     │                                                                                                                   │
     │ 新增文件                                                                                                          │
     │                                                                                                                   │
     │ backend/                                                                                                          │
     │ ├── model/                                                                                                        │
     │ │   ├── chat_session.gen.go          # GORM Gen 生成                                                              │
     │ │   └── chat_message.gen.go          # GORM Gen 生成                                                              │
     │ ├── dao/                                                                                                          │
     │ │   └── chat.go                       # 聊天数据访问层                                                            │
     │ ├── service/                                                                                                      │
     │ │   └── chat.go                       # 聊天服务层（核心逻辑）                                                    │
     │ ├── handler/                                                                                                      │
     │ │   └── chat.go                       # 聊天 API handler                                                          │
     │ ├── handler/dto/                                                                                                  │
     │ │   └── chat.go                       # 聊天 DTO                                                                  │
     │ └── ai/client/                                                                                                    │
     │     └── types.go                      # 在现有文件中添加 ChatMessage 类型                                         │
     │                                                                                                                   │
     │ 修改文件                                                                                                          │
     │                                                                                                                   │
     │ backend/                                                                                                          │
     │ ├── dao/                                                                                                          │
     │ │   ├── ai.go                         # 支持 chat_message embedding                                               │
     │ │   └── search.go                     # 支持搜索 chat_message                                                     │
     │ ├── service/                                                                                                      │
     │ │   └── ai.go                         # handleEmbedding 支持 chat_message                                         │
     │ ├── ai/client/                                                                                                    │
     │ │   ├── types.go                      # 添加 ChatMessage 类型                                                     │
     │ │   ├── client.go                     # 添加 ChatWithContext 方法                                                 │
     │ │   └── openai.go                     # 实现 ChatWithContext                                                      │
     │ ├── bootstrap/                                                                                                    │
     │ │   └── app.go                        # 注册 ChatService                                                          │
     │ ├── router/                                                                                                       │
     │ │   └── router.go                     # 注册聊天路由                                                              │
     │ └── cmd/gen/                                                                                                      │
     │     └── main.go                       # 生成新模型                                                                │
     │                                                                                                                   │
     │ ---                                                                                                               │
     │ API 设计                                                                                                          │
     │                                                                                                                   │
     │ POST   /api/v1/chat/sessions                    # 创建会话                                                        │
     │ GET    /api/v1/chat/sessions                    # 获取会话列表                                                    │
     │ DELETE /api/v1/chat/sessions/:id                # 删除会话                                                        │
     │ POST   /api/v1/chat/sessions/:id/messages       # 发送消息（核心）                                                │
     │ GET    /api/v1/chat/sessions/:id/messages       # 获取消息历史                                                    │
     │                                                                                                                   │
     │ ---                                                                                                               │
     │ 实现步骤                                                                                                          │
     │                                                                                                                   │
     │ 阶段 1：数据库和模型（完成）                                                                                              │
     │                                                                                                                   │
     │ 1. 在 sql/init.sql 或创建新的迁移文件，添加表结构                                                                 │
     │ 2. 修改 backend/cmd/gen/main.go，添加 chat_session 和 chat_message 的生成配置                                     │
     │ 3. 运行 go run cmd/gen/main.go 生成模型                                                                           │
     │                                                                                                                   │
     │ 阶段 2：DAO 层（完成）                                                                                                    │
     │                                                                                                                   │
     │ 1. 创建 backend/dao/chat.go：                                                                                     │
     │   - CreateSession, GetSessionsByUserID, DeleteSession                                                             │
     │   - CreateMessage, GetMessagesBySessionID, GetRecentMessages                                                      │
     │   - GetMessageByID, GetMessagesByIDs                                                                              │
     │ 2. 修改 backend/dao/search.go：                                                                                   │
     │   - SearchEmbeddingsByVector 支持 target_type='chat_message'                                                      │
     │   - 添加 JOIN chat_message 表的查询逻辑                                                                           │
     │                                                                                                                   │
     │ 阶段 3：AI Client 扩展（完成）                                                                                            │
     │                                                                                                                   │
     │ 1. 修改 backend/ai/client/types.go：                                                                              │
     │   - 添加 ChatMessage 结构体（Role, Content）                                                                      │
     │ 2. 修改 backend/ai/client/client.go：                                                                             │
     │   - 在 AIClient 接口添加 ChatWithContext(messages []ChatMessage) (string, error)                                  │
     │ 3. 修改 backend/ai/client/openai.go：                                                                             │
     │   - 实现 ChatWithContext 方法，转换为 OpenAI 格式并调用                                                           │
     │                                                                                                                   │
     │ 阶段 4：Service 层（核心）                                                                                        │
     │                                                                                                                   │
     │ 1. 创建 backend/service/chat.go：                                                                                 │
     │   - CreateSession, GetSessions, DeleteSession                                                                     │
     │   - SendMessage（核心方法）：                                                                                     │
     │ a. 保存用户消息                                                                                                   │
     │ b. 向量化用户消息（同步）                                                                                         │
     │ c. 调用 buildRAGContext 构建上下文                                                                                │
     │ d. 调用 ChatWithContext 生成回复                                                                                  │
     │ e. 保存 AI 回复                                                                                                   │
     │ f. 异步入队向量化任务                                                                                             │
     │ g. 返回响应                                                                                                       │
     │   - GetMessages                                                                                                   │
     │   - buildRAGContext（核心算法）：                                                                                 │
     │ a. 语义搜索聊天历史                                                                                               │
     │ b. 语义搜索知识库（notes + knowledge_points）                                                                     │
     │ c. 获取最近消息                                                                                                   │
     │ d. 合并上下文                                                                                                     │
     │ e. 动态裁剪                                                                                                       │
     │   - estimateTokens（token 估算）                                                                                  │
     │   - enqueueMessageEmbedding（入队向量化任务）                                                                     │
     │ 2. 修改 backend/service/ai.go：                                                                                   │
     │   - handleEmbedding 方法添加 case "chat_message" 分支                                                             │
     │   - 调用 chatDao.GetMessageByID 获取消息内容                                                                      │
     │   - 向量化并存储                                                                                                  │
     │                                                                                                                   │
     │ 阶段 5：Handler 和 DTO                                                                                            │
     │                                                                                                                   │
     │ 1. 创建 backend/handler/dto/chat.go：                                                                             │
     │   - SendMessageReq, GetMessagesReq                                                                                │
     │   - ChatSessionResp, ChatSessionListResp                                                                          │
     │   - ChatMessageResp, ChatMessageListResp                                                                          │
     │ 2. 创建 backend/handler/chat.go：                                                                                 │
     │   - HandlerCreateSession                                                                                          │
     │   - HandlerGetSessions                                                                                            │
     │   - HandlerDeleteSession                                                                                          │
     │   - HandlerSendMessage（核心）                                                                                    │
     │   - HandlerGetMessages                                                                                            │
     │   - 添加完整的 Swagger 注解                                                                                       │
     │                                                                                                                   │
     │ 阶段 6：集成和路由                                                                                                │
     │                                                                                                                   │
     │ 1. 修改 backend/bootstrap/app.go：                                                                                │
     │   - 初始化 ChatDao                                                                                                │
     │   - 初始化 ChatService（注入所有依赖）                                                                            │
     │   - 添加到 App 结构体                                                                                             │
     │ 2. 修改 backend/router/router.go：                                                                                │
     │   - 在 authorized 路由组添加 /chat 路由                                                                           │
     │   - 注册所有聊天端点                                                                                              │
     │ 3. 运行 swag init 更新 Swagger 文档                                                                               │
     │                                                                                                                   │
     │ 阶段 7：测试                                                                                                      │
     │                                                                                                                   │
     │ 1. 启动服务：docker-compose up -d && go run main.go                                                               │
     │ 2. 测试流程：                                                                                                     │
     │   - 创建会话                                                                                                      │
     │   - 发送消息（验证 RAG 上下文构建）                                                                               │
     │   - 查看消息历史                                                                                                  │
     │   - 验证异步向量化（检查 embedding 表）                                                                           │
     │   - 测试动态裁剪（发送长对话）                                                                                    │
     │ 3. 验证知识库集成（发送与笔记相关的问题）                                                                         │
     │                                                                                                                   │
     │ ---                                                                                                               │
     │ 关键文件路径                                                                                                      │
     │                                                                                                                   │
     │ - /Users/liyi/ejoy/school/TechMemo/backend/service/chat.go - 核心业务逻辑                                         │
     │ - /Users/liyi/ejoy/school/TechMemo/backend/dao/chat.go - 数据访问层                                               │
     │ - /Users/liyi/ejoy/school/TechMemo/backend/dao/search.go - 修改语义搜索                                           │
     │ - /Users/liyi/ejoy/school/TechMemo/backend/service/ai.go - 修改 embedding 处理                                    │
     │ - /Users/liyi/ejoy/school/TechMemo/backend/handler/chat.go - API handlers                                         │
     │ - /Users/liyi/ejoy/school/TechMemo/backend/handler/dto/chat.go - DTO 定义                                         │
     │ - /Users/liyi/ejoy/school/TechMemo/backend/ai/client/types.go - 添加 ChatMessage 类型                             │
     │ - /Users/liyi/ejoy/school/TechMemo/backend/bootstrap/app.go - 依赖注入                                            │
     │ - /Users/liyi/ejoy/school/TechMemo/backend/router/router.go - 路由注册                                            │
     │                                                                                                                   │
     │ ---                                                                                                               │
     │ 技术要点                                                                                                          │
     │                                                                                                                   │
     │ 动态裁剪策略（激进）                                                                                              │
     │                                                                                                                   │
     │ - Token 阈值：4000（优先响应速度）                                                                                │
     │ - 初始 topK：5，最小 topK：2                                                                                      │
     │ - 裁剪顺序：语义搜索结果 → 知识库上下文 → 保留最近消息                                                            │
     │                                                                                                                   │
     │ 异步处理                                                                                                          │
     │                                                                                                                   │
     │ - 用户消息：同步向量化（保证搜索准确性）                                                                          │
     │ - AI 回复：异步向量化（不阻塞响应）                                                                               │
     │ - 复用现有 ai/queue 和 ai/worker 系统                                                                             │
     │                                                                                                                   │
     │ 安全性                                                                                                            │
     │                                                                                                                   │
     │ - 权限检查：用户只能访问自己的会话和消息                                                                          │
     │ - 输入验证：消息长度限制（max 5000 字符）                                                                         │
     │ - SQL 注入防护：使用 GORM 参数化查询                                                                              │
     │                                                                                                                   │
     │ 性能优化                                                                                                          │
     │                                                                                                                   │
     │ - 批量查询：GetMessagesByIDs, GetNotesByIDs, GetKnowledgePointsByIDs                                              │
     │ - 索引优化：chat_message(session_id, created_at DESC)                                                             │
     │ - 向量搜索阈值：chat_message(0.3), note(0.4), knowledge(0.4)                                                      │
     │                                                                                                                   │
     │ ---                                                                                                               │
     │ 验证方式                                                                                                          │
     │                                                                                                                   │
     │ 1. 功能验证：                                                                                                     │
     │   - 创建会话并发送消息                                                                                            │
     │   - 验证 AI 回复包含知识库内容                                                                                    │
     │   - 验证长对话的动态裁剪                                                                                          │
     │ 2. 数据验证：                                                                                                     │
     │   - 检查 chat_message 表有用户和 AI 消息                                                                          │
     │   - 检查 embedding 表有 target_type='chat_message' 的记录                                                         │
     │   - 检查 ai_process_log 表有 process_type='chat_embedding' 的日志                                                 │
     │ 3. 性能验证：                                                                                                     │
     │   - 响应时间 < 3 秒（不含 LLM 调用时间）                                                                          │
     │   - 异步向量化不阻塞响应                                                                                          │
     │   - 语义搜索返回相关结果                                                                                          │
     │ 4. 边界测试：                                                                                                     │
     │   - 空会话发送消息                                                                                                │
     │   - 超长消息（5000+ 字符）                                                                                        │
     │   - 大量历史消息（100+ 条） 