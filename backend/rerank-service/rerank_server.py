from fastapi import FastAPI
from pydantic import BaseModel
from typing import List
from sentence_transformers import CrossEncoder
import uvicorn

# 初始化 FastAPI
app = FastAPI()

# 加载 rerank 模型（推荐这个）
model = CrossEncoder("BAAI/bge-reranker-base")

# 请求结构
class RerankRequest(BaseModel):
    query: str
    documents: List[str]

# 响应结构
class RerankResponse(BaseModel):
    scores: List[float]

@app.post("/rerank", response_model=RerankResponse)
def rerank(req: RerankRequest):
    query = req.query
    docs = req.documents

    # 构造输入：[(query, doc), ...]
    pairs = [(query, doc) for doc in docs]

    # 批量预测
    scores = model.predict(pairs)

    return {"scores": scores.tolist()}

# 启动服务
if __name__ == "__main__":
    uvicorn.run("rerank_server:app", host="0.0.0.0", port=8001)