# 项目主旨

这是一个 **LLM 自维护的 wiki 知识库**：知识不是每次查询时临时用 RAG 检索，而是由 LLM 增量沉淀为持久、互链、复利增长的 Markdown 页面；人通过编辑器浏览和修订，AI 通过 MCP 工具维护。

设计蓝本是 Karpathy 的 "LLM Wiki"（https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f）。核心概念：

- **三层结构**：原始资料（raw sources，不可变的事实来源）→ wiki 页面（LLM 提炼维护的 Markdown 层）→ schema（约定页面如何组织的规则）。
- **三类操作**：摄入（Ingest，从原始资料提炼/更新页面）、查询（Query，基于 wiki 回答并回写新知识）、体检（Lint，发现矛盾、过时、孤页并修复）。

## 当前架构

本仓库（llm-wiki-server）是该知识库的后端，Go 模块名 `llm-wiki-server`：

- 存储为 PostgreSQL，schema 由 golang-migrate 管理（`migrations/`）。核心表 `wiki_pages`：slug 唯一、title、content（Markdown）、tags、alias（模糊匹配用）。原始资料层由 `model.Source`（`sources` 表，origin 唯一、title、summary）与 `wiki_page_sources` 关联表组成；来源全库共享，关联随页面 Create/Update/Write 在同一事务内维护。
- 两个入口共享同一套 service/repository：
  - `cmd/server`：chi HTTP API（`/api/wiki`），服务于 llm-wiki-editor（React 前端，独立仓库）。
  - `cmd/mcp`：stdio MCP server，AI agent 经由它维护 wiki。
- 分层：route / tool（协议边界，输入结构体定义于此）→ service → repository → model。
- 常用命令：`make build-api`、`make build-mcp`、`make migrate-up`；配置见 `config.example.yml`。

## 面向 AI 的接口设计

工具是给 LLM 用的，语义要防呆：

- 改内容优先暴露**字符串替换**语义（精确匹配 + 唯一性校验 + `replace_all`），整页替换只留给真正的重写，避免模型重写长页时静默丢内容。
- 工具的 description 和错误文案是 AI 的行为提示，按功能对待：说清失败原因（未找到 / 不唯一 / slug 冲突），并指明恢复方式（如"重新读取后再试"）。

# 项目约定

## 代码风格

- **不要防御性编程。** 不要为一次性错误预定义 `ErrXxx` 哨兵变量，也不要把简单参数封装成请求/响应结构体。错误在发生处直接用 `errors.New` / `fmt.Errorf` 返回；同一段错误信息或逻辑被复用 **3 次以上** 时才考虑提取常量或类型。
- **参数保持扁平。** service 方法直接接收简单参数（如 `slug, oldString, newString string, replaceAll bool`）；只有跨越边界、需要描述 schema 时（如 MCP tool 的输入/输出）才定义结构体，并且定义在边界所在的包（tool 层）。
- **测试是临时的。** 默认不新增测试文件。需要验证行为时可以临时编写测试，但**任务完成后必须删除**，除非用户明确要求保留。
