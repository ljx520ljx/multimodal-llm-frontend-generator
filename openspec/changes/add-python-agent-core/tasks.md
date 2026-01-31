# Tasks: Add Python Agent Core Service

## 1. Project Setup

- [x] 1.1 创建 `agent-core/` 项目目录结构
- [x] 1.2 创建 `pyproject.toml` 配置文件（使用 Poetry 或 pip）
- [x] 1.3 创建 `requirements.txt` 依赖列表

## 2. FastAPI Application

- [x] 2.1 实现 FastAPI 入口 (`app/main.py`)
- [x] 2.2 实现 `/health` 健康检查接口
- [x] 2.3 配置 CORS 中间件

## 3. Configuration

- [x] 3.1 实现配置管理 (`app/config.py`)
- [x] 3.2 支持环境变量读取
- [x] 3.3 配置 Pydantic Settings

## 4. Logging

- [x] 4.1 配置结构化日志
- [x] 4.2 支持日志级别配置

## 5. Docker Support

- [x] 5.1 编写 Dockerfile
- [x] 5.2 创建 `.dockerignore`

## 6. Documentation

- [x] 6.1 创建 README.md
- [x] 6.2 创建 CLAUDE.md (L2 模块文档)

## 7. Verification

- [x] 7.1 本地运行测试
- [x] 7.2 验证 `/health` 接口返回 `{"status": "ok"}`
