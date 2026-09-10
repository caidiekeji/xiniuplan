# 贡献指南

感谢您对云枢面板的关注！本文档描述了如何参与项目开发，包括开发环境搭建、代码规范、提交流程和代码审查标准。

## 项目概览

云枢面板（CloudPivot）是一款轻量级 Linux 服务器可视化管理平台，使用 Go + Vue3 技术栈构建。我们欢迎各种形式的贡献，包括代码提交、Bug 报告、功能建议、文档改进和翻译。

## 贡献方式

### 报告 Bug

如果您发现了 Bug，请通过以下方式报告：

1. 搜索现有 Issue，确认是否已有人报告
2. 创建新 Issue，包含以下信息：
   - **标题**：简洁描述问题
   - **环境**：操作系统、面板版本、Go/Node 版本
   - **复现步骤**：详细的操作步骤
   - **预期行为**：您期望的正确行为
   - **实际行为**：实际发生的错误
   - **日志/截图**：相关的错误日志或截图
   - **复现概率**：每次都出现 / 偶发

### 提交功能建议

如果您有新功能或改进建议：

1. 搜索现有 Issue 和 PR，确认是否已有类似建议
2. 创建 Issue，描述：
   - **功能描述**：您希望添加什么功能
   - **使用场景**：这个功能解决什么问题
   - **实现思路**：您的初步实现想法（可选）
   - **替代方案**：是否考虑过其他实现方式

### 贡献代码

如果您希望直接贡献代码，请参考下方的开发流程。

### 改进文档

文档改进同样重要！如果您发现文档有误、不完整或可以优化，欢迎提交 PR。

## 开发环境搭建

### 前置要求

| 工具 | 版本要求 | 说明 |
|------|----------|------|
| Go | 1.22+ | 后端开发语言 |
| Node.js | 18+ | 前端构建，推荐 LTS 版本 |
| npm | 9+ | 随 Node.js 安装 |
| Docker | 20.10+ | 容器功能测试（可选） |
| Git | 2.0+ | 版本控制 |
| Make | 3.8+ | 构建脚本（可选） |

### 克隆项目

```bash
git clone <repository-url> cloudpivot
cd cloudpivot
```

### 后端开发环境

```bash
cd backend

# 下载依赖
go mod download

# 运行后端（开发模式）
go run cmd/agent/main.go

# 后端默认监听 8888 端口
# 健康检查: curl http://localhost:8888/health
```

### 前端开发环境

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 前端默认监听 5173 端口
# 访问: http://localhost:5173
# /api 请求自动代理到 http://localhost:8888
```

### 运行测试

```bash
# 后端测试
cd backend
go test ./... -v

# 前端测试
cd frontend
npm run test

# 前端 Lint
npm run lint

# 前端类型检查
npm run type-check
```

## 代码规范

### Go 代码规范

1. **格式化**：所有 Go 代码必须通过 `gofmt` 格式化
   ```bash
   gofmt -w .
   ```

2. **Lint**：使用 `golangci-lint` 进行代码检查
   ```bash
   golangci-lint run
   ```

3. **命名规范**：
   - 包名：小写、简短、有意义（如 `auth`、`site`、`database`）
   - 导出标识符：PascalCase（如 `UserService`、`CreateSiteRequest`）
   - 未导出标识符：camelCase（如 `userService`、`parseSiteID`）
   - 常量：PascalCase 或全大写下划线（如 `CodeSuccess`、`MAX_RETRY`）
   - 缩写词：保持大写（如 `URL`、`ID`、`HTTP`，不用 `Url`、`Id`）

4. **错误处理**：
   - 不忽略错误，除非明确注释原因
   - 错误信息使用小写开头，不包含标点
   - 使用 `fmt.Errorf("上下文: %w", err)` 包装错误
   - 优先使用自定义错误类型

5. **包结构**：
   - 每个包聚焦单一职责
   - API 层只做参数处理和响应组装
   - Service 层包含业务逻辑
   - Model 层只做数据操作
   - DTO 层定义请求/响应结构

6. **注释规范**：
   - 导出的类型、函数、方法必须有注释
   - 注释以标识符名称开头
   - 复杂逻辑添加行内注释说明原因
   - TODO/FIXME 注释包含负责人和日期

### 前端代码规范

1. **格式化**：使用 Prettier 格式化
   ```bash
   npm run format
   ```

2. **Lint**：使用 ESLint 检查
   ```bash
   npm run lint
   ```

3. **组件规范**：
   - 使用 `<script setup lang="ts">` 语法
   - 组件名使用 PascalCase（如 `SiteList.vue`、`UserForm.vue`）
   - Props 使用 TypeScript 类型定义
   - Emits 显式声明
   - 组合式 API，优先使用 Composition API

4. **状态管理**：
   - 使用 Pinia 管理全局状态
   - Store 命名：`useXxxStore`
   - 状态变更通过 Actions，不直接修改 State
   - 复杂状态使用 Getters 派生

5. **API 封装**：
   - 所有 API 请求封装在 `src/api/` 目录
   - 按模块划分文件（如 `website.ts`、`database.ts`）
   - 使用统一的 HTTP 客户端（axios 封装）
   - 请求/响应类型使用 TypeScript 接口定义

6. **样式规范**：
   - 使用 CSS 变量定义主题色
   - 组件样式使用 scoped
   - 优先使用 Naive UI 组件，不重复造轮子
   - 响应式设计，适配不同屏幕尺寸

### Git 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <subject>

<body>

<footer>
```

#### Type 类型

| 类型 | 说明 |
|------|------|
| feat | 新功能 |
| fix | Bug 修复 |
| docs | 文档变更 |
| style | 代码格式（不影响功能） |
| refactor | 重构（既不是新功能也不是修复） |
| perf | 性能优化 |
| test | 测试相关 |
| build | 构建系统或外部依赖变更 |
| ci | CI 配置变更 |
| chore | 其他杂项 |

#### Scope（可选）

影响的模块，如 `auth`、`site`、`database`、`frontend`、`docs`。

#### Subject

简短描述，不超过 50 字符，使用祈使句，小写开头，不加句号。

#### 示例

```
feat(site): 添加网站 SSL 证书自动续期功能

- 实现 Let's Encrypt 证书自动续期
- 添加续期任务调度
- 续期失败时发送告警通知

Closes #123
```

```
fix(auth): 修复 JWT 令牌过期后无法刷新的问题

刷新令牌时未正确验证令牌签名，导致合法令牌被拒绝。
增加令牌黑名单检查，确保登出后的令牌无法刷新。

Fixes #456
```

## 分支策略

### 分支模型

```
main (生产分支，保护分支)
  │
  └── develop (开发分支，保护分支)
        │
        ├── feature/xxx (功能分支)
        ├── fix/xxx (修复分支)
        ├── docs/xxx (文档分支)
        └── release/x.x.x (发布分支)
```

### 分支命名

| 分支类型 | 命名格式 | 示例 |
|----------|----------|------|
| 功能分支 | `feature/<功能描述>` | `feature/ssl-auto-renew` |
| 修复分支 | `fix/<问题描述>` | `fix/jwt-refresh-error` |
| 文档分支 | `docs/<文档描述>` | `docs/api-endpoints` |
| 发布分支 | `release/<版本号>` | `release/1.1.0` |
| 热修复分支 | `hotfix/<问题描述>` | `hotfix/login-crash` |

### 工作流程

1. **从 develop 分支创建功能分支**
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/my-feature
   ```

2. **开发并提交**
   ```bash
   git add .
   git commit -m "feat(module): 描述"
   ```

3. **推送分支**
   ```bash
   git push origin feature/my-feature
   ```

4. **创建 Pull Request**
   - 目标分支：`develop`
   - 填写 PR 标题和描述
   - 关联相关 Issue
   - 勾选自检清单

5. **代码审查**
   - 至少 1 位审查者批准
   - 修复审查中发现的问题
   - CI 检查全部通过

6. **合并**
   - 使用 Squash Merge 合并
   - 删除功能分支

## Pull Request 规范

### PR 标题

使用与提交信息相同的格式：
```
feat(site): 添加 SSL 自动续期功能
```

### PR 描述模板

```markdown
## 变更说明

简要描述本次 PR 的变更内容。

## 变更类型

- [ ] 新功能 (feat)
- [ ] Bug 修复 (fix)
- [ ] 文档更新 (docs)
- [ ] 代码重构 (refactor)
- [ ] 性能优化 (perf)
- [ ] 其他 (chore)

## 关联 Issue

Closes #123

## 测试说明

- [ ] 单元测试已通过
- [ ] 集成测试已通过
- [ ] 手动测试已完成

## 截图/录屏

（UI 变更请附上截图）

## 自检清单

- [ ] 代码已通过 gofmt / prettier 格式化
- [ ] 代码已通过 lint 检查
- [ ] 已添加/更新相关测试
- [ ] 已更新相关文档
- [ ] 无硬编码的敏感信息
- [ ] 已考虑安全性和性能影响
```

## 代码审查标准

审查者应关注以下方面：

### 功能正确性

- [ ] 代码实现了 PR 描述的功能
- [ ] 边界条件已处理（空值、极值、并发）
- [ ] 错误处理完善，无静默失败
- [ ] 没有引入新的 Bug

### 代码质量

- [ ] 代码简洁、可读、可维护
- [ ] 遵循项目代码规范
- [ ] 无重复代码，合理抽象
- [ ] 命名清晰、有意义
- [ ] 注释适当，解释"为什么"而非"做什么"

### 安全性

- [ ] 无 SQL 注入风险（使用参数化查询）
- [ ] 无命令注入风险（使用白名单或参数化）
- [ ] 无 XSS 风险（输出转义）
- [ ] 无路径遍历风险（路径校验）
- [ ] 敏感信息不硬编码、不记录日志
- [ ] 权限校验完整

### 性能

- [ ] 无明显的性能瓶颈
- [ ] 数据库查询合理（无 N+1 查询）
- [ ] 大文件/大数据处理考虑内存使用
- [ ] 并发安全（竞态条件、锁的使用）

### 测试

- [ ] 新功能有对应的单元测试
- [ ] Bug 修复有回归测试
- [ ] 测试覆盖关键路径和边界条件
- [ ] 测试可重复运行，无外部依赖

### 文档

- [ ] 新功能已更新用户文档
- [ ] API 变更已更新 API 文档
- [ ] 配置项变更已更新配置说明
- [ ] 代码注释完整

## 版本发布

### 版本号规范

使用 [语义化版本](https://semver.org/lang/zh-CN/)：`主版本号.次版本号.修订号`

| 版本类型 | 说明 | 示例 |
|----------|------|------|
| 主版本号 | 不兼容的 API 变更 | 1.0.0 → 2.0.0 |
| 次版本号 | 向下兼容的功能新增 | 1.0.0 → 1.1.0 |
| 修订号 | 向下兼容的问题修复 | 1.0.0 → 1.0.1 |

### 发布流程

1. 从 `develop` 创建 `release/x.x.x` 分支
2. 更新版本号（后端 `config.go`、前端 `package.json`）
3. 更新 CHANGELOG.md
4. 执行完整测试
5. 合并到 `main` 分支
6. 打 Tag：`git tag vx.x.x`
7. 推送 Tag：`git push origin vx.x.x`
8. GitHub Release 发布（包含变更说明和构建产物）
9. 合并回 `develop` 分支

## 行为准则

参与本项目即表示您同意以下行为准则：

1. **尊重他人**：使用包容和尊重的语言，尊重不同的观点和经验
2. **专业沟通**：对事不对人，建设性地表达不同意见
3. **承担责任**：承认错误并道歉，对自己的代码负责
4. **关注社区**：优先考虑社区和用户的利益
5. **保持友善**：欢迎新贡献者，提供帮助和指导

## 获取帮助

如果您在贡献过程中遇到问题：

1. 搜索现有文档和 Issue
2. 在 Discussion 区提问
3. 在相关 Issue 下留言
4. 联系维护者

感谢您的贡献！
