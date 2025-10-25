# Go Let's Encrypt 证书申请工具

这是一个用Go语言编写的自动化Let's Encrypt TLS证书申请工具，支持通过HTTP-01和DNS-01质询验证方式为指定域名申请免费的SSL/TLS证书。

## 功能特性

- 🔒 自动申请Let's Encrypt免费TLS证书
- 🌐 支持HTTP-01质询验证
- 📶 支持DNS-01质询验证（推荐）
- ☁️ 集成阿里云DNS自动化管理
- 📁 自动保存证书和私钥到本地目录
- 🔧 支持生产环境和测试环境
- 📊 显示证书详细信息
- 🔄 自动生成和管理ACME账户私钥

## 前提条件

### HTTP-01验证模式
1. **域名控制权**: 您必须拥有要申请证书的域名，并能够配置DNS A记录指向运行此程序的服务器
2. **公网访问**: 服务器必须能够通过公网访问，Let's Encrypt需要通过HTTP验证域名所有权
3. **端口开放**: 默认使甩80端口进行HTTP质询验证（可通过-port参数修改）

### DNS-01验证模式（推荐）
1. **DNS管理权限**: 您必须拥有域名的DNS管理权限
2. **阿里云凭证**: 如需自动化，需要阿里云AccessKey ID和AccessKey Secret
3. **无需HTTP访问**: 不需要公网IP或端口访问

### 通用要求
4. **Go环境**: Go 1.21或更高版本

## 安装和编译

```bash
# 克隆项目
git clone https://github.com/civet148/go-letsencrypt
cd go-letsencrypt

# 下载依赖
go mod tidy

# 编译程序
go build -o letsencrypt main.go
```

## 使用方法

### HTTP-01验证（适用于有公网IP的服务器）

```bash
# 为domain.com申请证书
./go-letsencrypt -domain example.com -email your@email.com

# 使用测试环境（推荐先测试）
./go-letsencrypt -domain example.com -email your@email.com -staging

# 指定证书保存目录
./go-letsencrypt -domain example.com -email your@email.com -cert-dir /path/to/certs

# 使用自定义端口进行HTTP验证
./go-letsencrypt -domain example.com -email your@email.com -port 8080
```

### DNS-01验证（推荐，适用于所有情况）

#### 1. 阿里云自动化模式（推荐）

```bash
# 测试环境 - 完全自动化
./go-letsencrypt -domain example.com -email your@email.com \
    --staging --dns \
    --aliyun-key YOUR_ACCESS_KEY_ID \
    --aliyun-secret YOUR_ACCESS_KEY_SECRET

# 生产环境 - 完全自动化
./go-letsencrypt -domain example.com -email your@email.com \
    --dns \
    --aliyun-key YOUR_ACCESS_KEY_ID \
    --aliyun-secret YOUR_ACCESS_KEY_SECRET

# 使用环境变量（更安全）
export ALIYUN_ACCESS_KEY_ID="YOUR_ACCESS_KEY_ID"
export ALIYUN_ACCESS_KEY_SECRET="YOUR_ACCESS_KEY_SECRET"
./go-letsencrypt -domain example.com -email your@email.com \
    --dns \
    --aliyun-key "$ALIYUN_ACCESS_KEY_ID" \
    --aliyun-secret "$ALIYUN_ACCESS_KEY_SECRET"
```

#### 2. 手动DNS模式

```bash
# 需要手动添加DNS TXT记录
./go-letsencrypt -domain example.com -email your@email.com --dns --manual
```

### 命令行参数详解

| 参数 | 简写 | 说明 | 默认值 | 必需 |
|------|------|------|---------|------|
| `--domain` | `-d` | 要申请证书的域名 | - | ✓ |
| `--email` | `-m` | 用于Let's Encrypt注册的邮箱地址 | - | ✓ |
| `--cert-dir` | `-c` | 证书保存目录 | `./certs` | - |
| `--staging` | `-s` | 使用Let's Encrypt测试环境 | `false` | - |
| `--port` | `-p` | HTTP质询验证端口 | `80` | - |
| `--dns-only` | `--dns` | 使用DNS-01验证方式 | `false` | - |
| `--manual` | `-M` | 手动添加DNS TXT记录 | `false` | - |
| `--aliyun-key` | `--ak` | 阿里云AccessKey ID | - | - |
| `--aliyun-secret` | `--as` | 阿里云AccessKey Secret | - | - |

## 阿里云DNS自动化配置

### 1. 获取阿里云AccessKey

1. 登录[阿里云控制台](https://ram.console.aliyun.com/manage/ak)
2. 创建AccessKey或使用现有的
3. 记录AccessKey ID和AccessKey Secret

### 2. 设置DNS权限

确保您的AccessKey具有以下权限：
- `AliyunDNSFullAccess` 或更精细的DNS管理权限
- 能够添加和删除TXT记录

### 3. 使用示例

```bash
# 实际域名示例（ruixiangyun.com）
./go-letsencrypt -domain ruixiangyun.com -email admin@example.com \
    --staging --dns \
    --aliyun-key ${YOUR_ACCESS_KEY_ID} \
    --aliyun-secret ${YOUR_ACCESS_KEY_SECRET}

# 生产环境
./go-letsencrypt -domain ruixiangyun.com -email admin@example.com \
    --dns \
    --aliyun-key ${YOUR_ACCESS_KEY_ID} \
    --aliyun-secret ${YOUR_ACCESS_KEY_SECRET}
```

### 4. 工作流程

1. **自动添加DNS记录**: 程序自动调用阿里云API添加`_acme-challenge.yourdomain.com` TXT记录
2. **等待DNS传播**: 程序等待30秒让DNS记录传播
3. **Let's Encrypt验证**: Let's Encrypt服务器通过DNS查询验证域名所有权
4. **自动清理**: 验证成功后自动删除临时DNS记录
5. **下载证书**: 保存证书和私钥到本地

### 5. 优势

- ✅ **无需公网IP**: 可在任意机器上运行
- ✅ **无需端口访问**: 不需要开放80端口
- ✅ **完全自动化**: 无需人工干预
- ✅ **支持内网域名**: 适用于内网或私有域名
- ✅ **自动清理**: 验证后自动删除临时记录

## 输出文件

程序成功运行后，会在指定目录生成以下文件：

```
certs/
├── account.key          # ACME账户私钥（自动生成）
├── example.com.crt      # 域名证书文件
└── example.com.key      # 域名私钥文件
```

## 验证方式对比

| 特性 | HTTP-01验证 | DNS-01验证 |
|------|-------------|-------------|
| **公网IP要求** | ✓ 必需 | ✗ 不需 |
| **端口访问** | ✓ 需老80端口 | ✗ 不需 |
| **DNS管理权限** | ✗ 不需 | ✓ 必需 |
| **防火墙要求** | ✓ 需开放端口 | ✗ 无要求 |
| **内网支持** | ✗ 不支持 | ✓ 支持 |
| **通配符证书** | ✗ 不支持 | ✓ 支持 |
| **自动化程度** | ✓ 中等 | ✓ 高（阿里云） |

## 域名验证流程

### HTTP-01验证流程
1. **DNS配置**: 确保域名的A记录指向运行程序的服务器IP
2. **端口开放**: 确保防火墙允许访问验证端口（默认80）
3. **HTTP验证**: Let's Encrypt会访问 `http://域名/.well-known/acme-challenge/token` 进行验证
4. **证书颁发**: 验证通过后自动下载证书

### DNS-01验证流程
1. **DNS权限验证**: 确保拥有域名的DNS管理权限
2. **添加TXT记录**: 在`_acme-challenge.域名`添加指定TXT记录
3. **DNS验证**: Let's Encrypt通过DNS查询验证域名所有权
4. **证书颁发**: 验证通过后自动下载证书
5. **清理记录**: 自动删除临时TXT记录（阿里云模式）

## 注意事项

### 重要提醒

- **首次使用建议先用测试环境** (`--staging` 参数)，避免触发生产环境的频率限制
- **生产环境限制**: Let's Encrypt对每个域名每周最多允许申请20张证书
- **续期管理**: 证书有效期为90天，需要定期续期
- **安全考虑**: 私钥文件权限为600，请妥善保管
- **DNS验证推荐**: 对于内网或无公网IP的场景，强烈推荐使用DNS-01验证

### 常见问题

**HTTP-01验证相关**:
1. **端口占用**: 如果80端口被占用，可以使用`--port`参数指定其他端口，但需要配置反向代理
2. **防火墙**: 确保服务器防火墙允许外部访问验证端口
3. **DNS延迟**: 如果刚配置DNS，可能需要等待DNS生效

**DNS-01验证相关**:
4. **阿里云权限**: 确保AccessKey具有DNS管理权限
5. **域名托管**: 确保域名在阿里云DNS管理
6. **网络连接**: 确保可以访问阿里云API

**通用问题**:
7. **域名格式**: 目前支持单个域名和通配符域名（DNS-01）
8. **证书续期**: 建议在证书过期前30天进行续期

## 证书使用示例

### Nginx配置

```nginx
server {
    listen 443 ssl;
    server_name example.com;
    
    ssl_certificate /path/to/certs/example.com.crt;
    ssl_certificate_key /path/to/certs/example.com.key;
    
    # 其他SSL配置...
}
```

### Apache配置

```apache
<VirtualHost *:443>
    ServerName example.com
    
    SSLEngine on
    SSLCertificateFile /path/to/certs/example.com.crt
    SSLCertificateKeyFile /path/to/certs/example.com.key
    
    # 其他SSL配置...
</VirtualHost>
```

## 自动化续期

### 传统方式（HTTP-01）

```bash
# 每月1号凌晨2点检查并续期证书
0 2 1 * * /path/to/go-letsencrypt -domain example.com -email your@email.com
```

### 阿里云自动化（DNS-01）

```bash
# 使用环境变量保存凭证
# 在 ~/.bashrc 或 /etc/environment 中添加:
export ALIYUN_ACCESS_KEY_ID="YOUR_ACCESS_KEY_ID"
export ALIYUN_ACCESS_KEY_SECRET="YOUR_ACCESS_KEY_SECRET"

# crontab 配置
0 2 1 * * /path/to/go-letsencrypt -domain example.com -email your@email.com --dns --aliyun-key "$ALIYUN_ACCESS_KEY_ID" --aliyun-secret "$ALIYUN_ACCESS_KEY_SECRET"

# 多域名续期脚本
cat > /usr/local/bin/renew-certs.sh << 'EOF'
#!/bin/bash
domains=("example.com" "api.example.com" "www.example.com")
for domain in "${domains[@]}"; do
    /path/to/go-letsencrypt -domain "$domain" -email your@email.com \
        --dns \
        --aliyun-key "$ALIYUN_ACCESS_KEY_ID" \
        --aliyun-secret "$ALIYUN_ACCESS_KEY_SECRET"
done
EOF
chmod +x /usr/local/bin/renew-certs.sh

# 每月1号凌晨2点执行续期
0 2 1 * * /usr/local/bin/renew-certs.sh
```

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 贡献

欢迎提交Issue和Pull Request！