#!/bin/bash

# UDP 隧道程序多平台构建脚本

echo "🚀 开始构建 UDP 隧道程序的多平台版本..."
echo ""

# 清理旧的构建文件
echo "🧹 清理旧的构建文件..."
rm -f udptunnel-*

# 创建构建目录
mkdir -p build

# 构建配置
declare -a platforms=(
    "darwin/amd64"      # Intel Mac
    "darwin/arm64"      # Apple M1/M2 Mac  
    "linux/amd64"       # Linux x86_64
    "linux/arm64"       # Linux ARM64
    "windows/amd64"     # Windows x86_64
    "windows/arm64"     # Windows ARM64
)

echo "📦 开始交叉编译..."
echo ""

for platform in "${platforms[@]}"; do
    IFS='/' read -r os arch <<< "$platform"
    
    # 设置输出文件名
    output="udptunnel-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        output="${output}.exe"
    fi
    
    echo "  🔨 构建 ${os}/${arch}..."
    
    # 交叉编译
    if GOOS="$os" GOARCH="$arch" go build -ldflags="-s -w" -o "build/$output" .; then
        # 获取文件大小
        size=$(du -h "build/$output" | cut -f1)
        echo "    ✅ 成功: build/$output ($size)"
    else
        echo "    ❌ 失败: ${os}/${arch}"
    fi
done

echo ""
echo "📊 构建结果:"
echo ""
ls -la build/

echo ""
echo "🎯 Apple M1/M2 版本:"
if [ -f "build/udptunnel-darwin-arm64" ]; then
    file build/udptunnel-darwin-arm64
    echo "文件大小: $(du -h build/udptunnel-darwin-arm64 | cut -f1)"
    echo ""
    echo "📋 使用方法:"
    echo "  1. 将 build/udptunnel-darwin-arm64 复制到 Mac"
    echo "  2. 添加执行权限: chmod +x udptunnel-darwin-arm64"
    echo "  3. 运行: ./udptunnel-darwin-arm64 -help"
else
    echo "❌ Apple M1 版本构建失败"
fi

echo ""
echo "✨ 构建完成！所有平台的二进制文件都在 build/ 目录中"

