# RDR 功能概览

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                     RDR Web Interface                        │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Upload     │  │   History    │  │    Ops       │      │
│  │   Manager    │  │   Tracking   │  │  Analysis    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                               │
├─────────────────────────────────────────────────────────────┤
│                      API Layer                                │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  /api/upload          /api/history      /api/ops/analysis    │
│  /api/progress        /list             /api/ops/health      │
│                                                               │
├─────────────────────────────────────────────────────────────┤
│                    Analysis Engine                            │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Counter    │  │ OpsAnalyzer  │  │   History    │      │
│  │              │  │              │  │   Manager    │      │
│  │ - TypeNum    │  │ - Anomalies  │  │              │      │
│  │ - TypeBytes  │  │ - Hotspots   │  │ - Persist    │      │
│  │ - Prefixes   │  │ - Health     │  │ - Load       │      │
│  │ - Slots      │  │ - Recommend  │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                               │
├─────────────────────────────────────────────────────────────┤
│                    RDB Decoder                                │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
                      RDB Files (.rdb)
```

## 功能模块

### 1. 文件上传模块
- ✅ 拖拽上传
- ✅ 多文件上传
- ✅ 进度跟踪
- ✅ 实时日志
- ✅ 后台解析

### 2. 历史管理模块
- ✅ 自动保存分析历史
- ✅ 持久化到 JSON 文件
- ✅ 侧边栏快速访问
- ✅ 元数据存储（大小、键数、内存）
- ✅ 自动去重
- ⏳ 历史趋势对比（计划中）
- ⏳ 历史记录导出（计划中）

### 3. 基础分析模块
- ✅ 键数量统计
- ✅ 内存使用统计
- ✅ 数据类型分布
- ✅ Top 100 大键
- ✅ 键前缀分析
- ✅ 长度级别分布
- ✅ 集群槽位分析

### 4. 运维分析模块

#### 4.1 健康评分系统
- ✅ 0-100 分综合评分
- ✅ 5 级健康状态
- ✅ 可视化健康仪表盘
- ✅ 实时健康检查 API

#### 4.2 异常检测引擎
```
异常类型          检测维度              阈值
─────────────────────────────────────────────
超大键            单键大小              10MB/50MB
内存热点          前缀占比              30%
键爆炸            微小键数量            >30%
数据类型主导      类型占比              50%
键数量过高        总键数                1000万
巨型集合          集合元素数            100万
类型效率低        变异系数              <50%
槽位不平衡        槽位差异              50%
```

#### 4.3 内存热点分析
- ✅ 按键前缀分析 (Top 20)
- ✅ 按数据类型分析
- ✅ 内存占比可视化
- ✅ 平均键大小计算
- ⏳ 热点演变趋势（计划中）

#### 4.4 键模式分析
- ✅ Top 50 常见模式识别
- ✅ 模式统计（数量、内存、占比）
- ✅ 示例键展示
- ⏳ 模式推荐优化（计划中）

#### 4.5 类型效率分析
```
指标              说明                    用途
─────────────────────────────────────────────
平均大小          Average Size           了解平均情况
中位数            Median Size            排除极值影响
P95/P99           Percentile             识别异常值
效率评分          Efficiency Score       量化效率
浪费内存          Wasted Memory          优化潜力
建议类型          Optimal Type           优化方向
```

#### 4.6 智能推荐系统
```
优先级    类别           典型建议                     实施难度
────────────────────────────────────────────────────
1        TTL           设置大键过期时间              Medium
2        Memory        启用内存驱逐策略              Low
3        Performance   使用 Hash 优化小字符串         High
4        Monitoring    启用慢日志                    Low
5        Cluster       槽位重新平衡                  Medium
```

### 5. 可视化模块

#### 5.1 图表类型
- ✅ 饼图 (数据类型分布)
- ✅ 柱状图 (内存使用、长度级别)
- ✅ 进度条 (百分比展示)
- ✅ 仪表盘 (健康评分)
- ✅ 表格 (大键列表、前缀统计)

#### 5.2 交互功能
- ✅ 标签页切换 (Overview/Ops/Details)
- ✅ 可折叠面板
- ✅ 悬停提示
- ✅ 排序功能
- ✅ 实时数据刷新

## 数据流

### 上传流程
```
用户上传 RDB
    │
    ▼
文件验证 (.rdb)
    │
    ▼
保存到 uploads/
    │
    ▼
创建解析进度
    │
    ▼
后台启动 Decoder
    │
    ├──▶ 解析 RDB
    │
    ├──▶ Counter 统计
    │    ├─ TypeNum/TypeBytes
    │    ├─ Prefixes
    │    ├─ LargestKeys
    │    └─ Slots
    │
    ├──▶ OpsAnalyzer 分析
    │    ├─ 检测异常
    │    ├─ 分析热点
    │    ├─ 评估效率
    │    ├─ 计算健康分数
    │    └─ 生成建议
    │
    ▼
保存到 History
    │
    ▼
通知前端完成
```

### 查询流程
```
用户访问分析页
    │
    ▼
加载 Overview 数据
    ├─ Counter 数据
    └─ 基础统计
    │
    ▼
用户切换到 Ops 标签
    │
    ▼
调用 /api/ops/analysis
    │
    ▼
OpsAnalyzer 实时分析
    ├─ 计算健康评分
    ├─ 检测异常
    ├─ 分析热点
    └─ 生成建议
    │
    ▼
返回 JSON 数据
    │
    ▼
前端渲染展示
```

## API 端点总览

### 文件管理
```
POST   /api/upload                 上传 RDB 文件
GET    /api/progress/:path         获取解析进度
GET    /api/stream/:path           流式日志输出
```

### 实例管理
```
GET    /list                       列出所有实例
GET    /api/history                获取历史记录
GET    /instance/:path             查看实例分析
GET    /terminal/:path             查看解析终端
```

### 运维分析
```
GET    /api/ops/analysis/:path     完整运维分析
GET    /api/ops/health/:path       健康快速检查
GET    /api/ops/anomalies/:path    仅异常列表
GET    /api/ops/recommendations/:path  仅优化建议
```

## 性能指标

### 解析性能
```
文件大小         解析时间（估算）        内存占用
──────────────────────────────────────────────
100 MB          ~5 秒                  ~200 MB
1 GB            ~30 秒                 ~500 MB
5 GB            ~2 分钟                ~1.5 GB
10 GB           ~4 分钟                ~3 GB
```

### API 响应时间
```
端点                    响应时间（估算）
───────────────────────────────────────
/api/ops/health        <100 ms
/api/ops/anomalies     <200 ms
/api/ops/analysis      <500 ms
/api/upload            取决于文件大小
```

## 技术栈

### 后端
- **语言**: Go 1.x
- **Web 框架**: httprouter
- **RDB 解析**: github.com/dongmx/rdb
- **模板引擎**: Go html/template
- **静态资源**: go-bindata

### 前端
- **UI 框架**: Bootstrap 3
- **图表库**: Chart.js
- **图标**: Font Awesome
- **脚本**: Vanilla JavaScript

### 存储
- **历史记录**: JSON 文件 (history.json)
- **解析结果**: 内存存储 (SafeMap)
- **上传文件**: 本地文件系统 (uploads/)

## 扩展性

### 当前支持
- ✅ 单机模式
- ✅ 多文件并发解析
- ✅ 实时进度跟踪
- ✅ 历史记录持久化

### 计划支持
- ⏳ 分布式部署
- ⏳ 数据库存储
- ⏳ 用户认证
- ⏳ 角色权限
- ⏳ 定时任务
- ⏳ 告警通知
- ⏳ 报表导出

## 安全考虑

### 当前实现
- ✅ 文件类型验证 (.rdb)
- ✅ 文件大小限制 (10GB)
- ✅ 上传互斥锁
- ✅ 安全文件路径处理

### 建议加强
- ⚠️ 添加用户认证
- ⚠️ 启用 HTTPS
- ⚠️ 实施访问控制
- ⚠️ 审计日志
- ⚠️ 敏感数据脱敏

## 部署建议

### 开发环境
```bash
# 直接运行
go run main.go web -p 8080

# 或编译后运行
go build
./rdr web -p 8080
```

### 生产环境
```bash
# 使用 systemd
[Unit]
Description=RDR Web Service
After=network.target

[Service]
Type=simple
User=rdr
WorkingDirectory=/opt/rdr
ExecStart=/opt/rdr/rdr web -p 8080
Restart=always

[Install]
WantedBy=multi-user.target

# Nginx 反向代理
location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

### Docker 部署
```dockerfile
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o rdr

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/rdr .
EXPOSE 8080
CMD ["./rdr", "web", "-p", "8080"]
```

## 监控指标

### 建议监控
```
指标                   说明                   告警阈值
────────────────────────────────────────────────────
health_score          健康评分               <60
critical_issues       严重问题数             >0
warning_count         警告数量               >5
parse_duration        解析耗时               >10分钟
memory_usage          内存使用               >80%
disk_usage            磁盘使用               >90%
```

## 故障排查

### 常见问题

#### 1. 上传失败
```
检查项:
- 文件是否为 .rdb 格式
- 文件是否 <10GB
- uploads/ 目录是否有写权限
- 磁盘空间是否充足
```

#### 2. 解析卡住
```
检查项:
- 查看 /terminal/:path 页面日志
- 检查服务器内存是否充足
- 检查 RDB 文件是否损坏
- 重启服务重试
```

#### 3. API 无响应
```
检查项:
- 检查实例是否解析完成
- 查看服务器日志
- 检查网络连接
- 验证实例名称是否正确
```

## 贡献指南

欢迎贡献！请参考：
1. Fork 项目
2. 创建特性分支
3. 提交改动
4. 推送到分支
5. 创建 Pull Request

## 许可证

Apache License 2.0
