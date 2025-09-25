#!/bin/bash

# Go 环境设置脚本
# 用于确保 IDE 和终端都能正确识别 Go 环境

echo "=== Go 环境检查 ==="
echo "Go 版本: $(go version)"
echo "Go 路径: $(which go)"
echo "GOROOT: $(go env GOROOT)"
echo "GOPATH: $(go env GOPATH)"
echo ""

echo "=== 设置环境变量 ==="
export GOROOT=$(go env GOROOT)
export GOPATH=$(go env GOPATH)
export PATH=$GOROOT/bin:$GOPATH/bin:$PATH

echo "GOROOT 已设置为: $GOROOT"
echo "GOPATH 已设置为: $GOPATH"
echo ""

echo "=== 验证编译 ==="
if go build -o udptunnel . ; then
    echo "✅ 编译成功！"
    echo "可执行文件: $(ls -la udptunnel)"
else
    echo "❌ 编译失败"
fi

echo ""
echo "=== 使用说明 ==="
echo "如果 IDE 仍有问题，请："
echo "1. 重启 IDE"
echo "2. 在 IDE 中运行 'Go: Install/Update Tools'"
echo "3. 检查 IDE 的 Go 扩展设置"


