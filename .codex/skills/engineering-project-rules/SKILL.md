---
name: engineering-project-rules
description: 在已有代码库中进行功能开发、Bug 修复、问题排查、SQL/缓存/性能相关改动、前端页面改动、代码审查或项目规则沉淀时使用。强调先读规则与同类实现、先给方案、严格考虑缓存与索引、避免重复扫表、保持领域语义清晰，并完成必要验证。
metadata:
  short-description: 通用工程项目规则
---

# 通用工程项目规则

## 使用方式

当任务发生在一个已有代码库中时，先用本 skill 建立工作框架，再按当前仓库的 `AGENTS.md`、README、贡献指南或等价规则执行。仓库内规则优先于本 skill。

## 核心流程

1. 先读当前项目规则与相关代码，不凭记忆回答差异、原因或实现细节。
2. 改动前先检索同类实现，沿用已有分层、命名、交互、错误处理和缓存模式。
3. 写代码前给出明确方案：修改文件 + 主要改动点 + 查询/缓存/索引影响。
4. 实现时只做当前问题需要的改动，避免扩散式重构。
5. 改完后做自检：格式化、编译/测试、相关检索、SQL 执行计划或索引命中检查。
6. 如果启动了临时服务、脚本或后台进程，验证结束后必须关闭并说明清理结果。

## 必须注意

- 不要绕过项目已有缓存、model/repo、service、模板组件或设计系统。
- 不要把单源查询写成通用多源交集流程。
- 不要重复扫描同一张表，除非已经说明原因和成本。
- 不要在业务代码中随意增加隐式过滤、阈值、上限或下限。
- 不要混用展示 ID、内部 ID、外部 ID、路由参数和数据库关联键。

## Reference

- 通用代码库工作流：`references/codebase-workflow.md`
- 数据库、缓存与性能：`references/database-cache-performance.md`
- 前端改动工作流：`references/frontend-workflow.md`
- 领域语义与链接规则：`references/domain-semantics.md`
- 验证与收尾清单：`references/verification-checklist.md`
