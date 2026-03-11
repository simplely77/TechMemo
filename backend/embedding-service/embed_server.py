import os
from flask import Flask,request,jsonify # pyright: ignore[reportMissingImports]
from sentence_transformers import SentenceTransformer # pyright: ignore[reportMissingImports]

model_name = os.getenv("MODEL_NAME", "all-MiniLM-L6-v2")
model = SentenceTransformer(model_name)

app = Flask(__name__)

@app.route("/embed",methods=["POST"])
def embed():
    """
    POST body:
    {
        "texts": ["文本1","文本二"]
    }
    返回:
    {
        "vectors": [[0.01,0.23, ...], [...]]
    }
    """
    data = request.get_json()
    texts = data.get("texts",[])
    if not texts:
        return jsonify({"error":"no texts provided"}),400
    
    vectors = model.encode(texts).tolist()
    return jsonify({"vectors": vectors})

if __name__ == "__main__":
    print("Embedding service started...")
    app.run(host="0.0.0.0",port=5000)