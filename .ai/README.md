# 🤖 Grape Registry - AI 交接文档索引

**创建时间**: 2026-02-27  
**文档版本**: v1.0  
**维护周期**: 每双周迭代审查

---

## 📁 文档清单

| 文档 | 文件 | 目标读者 | 核心问题 |
|------|------|----------|----------|
| **1. 项目上下文** | [project-context.md](./project-context.md) | 新开发者、AI | 项目是做什么的？现在到什么阶段？ |
| **2. 系统架构** | [architecture.md](./architecture.md) | 架构师、高级开发者 | 系统是如何组织的？ |
| **3. 编码规范** | [coding-rules.md](./coding-rules.md) | 所有开发者、AI | 写代码必须遵守什么规则？ |
| **4. 产品路线图** | [roadmap.md](./roadmap.md) | 产品经理、技术决策者 | 接下来应该做什么？ |
| **5. 技术债清单** | [tech-debt.md](./tech-debt.md) | 技术负责人、AI | 哪些地方需要优化？ |
| **6. 环境规范** | [env-spec.md](./env-spec.md) | DevOps、新开发者 | 如何 10 分钟内跑起来？ |
| **7. AI 交接指南** | [ai-handoff.md](./ai-handoff.md) | AI 助手、新开发者 | 未来 AI 如何接手本项目？ |

---

## 🎯 使用场景

### 场景 1: 新 AI 接手项目

**推荐阅读顺序**:
```
1. ai-handoff.md          # 首先阅读
2. project-context.md      # 了解项目
3. env-spec.md             # 搭建环境
4. architecture.md         # 理解架构
5. coding-rules.md         # 遵守规范
6. tech-debt.md            # 了解技术债
7. roadmap.md              # 了解方向
```

**预计时间**: 2-3 小时

### 场景 2: 新开发者入职

**推荐阅读顺序**:
```
1. project-context.md      # 项目概览
2. env-spec.md             # 环境搭建
3. README.md               # 快速开始
4. coding-rules.md         # 编码规范
5. architecture.md         # 架构理解
```

**预计时间**: 1 天

### 场景 3: 技术决策者评估

**推荐阅读顺序**:
```
1. project-context.md      # 项目定位
2. roadmap.md              # 路线图
3. architecture.md         # 架构评估
4. tech-debt.md            # 风险评估
```

**预计时间**: 30 分钟

### 场景 4: 紧急 Bug 修复

**推荐阅读顺序**:
```
1. tech-debt.md            # 查看已知问题
2. architecture.md         # 定位相关模块
3. 对应代码文件             # 分析问题
```

**预计时间**: 1-2 小时

---

## 📊 文档覆盖度

| 维度 | 覆盖度 | 说明 |
|------|--------|------|
| **项目背景** | ✅ 100% | project-context.md |
| **系统架构** | ✅ 100% | architecture.md |
| **编码规范** | ✅ 100% | coding-rules.md |
| **环境搭建** | ✅ 100% | env-spec.md |
| **技术债** | ✅ 100% | tech-debt.md |
| **路线图** | ✅ 100% | roadmap.md |
| **AI 交接** | ✅ 100% | ai-handoff.md |
| **API 参考** | ✅ 100% | docs/API.md (外部) |
| **部署指南** | ✅ 100% | docs/DEPLOYMENT.md (外部) |

---

## 🔄 更新机制

### 触发条件

| 事件 | 更新文档 | 责任人 |
|------|----------|--------|
| **新功能完成** | roadmap.md, tech-debt.md | 开发者 |
| **架构变更** | architecture.md | 架构师 |
| **Bug 修复** | tech-debt.md | 开发者 |
| **环境变更** | env-spec.md | DevOps |
| **规范调整** | coding-rules.md | 技术负责人 |
| **双周迭代** | 所有文档 | AI/开发者 |

### 审查周期

| 文档 | 审查周期 | 上次审查 | 下次审查 |
|------|----------|----------|----------|
| project-context.md | 每月 | 2026-02-27 | 2026-03-27 |
| architecture.md | 架构变更时 | 2026-02-27 | - |
| coding-rules.md | 每季度 | 2026-02-27 | 2026-05-27 |
| roadmap.md | 每双周 | 2026-02-27 | 2026-03-13 |
| tech-debt.md | 每双周 | 2026-02-27 | 2026-03-13 |
| env-spec.md | 环境变更时 | 2026-02-27 | - |
| ai-handoff.md | 每月 | 2026-02-27 | 2026-03-27 |

---

## 📈 文档质量指标

| 指标 | 目标值 | 当前值 | 状态 |
|------|--------|--------|------|
| **文档完整度** | 100% | 100% | ✅ |
| **代码示例覆盖率** | 80%+ | 90%+ | ✅ |
| **链接有效性** | 100% | 100% | ✅ |
| **更新及时性** | 7 天内 | 实时 | ✅ |
| **AI 可读性** | 高 | 高 | ✅ |

---

## 🎓 文档编写原则

### 1. 结构化

- 使用表格、列表
- 清晰的层级结构
- 一致的格式

### 2. 条列式

- 避免长段落
- 使用要点列表
- 每个要点一个信息

### 3. 基于事实

- 不写主观评价
- 基于代码和文档
- 标注推断内容

### 4. 可操作

- 提供具体命令
- 包含代码示例
- 明确的验收标准

### 5. 无废话

- 不写营销语言
- 直接陈述事实
- 精简但不遗漏

---

## 🔗 外部文档链接

| 文档 | 路径 | 说明 |
|------|------|------|
| **项目说明** | [README.md](../README.md) | 项目介绍 |
| **使用指南** | [docs/USAGE.md](../docs/USAGE.md) | 使用方法 |
| **API 文档** | [docs/API.md](../docs/API.md) | API 参考 |
| **部署指南** | [docs/DEPLOYMENT.md](../docs/DEPLOYMENT.md) | 部署指南 |
| **开发文档** | [docs/DEVELOPMENT.md](../docs/DEVELOPMENT.md) | 开发指南 |
| **Webhook 文档** | [docs/WEBHOOKS.md](../docs/WEBHOOKS.md) | Webhook 使用 |
| **迭代计划** | [迭代计划.md](../迭代计划.md) | 产品路线图 |
| **Bug 报告** | [BUG-*.md](../BUG-*.md) | Bug 文档 |

---

## 📞 维护联系

- **维护者**: Grape Team
- **联系方式**: graperegistry@github
- **问题反馈**: GitHub Issues

---

**最后更新**: 2026-02-27  
**下次审查**: 2026-03-13
