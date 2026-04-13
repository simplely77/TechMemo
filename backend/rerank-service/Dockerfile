# 使用轻量 Python 镜像
FROM python:3.10-slim

# 设置工作目录
WORKDIR /app

# 防止缓存 & 加快安装
ENV PIP_NO_CACHE_DIR=1

# 安装系统依赖（避免 torch / transformers 报错）
RUN apt-get update && apt-get install -y \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 拷贝依赖文件
COPY requirements.txt .

# 安装 Python 依赖
RUN pip install --upgrade pip && \
    pip install -r requirements.txt

# 拷贝代码
COPY rerank_server.py .

# 下载模型（提前缓存，避免启动慢）
RUN python -c "from sentence_transformers import CrossEncoder; CrossEncoder('BAAI/bge-reranker-base')"

# 暴露端口
EXPOSE 8001

# 启动服务
CMD ["uvicorn", "rerank_server:app", "--host", "0.0.0.0", "--port", "8001"]