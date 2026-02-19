# AgentFramework Makefile
# 多渠道机器人框架 - 构建和开发工具

.PHONY: all build run test clean fmt lint deps help docker

# Variables
BINARY_NAME=agentframework
CMD_DIR=./cmd
BUILD_DIR=./build
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Default target
all: deps build

## help: 显示帮助信息
help:
	@echo "AgentFramework 多渠道机器人框架"
	@echo ""
	@echo "使用方法: make [target]"
	@echo ""
	@echo "常用命令:"
	@echo "  make deps          - 下载依赖"
	@echo "  make build         - 构建程序"
	@echo "  make run           - 运行简单机器人"
	@echo "  make test          - 运行测试"
	@echo "  make clean         - 清理构建文件"
	@echo "  make fmt           - 格式化代码"
	@echo "  make dev-setup     - 开发环境设置"

## deps: 下载依赖
deps:
	@echo "📦 下载依赖..."
	$(GOMOD) tidy
	$(GOMOD) download

## build: 构建程序
build:
	@echo "🔨 构建程序..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/*/

## build-simplebot: 构建简单机器人
build-simplebot:
	@echo "🤖 构建简单机器人..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/simplebot $(CMD_DIR)/simplebot/main.go

## run: 运行简单机器人
run: build-simplebot
	@echo "🚀 运行机器人..."
	./$(BUILD_DIR)/simplebot

## test: 运行所有测试
test:
	@echo "🧪 运行测试..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

## test-channels: 仅测试渠道模块
test-channels:
	@echo "🧪 测试多渠道模块..."
	$(GOTEST) -v -race ./pkg/channels/...

## fmt: 格式化代码
fmt:
	@echo "🎨 格式化代码..."
	$(GOFMT) -s -w .

## vet: 代码静态分析
vet:
	@echo "🔬 Go vet 静态分析..."
	$(GOCMD) vet ./...

## clean: 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -f config/channels.generated.yaml

## gen-keys: 生成示例环境变量文件
gen-keys:
	@echo "🔑 生成测试配置..."
	@cat > .env.example << 'EOF'
# Telegram
TELEGRAM_BOT_TOKEN=your_telegram_bot_token

# Discord
DISCORD_BOT_TOKEN=your_discord_bot_token

# Slack
SLACK_BOT_TOKEN=xoxb-your-slack-bot-token
SLACK_APP_TOKEN=xapp-your-slack-app-token

# Feishu
FEISHU_APP_ID=cli_your_feishu_app_id
FEISHU_APP_SECRET=your_feishu_app_secret

# WeWork
WEWORK_CORP_ID=your_wework_corp_id
WEWORK_CORP_SECRET=your_wework_corp_secret
WEWORK_AGENT_ID=your_wework_agent_id

# DingTalk
DINGTALK_APP_KEY=your_dingtalk_app_key
DINGTALK_APP_SECRET=your_dingtalk_app_secret

# QQ
QQ_BOT_ENABLED=true
QQ_BOT_API_BASE=http://127.0.0.1:3000
EOF
	@echo "✅ .env.example 文件已创建"

## dev-setup: 开发环境设置
dev-setup: deps gen-keys
	@echo "✅ 开发环境设置完成"
	@echo ""
	@echo "下一步:"
	@echo "1. 复制 .env.example 到 .env 并填入你的 API 凭证"
	@echo "2. 运行 'make run' 启动机器人"
	@echo "3. 运行 'make test' 运行测试"

## check: 完整检查
check: fmt vet test
	@echo "✅ 所有检查通过"

## version: 显示版本信息
version:
	@echo "Version: $(VERSION)"
