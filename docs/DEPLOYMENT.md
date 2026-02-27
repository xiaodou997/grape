# Grape 部署指南

本文档介绍如何在不同环境中部署 Grape。

## 目录

- [系统要求](#系统要求)
- [单机部署](#单机部署)
- [Systemd 服务部署](#systemd-服务部署)
- [Docker 部署](#docker-部署)
- [Docker Compose 部署](#docker-compose-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [Nginx 反向代理](#nginx-反向代理)
- [HTTPS 配置](#https-配置)
- [备份与恢复](#备份与恢复)
- [监控与告警](#监控与告警)
- [故障排查](#故障排查)

---

## 系统要求

### 最低配置

| 资源 | 要求 |
|------|------|
| CPU | 1 核 |
| 内存 | 64MB (推荐 128MB+) |
| 磁盘 | 1GB+ (根据包数量) |
| 操作系统 | Linux / macOS / Windows |

### 推荐配置 (小型团队)

| 资源 | 要求 |
|------|------|
| CPU | 2 核 |
| 内存 | 256MB |
| 磁盘 | 10GB+ SSD |
| 操作系统 | Linux (Ubuntu 20.04+ / CentOS 7+) |

### 网络要求

| 端口 | 用途 | 说明 |
|------|------|------|
| 4873 | HTTP 服务 | 默认端口，可自定义 |
| - | 上游访问 | 需要能访问公共 npm 仓库 |

---

## 单机部署

### 1. 下载二进制

```bash
# macOS (Intel)
curl -sL https://github.com/graperegistry/grape/releases/latest/download/grape-darwin-amd64 -o grape
chmod +x grape

# macOS (Apple Silicon)
curl -sL https://github.com/graperegistry/grape/releases/latest/download/grape-darwin-arm64 -o grape
chmod +x grape

# Linux (amd64)
curl -sL https://github.com/graperegistry/grape/releases/latest/download/grape-linux-amd64 -o grape
chmod +x grape

# Linux (arm64)
curl -sL https://github.com/graperegistry/grape/releases/latest/download/grape-linux-arm64 -o grape
chmod +x grape
```

### 2. 创建目录结构

```bash
# 创建安装目录
sudo mkdir -p /opt/grape
sudo cp grape /opt/grape/grape

# 创建配置目录
sudo mkdir -p /etc/grape

# 创建数据目录
sudo mkdir -p /var/lib/grape/data
sudo chown -R $USER:$USER /var/lib/grape
```

### 3. 创建配置文件

```bash
sudo tee /etc/grape/config.yaml > /dev/null << 'EOF'
server:
  host: "0.0.0.0"
  port: 4873

log:
  level: "info"

auth:
  jwt_secret: "change-this-to-a-secure-random-string"
  jwt_expiry: 24h
  allow_registration: false

storage:
  path: "/var/lib/grape/data"

database:
  dsn: "/var/lib/grape/data/grape.db"
EOF
```

### 4. 启动服务

```bash
# 测试启动
/opt/grape/grape -c /etc/grape/config.yaml

# 后台运行
nohup /opt/grape/grape -c /etc/grape/config.yaml > /var/log/grape.log 2>&1 &

# 验证运行
ps aux | grep grape
curl http://localhost:4873/-/health
```

---

## Systemd 服务部署

### 1. 创建服务文件

```bash
sudo tee /etc/systemd/system/grape.service > /dev/null << 'EOF'
[Unit]
Description=Grape npm Registry
Documentation=https://github.com/graperegistry/grape
After=network.target

[Service]
Type=simple
User=grape
Group=grape
WorkingDirectory=/var/lib/grape
ExecStart=/opt/grape/grape -c /etc/grape/config.yaml
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

# 安全加固
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/grape

# 环境变量
Environment="GRAPE_LOG_LEVEL=info"

[Install]
WantedBy=multi-user.target
EOF
```

### 2. 创建用户和目录

```bash
# 创建系统用户
sudo useradd -r -s /bin/false -d /var/lib/grape grape

# 创建目录并设置权限
sudo mkdir -p /var/lib/grape/data
sudo chown -R grape:grape /var/lib/grape
sudo chmod 750 /var/lib/grape
```

### 3. 启用并启动服务

```bash
# 重载 systemd
sudo systemctl daemon-reload

# 启用服务（开机自启）
sudo systemctl enable grape

# 启动服务
sudo systemctl start grape

# 查看状态
sudo systemctl status grape

# 查看日志
sudo journalctl -u grape -f
```

### 4. 常用命令

```bash
# 停止服务
sudo systemctl stop grape

# 重启服务
sudo systemctl restart grape

# 重新加载配置（需要重启服务）
sudo systemctl restart grape

# 禁用开机自启
sudo systemctl disable grape
```

---

## Docker 部署

### 1. 拉取镜像

```bash
docker pull graperegistry/grape:latest
```

### 2. 运行容器

```bash
docker run -d \
  --name grape \
  -p 4873:4873 \
  -v grape-data:/data \
  -v ./config.yaml:/etc/grape/config.yaml:ro \
  -e GRAPE_AUTH_JWT_SECRET="your-secret-key" \
  --restart unless-stopped \
  graperegistry/grape:latest
```

### 3. 验证运行

```bash
# 查看日志
docker logs -f grape

# 健康检查
curl http://localhost:4873/-/health
```

### 4. 停止和删除

```bash
# 停止
docker stop grape

# 启动
docker start grape

# 删除（数据卷保留）
docker rm grape
```

---

## Docker Compose 部署

### docker-compose.yml

```yaml
version: '3.8'

services:
  grape:
    image: graperegistry/grape:latest
    container_name: grape
    ports:
      - "4873:4873"
    volumes:
      - grape-data:/data
      - ./config.yaml:/etc/grape/config.yaml:ro
    environment:
      - GRAPE_AUTH_JWT_SECRET=${GRAPE_AUTH_JWT_SECRET}
      - GRAPE_LOG_LEVEL=info
    command: -c /etc/grape/config.yaml
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:4873/-/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

volumes:
  grape-data:
    driver: local
```

### .env 文件

```bash
# .env
GRAPE_AUTH_JWT_SECRET=your-secure-random-string-here
```

### 启动命令

```bash
# 启动
docker-compose up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止
docker-compose down

# 重启
docker-compose restart
```

### 多实例部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  grape-1:
    <<: *grape-base
    container_name: grape-1
    ports:
      - "4873:4873"

  grape-2:
    <<: *grape-base
    container_name: grape-2
    ports:
      - "4874:4873"

x-grape-base: &grape-base
  image: graperegistry/grape:latest
  volumes:
    - shared-data:/data
    - ./config.yaml:/etc/grape/config.yaml:ro
  environment:
    - GRAPE_AUTH_JWT_SECRET=${GRAPE_AUTH_JWT_SECRET}
  command: -c /etc/grape/config.yaml
  restart: unless-stopped

volumes:
  shared-data:
```

---

## Kubernetes 部署

### grape-deployment.yaml

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: grape-registry

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: grape-config
  namespace: grape-registry
data:
  config.yaml: |
    server:
      host: "0.0.0.0"
      port: 4873
    log:
      level: "info"
    auth:
      jwt_secret: "${GRAPE_AUTH_JWT_SECRET}"
      jwt_expiry: 24h
      allow_registration: false
    storage:
      path: "/data"
    database:
      dsn: "/data/grape.db"

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grape
  namespace: grape-registry
  labels:
    app: grape
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grape
  template:
    metadata:
      labels:
        app: grape
    spec:
      containers:
      - name: grape
        image: graperegistry/grape:latest
        ports:
        - containerPort: 4873
          name: http
        volumeMounts:
        - name: config
          mountPath: /etc/grape
          readOnly: true
        - name: data
          mountPath: /data
        env:
        - name: GRAPE_AUTH_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: grape-secret
              key: jwt-secret
        livenessProbe:
          httpGet:
            path: /-/health
            port: 4873
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /-/health
            port: 4873
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: grape-config
      - name: data
        persistentVolumeClaim:
          claimName: grape-pvc

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: grape-pvc
  namespace: grape-registry
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi

---
apiVersion: v1
kind: Secret
metadata:
  name: grape-secret
  namespace: grape-registry
type: Opaque
stringData:
  jwt-secret: "change-this-to-a-secure-random-string"

---
apiVersion: v1
kind: Service
metadata:
  name: grape
  namespace: grape-registry
spec:
  selector:
    app: grape
  ports:
  - port: 80
    targetPort: 4873
    protocol: TCP
  type: ClusterIP

---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: grape
  namespace: grape-registry
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
spec:
  tls:
  - hosts:
    - npm.example.com
    secretName: grape-tls
  rules:
  - host: npm.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: grape
            port:
              number: 80
```

### 部署命令

```bash
# 应用配置
kubectl apply -f grape-deployment.yaml

# 查看状态
kubectl get pods -n grape-registry
kubectl get svc -n grape-registry

# 查看日志
kubectl logs -f deployment/grape -n grape-registry

# 删除
kubectl delete -f grape-deployment.yaml
```

---

## Nginx 反向代理

### 基础配置

```nginx
# /etc/nginx/conf.d/grape.conf
upstream grape_backend {
    server 127.0.0.1:4873;
    keepalive 32;
}

server {
    listen 80;
    server_name npm.example.com;

    # 强制 HTTPS（可选）
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name npm.example.com;

    # SSL 证书
    ssl_certificate /etc/ssl/certs/npm.example.com.crt;
    ssl_certificate_key /etc/ssl/private/npm.example.com.key;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;

    # 现代 SSL 配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;

    # HSTS
    add_header Strict-Transport-Security "max-age=63072000" always;

    location / {
        proxy_pass http://grape_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection "";

        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 120s;
        proxy_read_timeout 120s;

        # 请求体大小限制（npm 包可能很大）
        client_max_body_size 100M;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### 启用配置

```bash
# 测试配置
sudo nginx -t

# 重载 nginx
sudo systemctl reload nginx
```

---

## HTTPS 配置

### 使用 Let's Encrypt

```bash
# 安装 certbot
sudo apt-get install certbot python3-certbot-nginx  # Ubuntu/Debian
sudo yum install certbot python3-certbot-nginx     # CentOS/RHEL

# 获取证书
sudo certbot --nginx -d npm.example.com

# 自动续期（添加到 crontab）
sudo crontab -e
# 添加以下行
0 3 * * * certbot renew --quiet
```

### 使用自签名证书（开发环境）

```bash
# 生成自签名证书
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/grape.key \
  -out /etc/ssl/certs/grape.crt \
  -subj "/C=CN/ST=State/L=City/O=Organization/CN=npm.example.com"

# 设置权限
sudo chmod 600 /etc/ssl/private/grape.key
sudo chmod 644 /etc/ssl/certs/grape.crt
```

---

## 备份与恢复

### 备份数据

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/grape"
DATE=$(date +%Y%m%d_%H%M%S)
DATA_DIR="/var/lib/grape/data"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据
tar -czf $BACKUP_DIR/grape-data-$DATE.tar.gz \
  -C $(dirname $DATA_DIR) $(basename $DATA_DIR)

# 保留最近 7 天的备份
find $BACKUP_DIR -name "grape-data-*.tar.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_DIR/grape-data-$DATE.tar.gz"
```

### 恢复数据

```bash
#!/bin/bash
# restore.sh

BACKUP_FILE=$1
DATA_DIR="/var/lib/grape/data"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

# 停止服务
sudo systemctl stop grape

# 恢复数据
sudo rm -rf $DATA_DIR
sudo tar -xzf $BACKUP_FILE -C $(dirname $DATA_DIR)

# 设置权限
sudo chown -R grape:grape $DATA_DIR

# 启动服务
sudo systemctl start grape

echo "Restore completed from: $BACKUP_FILE"
```

### 定时备份

```bash
# 添加到 crontab
sudo crontab -e
# 每天凌晨 2 点备份
0 2 * * * /opt/scripts/backup.sh
```

---

## 监控与告警

### Prometheus 配置

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'grape'
    static_configs:
      - targets: ['localhost:4873']
    metrics_path: '/-/metrics'
    scrape_interval: 30s
```

### Grafana 仪表盘

导入以下指标面板：

1. **HTTP 请求总数**: `grape_http_requests_total`
2. **请求耗时**: `grape_http_request_duration_seconds`
3. **包下载次数**: `grape_package_downloads_total`
4. **包发布次数**: `grape_package_publish_total`
5. **存储包数量**: `grape_stored_packages_total`
6. **注册用户数**: `grape_registered_users_total`

### 告警规则

```yaml
# alert_rules.yml
groups:
  - name: grape
    rules:
      - alert: GrapeDown
        expr: up{job="grape"} == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Grape 服务宕机"
          description: "Grape 实例 {{ $labels.instance }} 已宕机超过 5 分钟"

      - alert: GrapeHighErrorRate
        expr: |
          sum(rate(grape_http_requests_total{status=~"5.."}[5m])) 
          / sum(rate(grape_http_requests_total[5m])) > 0.05
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Grape 错误率过高"
          description: "Grape 错误率超过 5%"

      - alert: GrapeHighLatency
        expr: |
          histogram_quantile(0.95, 
            sum(rate(grape_http_request_duration_seconds_bucket[5m])) 
            by (le)) > 1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Grape 响应延迟过高"
          description: "Grape P95 延迟超过 1 秒"
```

---

## 故障排查

### 查看日志

```bash
# Systemd 服务
sudo journalctl -u grape -f

# Docker 容器
docker logs -f grape

# 日志文件
tail -f /var/log/grape.log
```

### 常见问题

#### 1. 服务无法启动

```bash
# 检查端口占用
sudo lsof -i :4873
sudo netstat -tlnp | grep 4873

# 检查配置文件
./grape -c config.yaml  # 测试启动

# 检查权限
ls -la /var/lib/grape
```

#### 2. 包发布失败

```bash
# 检查认证
npm whoami --registry http://localhost:4873

# 重新登录
npm logout --registry http://localhost:4873
npm login --registry http://localhost:4873

# 检查磁盘空间
df -h /var/lib/grape
```

#### 3. 上游连接失败

```bash
# 测试上游连接
curl -I https://registry.npmjs.org/lodash

# 检查网络
ping registry.npmjs.org

# 检查代理配置
echo $http_proxy
echo $https_proxy
```

#### 4. 数据库损坏

```bash
# 备份当前数据库
cp /var/lib/grape/data/grape.db /var/lib/grape/data/grape.db.bak

# 删除数据库（重启会自动重建）
rm /var/lib/grape/data/grape.db

# 重启服务
sudo systemctl restart grape
```

### 性能优化

#### 调整文件描述符限制

```bash
# 编辑 /etc/security/limits.conf
grape soft nofile 65536
grape hard nofile 65536
```

#### 调整内核参数

```bash
# 编辑 /etc/sysctl.conf
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535

# 应用
sudo sysctl -p
```

---

## 相关文档

- [配置指南](../configs/README.md) - 配置文件详解
- [API 文档](API.md) - API 参考
- [使用指南](USAGE.md) - 包管理器使用
