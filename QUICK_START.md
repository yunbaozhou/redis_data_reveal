# RDR 快速开始指南

## 5 分钟上手

### 步骤 1: 启动服务
```bash
# Windows
rdr.exe web -p 8080

# Linux/Mac
./rdr web -p 8080
```

### 步骤 2: 上传 RDB 文件
1. 打开浏览器访问 `http://localhost:8080`
2. 点击 "Upload RDB File" 按钮或直接拖拽文件
3. 选择你的 `.rdb` 文件
4. 等待解析完成（进度条会显示）

### 步骤 3: 查看分析结果

#### Overview 标签（概览）
- **总体统计**: 键数量、内存使用、数据类型分布
- **可视化图表**: 键类型分布饼图、内存使用柱状图
- **Top 100 大键**: 最占内存的键列表
- **前缀分析**: 按键名前缀分组统计

#### 🆕 Ops Analysis 标签（运维分析）
- **健康评分**: 显示 0-100 分的健康分数
- **问题检测**: Critical/Warning/Info 三级异常
- **内存热点**: 识别内存集中的区域
- **优化建议**: 可操作的优化方案

## 运维分析功能详解

### 健康评分解读

#### 90-100 分 (优秀) 🟢
- 状态：健康
- 行动：保持现状，定期监控

#### 75-89 分 (良好) 🔵
- 状态：基本健康，有小问题
- 行动：查看 Warning 级别告警，计划优化

#### 60-74 分 (一般) 🟠
- 状态：需要关注
- 行动：尽快处理 Warning，评估 Critical

#### 40-59 分 (较差) 🟠
- 状态：有明显问题
- 行动：立即处理 Critical，制定优化计划

#### 0-39 分 (危险) 🔴
- 状态：严重问题
- 行动：紧急处理，可能影响生产

### 常见异常及处理

#### 1. 超大键检测 (Critical)
**问题**: 发现 >50MB 的键
```
异常: Extremely Large Key Detected
键名: user:session:abc123
大小: 85.3 MB
```

**影响**:
- 操作阻塞
- 内存压力
- 复制延迟

**处理方案**:
1. 拆分大键为多个小键
2. 使用 Hash 结构存储
3. 考虑分片存储

#### 2. 内存热点 (Warning)
**问题**: 单个前缀占用 >30% 内存
```
异常: Memory Hotspot Detected
前缀: cache:product:
占比: 45.2%
```

**影响**:
- 集群负载不均
- 单节点压力大

**处理方案**:
1. 重新设计键分布策略
2. 使用 hash tag 控制槽位
3. 考虑业务拆分

#### 3. 键爆炸 (Warning)
**问题**: 大量微小键
```
异常: Many Tiny Keys Detected
数量: 1500 个 <100 字节的键
```

**影响**:
- 键开销 > 值开销
- 内存浪费

**处理方案**:
```redis
# 原来：多个 String 键
SET user:1:name "Alice"
SET user:1:age "25"
SET user:1:city "NYC"

# 优化：使用 Hash 整合
HSET user:1 name "Alice" age "25" city "NYC"

# 可节省 50-70% 内存
```

#### 4. 巨型集合 (Warning)
**问题**: 单个集合 >100万 元素
```
异常: Huge Collection Detected
键名: active:users
元素数: 2,500,000
```

**影响**:
- 操作阻塞 Redis
- 延迟峰值

**处理方案**:
1. 拆分为多个小集合
2. 使用分片策略
3. 异步批量处理

### 优化建议使用

#### 查看建议
在 "Ops Analysis" 标签下找到 "Optimization Recommendations" 卡片

#### 建议优先级
- **Priority 1**: 最高优先级，建议立即处理
- **Priority 2**: 高优先级，近期处理
- **Priority 3**: 中优先级，计划处理
- **Priority 4-5**: 低优先级，可选处理

#### 实施难度
- **Low**: 配置修改，几分钟完成
- **Medium**: 需要一些运维操作，可能需要测试
- **High**: 需要代码改动，需要完整的开发测试流程

#### 示例建议

**建议 1: 启用内存驱逐策略**
```
优先级: 2
难度: Low
分类: Memory

描述: Database is using significant memory (>10GB)
操作: Configure 'maxmemory' and 'maxmemory-policy' in redis.conf
影响: Prevents OOM errors and automatic eviction

实施步骤:
1. 编辑 redis.conf
2. 添加: maxmemory 10gb
3. 添加: maxmemory-policy allkeys-lru
4. 重启 Redis 或 CONFIG SET
```

**建议 2: 使用 Hash 优化小字符串**
```
优先级: 3
难度: High
分类: Performance

描述: String type shows low efficiency (45.2%)
操作: Group related small string values into Hash structures
影响: Can reduce memory overhead by 30-50%

示例代码:
# Before
SET user:1001:name "Alice"
SET user:1001:email "alice@example.com"
SET user:1002:name "Bob"

# After
HSET users:1 name "Alice" email "alice@example.com"
HSET users:2 name "Bob" email "bob@example.com"
```

## API 使用示例

### 获取健康状态
```bash
curl http://localhost:8080/api/ops/health/your_rdb_filename.rdb
```

响应:
```json
{
  "health_score": 75,
  "health_status": "good",
  "critical_issues": 0,
  "warnings": 3,
  "total_anomalies": 8,
  "recommendations": 5
}
```

### 获取所有异常
```bash
curl http://localhost:8080/api/ops/anomalies/your_rdb_filename.rdb
```

### 获取完整分析
```bash
curl http://localhost:8080/api/ops/analysis/your_rdb_filename.rdb
```

## 集成到监控系统

### Prometheus 集成示例
```python
import requests
from prometheus_client import Gauge, generate_latest

health_score = Gauge('redis_rdb_health_score', 'Redis RDB Health Score')
critical_issues = Gauge('redis_rdb_critical_issues', 'Critical Issues Count')

def update_metrics():
    response = requests.get('http://localhost:8080/api/ops/health/prod.rdb')
    data = response.json()

    health_score.set(data['health_score'])
    critical_issues.set(data['critical_issues'])
```

### 告警规则示例
```yaml
# Prometheus Alert Rules
groups:
  - name: redis_rdb_alerts
    rules:
      - alert: RedisHealthScoreLow
        expr: redis_rdb_health_score < 60
        for: 5m
        annotations:
          summary: "Redis health score is low"

      - alert: RedisCriticalIssues
        expr: redis_rdb_critical_issues > 0
        annotations:
          summary: "Redis has critical issues"
```

## 最佳实践

### 1. 定期分析
```bash
# 每日定时分析脚本
#!/bin/bash
DATE=$(date +%Y%m%d)
redis-cli --rdb /backup/redis_${DATE}.rdb BGSAVE

# 等待 BGSAVE 完成
sleep 60

# 上传到 RDR 分析
curl -F "files=@/backup/redis_${DATE}.rdb" \
     http://rdr.example.com:8080/api/upload
```

### 2. 对比历史
- 保存每次分析的健康评分
- 制作趋势图表
- 识别异常变化

### 3. 建立基线
- 记录正常状态的各项指标
- 设置合理的告警阈值
- 定期回顾和调整

### 4. 优化工作流
```
分析 → 发现问题 → 制定方案 → 测试验证 → 上线 → 再次分析
```

## 常见问题

### Q: 为什么我的健康评分是 100 但还有 Info 级别异常？
A: Info 级别异常扣分较少（每个 3 分），某些 Info 异常是正常的，不影响整体健康。

### Q: 如何处理 "Slot Imbalance" 异常？
A: 使用 `redis-cli --cluster rebalance` 命令重新平衡槽位。

### Q: 建议中的 "wasted memory" 怎么计算的？
A: 基于 Redis 内部开销估算，包括指针、元数据等。

### Q: 可以分析在线 Redis 吗？
A: 需要先生成 RDB 文件。使用 `BGSAVE` 命令或从备份获取。

### Q: 分析大文件（>10GB）会很慢吗？
A: 会比较慢，建议:
- 在后台运行
- 查看终端页面的进度
- 考虑分片分析

## 下一步

1. **深入学习**: 阅读 [运维功能文档](OPS_FEATURES.md)
2. **定制化**: 根据业务场景调整分析重点
3. **自动化**: 集成到 CI/CD 或监控系统
4. **优化实践**: 建立优化知识库
5. **反馈改进**: 提出需求和建议

## 获取帮助

- GitHub Issues: https://github.com/xueqiu/rdr/issues
- 文档: [OPS_FEATURES.md](OPS_FEATURES.md)
- README: [README.md](README.md)
