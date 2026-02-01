import secrets

# 生成 32 字节的密钥
secret_key = secrets.token_urlsafe(32)
print(secret_key)

# 或者生成 hex 格式
secret_key_hex = secrets.token_hex(32)
print(secret_key_hex)