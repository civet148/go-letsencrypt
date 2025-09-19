#!/bin/bash

# Let's Encrypt 证书申请示例脚本
# 使用前请修改以下变量

DOMAIN="example.com"          # 替换为您的域名
EMAIL="your@email.com"        # 替换为您的邮箱
CERT_DIR="./certs"           # 证书保存目录
USE_STAGING=true             # 是否使用测试环境（首次使用建议设为true）

echo "=== Let's Encrypt 证书申请工具 ==="
echo "域名: $DOMAIN"
echo "邮箱: $EMAIL"
echo "证书目录: $CERT_DIR"
echo "测试环境: $USE_STAGING"
echo

# 检查域名和邮箱是否已设置
if [ "$DOMAIN" = "example.com" ] || [ "$EMAIL" = "your@email.com" ]; then
    echo "❌ 错误: 请先修改脚本中的DOMAIN和EMAIL变量"
    echo "   DOMAIN=$DOMAIN"
    echo "   EMAIL=$EMAIL"
    exit 1
fi

# 构建命令参数
ARGS="-domain $DOMAIN -email $EMAIL -cert-dir $CERT_DIR"
if [ "$USE_STAGING" = true ]; then
    ARGS="$ARGS -staging"
    echo "🧪 使用测试环境（推荐首次使用）"
else
    echo "🔥 使用生产环境"
fi

echo
echo "执行命令: ./go-letsencrypt $ARGS"
echo

# 检查程序是否存在
if [ ! -f "./go-letsencrypt" ]; then
    echo "❌ 未找到 go-letsencrypt 程序，请先编译："
    echo "   go build -o go-letsencrypt main.go"
    exit 1
fi

# 确认执行
read -p "是否继续？(y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "已取消"
    exit 0
fi

# 执行证书申请
echo "开始申请证书..."
./go-letsencrypt $ARGS

# 检查结果
if [ $? -eq 0 ]; then
    echo
    echo "✅ 证书申请成功！"
    echo "证书文件: $CERT_DIR/$DOMAIN.crt"
    echo "私钥文件: $CERT_DIR/$DOMAIN.key"
    echo
    echo "下一步:"
    if [ "$USE_STAGING" = true ]; then
        echo "1. 测试证书申请成功，现在可以申请正式证书"
        echo "2. 将脚本中的 USE_STAGING 设为 false"
        echo "3. 重新运行脚本申请正式证书"
    else
        echo "1. 配置您的Web服务器使用新证书"
        echo "2. 设置自动续期任务"
    fi
else
    echo
    echo "❌ 证书申请失败，请检查错误信息"
    echo
    echo "常见问题:"
    echo "1. 确保域名DNS A记录指向此服务器"
    echo "2. 确保防火墙允许80端口访问"
    echo "3. 检查域名是否可以从外网访问"
fi