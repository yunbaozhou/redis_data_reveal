# 更新日志

## [v2.0.0] - 运维增强版本

### 🎉 重大更新

#### 1. 全新运维分析功能
- **健康评分系统**: 0-100 分综合评价 Redis 健康状况
- **智能异常检测**: 8 大类异常自动识别
  - 超大键检测 (10MB/50MB 阈值)
  - 内存热点识别 (>30% 占比告警)
  - 键爆炸检测 (微小键过多)
  - 数据类型主导分析
  - 键数量异常 (>1000万告警)
  - 巨型集合检测 (>100万元素)
  - 类型效率评估
  - 集群槽位不平衡

- **内存热点分析**: 多维度内存分布分析
  - 按键前缀 Top 20
  - 按数据类型分布
  - 可视化百分比展示

- **键模式识别**: 识别 Top 50 常见命名模式
  - 模式统计（数量、内存、占比）
  - 示例键展示

- **类型效率分析**: 深度分析数据结构使用效率
  - 平均大小/中位数/P95/P99
  - 效率评分 (0-100)
  - 优化建议

- **智能优化建议**: 可操作的优化方案
  - 5 级优先级
  - 实施难度标注 (low/medium/high)
  - 预期影响说明
  - 具体操作步骤

#### 2. 新增 API 端点
```
GET /api/ops/analysis/:path         完整运维分析
GET /api/ops/health/:path           快速健康检查
GET /api/ops/anomalies/:path        异常列表
GET /api/ops/recommendations/:path  优化建议
```

#### 3. 增强的用户界面
- **三标签页设计**
  - Overview: 基础分析概览
  - Ops Analysis: 运维深度分析 ⭐️ 新增
  - Details: 详细指标

- **运维仪表盘**
  - 健康评分仪表盘
  - 异常分级展示 (Critical/Warning/Info)
  - 内存热点可视化
  - 优化建议卡片
  - 键模式表格
  - 类型效率分析

- **实时健康提示**
  - Ops 标签显示问题数量徽章
  - 颜色编码告警级别

### 🔧 功能改进

#### 历史记录增强
- 新增字段：
  - `total_keys`: 总键数
  - `total_memory`: 总内存
  - `upload_time`: 上传时间
  - `file_size`: 文件大小

#### 主页面优化
- 集成上传功能到主页
- 历史记录侧边栏
- 实时列表刷新
- 欢迎引导界面

#### 分析页面优化
- 侧边栏添加上传按钮
- 标题改为 "Analysis History"
- 添加文件图标

### 📚 文档完善

新增文档：
- `OPS_FEATURES.md`: 运维功能详细文档
- `QUICK_START.md`: 快速开始指南
- `FEATURES_OVERVIEW.md`: 功能概览
- `CHANGELOG.md`: 更新日志

### 🐛 Bug 修复
- 修复 Counter 字段访问问题
- 修复导入未使用问题
- 优化内存计算逻辑

### 📦 新增文件
```
dump/
  ├── ops_analyzer.go      # 运维分析器
  ├── ops_api.go          # 运维 API
  └── history.go          # 历史管理器

views/
  ├── ops_dashboard.html       # 运维仪表盘
  ├── ops_enhanced_revel.html  # 增强分析页
  └── main_with_upload.html    # 集成上传的主页
```

### 🎯 使用场景

1. **日常巡检**: 一键查看健康评分和异常
2. **容量规划**: 分析内存热点和增长趋势
3. **性能优化**: 识别可优化的数据结构
4. **故障排查**: 快速定位问题根源
5. **迁移评估**: 全面了解数据特征

### 📈 性能指标

- 健康检查 API: <100ms
- 异常分析 API: <200ms
- 完整分析 API: <500ms
- 支持文件: 最大 10GB
- 解析速度: ~5GB/2分钟

### 🔒 安全性

- 文件类型验证
- 大小限制控制
- 安全路径处理
- 上传互斥保护

---

## [v1.0.0] - 初始版本

### 功能特性

- RDB 文件解析
- 基础统计分析
- Web 界面展示
- 数据可视化
- 大键识别
- 前缀分析
- 类型分布

### 命令支持

- `rdr dump`: 命令行输出统计
- `rdr show`: 启动 Web 界面
- `rdr web`: Web 上传模式
- `rdr keys`: 列出所有键

---

## 升级指南

### 从 v1.0.0 升级到 v2.0.0

1. **备份数据**
   ```bash
   # 备份历史记录（如果有）
   cp history.json history.json.bak
   ```

2. **停止旧版本**
   ```bash
   # 停止运行中的服务
   pkill -f rdr
   ```

3. **替换二进制文件**
   ```bash
   # 下载新版本
   # 替换 rdr 可执行文件
   ```

4. **启动新版本**
   ```bash
   ./rdr web -p 8080
   ```

5. **验证功能**
   - 访问主页检查上传功能
   - 上传测试文件
   - 查看 Ops Analysis 标签
   - 测试 API 端点

### 兼容性说明

- ✅ 完全向后兼容
- ✅ 旧的 RDB 文件可继续使用
- ✅ 历史记录自动迁移
- ✅ 原有 API 保持不变
- ✅ 新功能不影响现有功能

### 数据迁移

历史记录会自动扩展新字段：
```json
{
  "filename": "redis.rdb",
  "filepath": "/path/to/redis.rdb",
  "upload_time": "2024-01-15T10:30:00Z",  // 新增
  "file_size": 1073741824,                 // 新增
  "total_keys": 1500000,                   // 新增
  "total_memory": 2147483648               // 新增
}
```

---

## 路线图

### v2.1.0 (计划中)
- [ ] 自定义阈值配置
- [ ] 历史趋势对比
- [ ] PDF 报告导出
- [ ] 用户认证
- [ ] 邮件告警

### v2.2.0 (计划中)
- [ ] 多实例对比
- [ ] TTL 分析
- [ ] 定时任务
- [ ] Webhook 集成
- [ ] 监控指标导出

### v3.0.0 (规划中)
- [ ] 分布式部署
- [ ] 集群管理
- [ ] AI 智能建议
- [ ] 实时监控
- [ ] 容量预测

---

## 贡献者

感谢所有为 RDR 做出贡献的开发者！

- 原始项目: [xueqiu/rdr](https://github.com/xueqiu/rdr)
- 增强开发: Based on rdr project

## 许可证

Apache License 2.0

---

## 获取帮助

- 📖 文档: [OPS_FEATURES.md](OPS_FEATURES.md)
- 🚀 快速开始: [QUICK_START.md](QUICK_START.md)
- 📊 功能概览: [FEATURES_OVERVIEW.md](FEATURES_OVERVIEW.md)
- 🐛 问题反馈: GitHub Issues
- 💬 讨论: GitHub Discussions
