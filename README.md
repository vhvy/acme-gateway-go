# WARNING
本人是 `golang` 初学者，对 `https` 证书颁发的完整流程也不甚清楚，如果你要使用本项目，带来的一切后果请自负!

# 绝对不要在公网部署该服务！仅限局域网使用！
# 绝对不要在公网部署该服务！仅限局域网使用！
# 绝对不要在公网部署该服务！仅限局域网使用！

### ACME-GATEWAY-GO

使用 lego 开发的 `https` 证书生成网关，可以通过 api 调用接口实现自动生成 https 证书，目前仅支持 `cloudflare dns01` 的验证方式。

### 使用方法
目前仅支持 Docker
新建 `compose.yaml` 文件，内容如下：
```docker
services:
  acme-gateway-go:
    image: ghcr.io/vhvy/acme-gateway-go:master
    container_name: acme-gateway-go
    restart: unless-stopped
    env_file: .env
    volumes:
      - ./data:/app/data
      - /etc/localtime:/etc/localtime:ro
    ports:
      - 48952:48952
    cap_add:
      - NET_ADMIN
```
然后在同级目录新建 `.env` 文件，内容如下：
```env
CLOUDFLARE_DNS_API_TOKEN=
GATEWAY_TOKEN=
ACME_EMAIL=
```
+ CLOUDFLARE_DNS_API_TOKEN
  + `Cloudflare` 的 API TOKEN
+ GATEWAY_TOKEN
  + 调用本网关时的密钥，自定义
+ ACME_EMAIL
  + 用于注册 ACME 账户的邮箱地址
  
配置完成后，执行 `docker compose up -d`即可开始运行。

### Q&A
1. 如何调用接口？
```bash
curl http://127.0.0.1:48952/renew?domain=f.example.com&format=json
```

支持的参数：
  + domain
    + `domain`支持多个域名，例如`domain=f.example.com,x.example.com,z.example.com`。
    + 注意多域名不是生成多个证书，而是为第一个域名`f.example.com`生成证书，其他域名包含在`f.example.com`的`subject alternative name`中，该证书可以给提交的所有域名使用。

  + format
    + 默认输出纯文本证书，格式为证书 + 私钥
    + json
      + 将证书和私钥放在 json 的不同字段中返回
      + example: { "certificate": "xxx", "private_key": "xxx" }

2. 能不能再具体点？
  + 参考如下脚本。

```shell
#!/bin/bash

DOMAINS=$(find ./nginx.conf.d -name '*.conf' \
  | xargs awk '/server_name/ {for (i=2; i<=NF; i++) print $i}' \
  | sed 's/;//' \
  | sort -u \
  | paste -sd, -)

echo $DOMAINS

# API 地址
API_URL="http://127.0.0.1:48952/renew?domain=f.example.com,$DOMAINS&format=json"

# 自定义 Header
CUSTOM_HEADER="verify-token: YOUR_GATEWAY_TOKEN"

# 发送 GET 请求获取 JSON，并携带自定义 Header
json_response=$(curl -s -H "$CUSTOM_HEADER" "$API_URL")

# 检查 curl 命令是否成功
if [ $? -ne 0 ]; then
    echo "Error: Failed to fetch data from $API_URL"
    exit 1
fi

# 使用 jq 提取 cert 字段
cert_content=$(echo "$json_response" | jq -r '.certificate')

# 检查 cert 字段是否为空
if [ -z "$cert_content" ]; then
    echo "Error: 'cert' field not found or is empty in the JSON response."
    exit 1
fi

# 使用 jq 提取 key 字段
key_content=$(echo "$json_response" | jq -r '.private_key')

# 检查 key 字段是否为空
if [ -z "$key_content" ]; then
    echo "Error: 'key' field not found or is empty in the JSON response."
    exit 1
fi

cd /root/nginx/certs/certificates

echo "$cert_content" > cert.pem
echo "$key_content" > key.pe
```