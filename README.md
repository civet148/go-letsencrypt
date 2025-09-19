# Go Let's Encrypt 证书申请工具

这是一个用Go语言编写的自动化Let's Encrypt TLS证书申请工具，支持通过HTTP-01质询验证方式为指定域名申请免费的SSL/TLS证书。

## 功能特性

- 🔒 自动申请Let's Encrypt免费TLS证书
- 🌐 支持HTTP-01质询验证
- 📁 自动保存证书和私钥到本地目录
- 🔧 支持生产环境和测试环境
- 📊 显示证书详细信息
- 🔄 自动生成和管理ACME账户私钥

## 前提条件

1. **域名控制权**: 您必须拥有要申请证书的域名，并能够配置DNS A记录指向运行此程序的服务器
2. **公网访问**: 服务器必须能够通过公网访问，Let's Encrypt需要通过HTTP验证域名所有权
3. **端口开放**: 默认使用80端口进行HTTP质询验证（可通过-port参数修改）
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

### 基本用法

```bash
# 为domain.com申请证书
./letsencrypt -domain example.com -email your@email.com

# 使用测试环境（推荐先测试）
./letsencrypt -domain example.com -email your@email.com -staging

# 指定证书保存目录
./letsencrypt -domain example.com -email your@email.com -cert-dir /path/to/certs

# 使用自定义端口进行HTTP验证
./letsencrypt -domain example.com -email your@email.com -port 8080
```

### 命令行参数

- `-domain`: 要申请证书的域名（必需）
- `-email`: 用于Let's Encrypt注册的邮箱地址（必需）
- `-cert-dir`: 证书保存目录（默认: ./certs）
- `-staging`: 使用Let's Encrypt测试环境（默认: false）
- `-port`: HTTP质询验证端口（默认: 80）

## 输出文件

程序成功运行后，会在指定目录生成以下文件：

```
certs/
├── account.key          # ACME账户私钥（自动生成）
├── example.com.crt      # 域名证书文件
└── example.com.key      # 域名私钥文件
```

## 域名验证流程

1. **DNS配置**: 确保域名的A记录指向运行程序的服务器IP
2. **端口开放**: 确保防火墙允许访问验证端口（默认80）
3. **HTTP验证**: Let's Encrypt会访问 `http://域名/.well-known/acme-challenge/token` 进行验证
4. **证书颁发**: 验证通过后自动下载证书

## 注意事项

### 重要提醒

- **首次使用建议先用测试环境** (`-staging` 参数)，避免触发生产环境的频率限制
- **生产环境限制**: Let's Encrypt对每个域名每周最多允许申请5张证书
- **续期管理**: 证书有效期为90天，需要定期续期
- **安全考虑**: 私钥文件权限为600，请妥善保管

### 常见问题

1. **端口占用**: 如果80端口被占用，可以使用`-port`参数指定其他端口，但需要配置反向代理
2. **防火墙**: 确保服务器防火墙允许外部访问验证端口
3. **DNS延迟**: 如果刚配置DNS，可能需要等待DNS生效
4. **域名格式**: 目前仅支持单个域名，不支持通配符证书

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

您可以通过cron job实现自动续期：

```bash
# 每月1号凌晨2点检查并续期证书
0 2 1 * * /path/to/letsencrypt -domain example.com -email your@email.com
```

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 贡献

欢迎提交Issue和Pull Request！