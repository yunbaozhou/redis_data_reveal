# RDR Web 功能特性详解

## 项目概述

本项目基于原有的 RDR (Redis Data Reveal) 工具,新增了完整的 Web 端文件上传和实时分析功能,使用户能够通过浏览器直接分析 Redis RDB 文件,无需任何命令行操作。

## 核心功能

### 1. 文件上传界面

#### 设计特点
- **现代化 UI**: 采用渐变色背景和卡片式设计
- **直观操作**: 大图标和清晰的提示文字
- **响应式设计**: 适配不同屏幕尺寸

#### 上传方式
1. **拖拽上传**
   - 支持从文件管理器拖拽 .rdb 文件到上传区域
   - 拖拽时有视觉反馈(边框变色、区域缩放)

2. **点击上传**
   - 点击上传区域打开文件选择对话框
   - 支持多文件同时选择

3. **文件验证**
   - 自动过滤非 .rdb 文件
   - 显示文件名和文件大小
   - 支持移除已选文件

#### 上传流程
```
选择文件 → 显示文件列表 → 点击"开始分析" → 上传到服务器 → 实时解析 → 跳转到结果页面
```

### 2. 实时解析功能

#### 后端处理
- **文件存储**: 上传的文件保存在 `uploads/` 目录
- **异步解析**: 使用 goroutine 并发解析 RDB 文件
- **内存缓存**: 解析结果缓存在内存中,无需重复解析

#### 解析过程
1. 接收上传的文件
2. 验证文件格式(.rdb)
3. 保存到服务器
4. 启动后台解析任务
5. 生成统计数据
6. 返回实例标识

#### API 设计
```
POST /api/upload        # 上传文件
GET  /list             # 获取所有实例列表
GET  /instance/:path   # 查看特定实例的分析结果
```

### 3. 数据可视化展示

#### 概览卡片
四个信息卡片显示核心指标:
- **Total Keys**: 总键数量
- **Total Memory**: 总内存使用
- **Data Types**: 数据类型数量
- **RDB File**: 文件名称

每个卡片都有:
- 彩色背景(蓝、绿、黄、红)
- 图标标识
- 大号数字显示
- 人性化格式(千位分隔符)

#### 图表展示

##### 1. 类型分布饼图
- **Key Count by Type**: 按类型统计键数量
- **Memory Usage by Type**: 按类型统计内存使用
- 使用 Chart.js 绘制交互式饼图
- 支持悬停显示详细信息
- 彩色区分不同类型

##### 2. 长度级别柱状图
- **Length Level Count**: 不同长度级别的键数量
- **Length Level Memory**: 不同长度级别的内存使用
- 按数据类型分Tab展示
- 横向柱状图,易于比较

#### 表格展示

##### 1. Top 100 最大 Keys
| # | Key | Type | Memory | Elements |
|---|-----|------|--------|----------|
| 1 | key_name | string | 1.2 MB | 1000 |

特点:
- 序号编号
- Key 名称使用代码格式
- 类型标签(带颜色)
- 人性化内存显示(KB/MB/GB)
- 元素数量(千位分隔符)
- 支持点击列头排序

##### 2. Key 前缀分析
按数据类型分组显示:
- Key 前缀
- 占用内存
- Key 数量
- 内存占比进度条

特点:
- 多Tab切换不同类型
- 进度条可视化内存占比
- 支持排序

### 4. 技术实现

#### 前端技术
```
- HTML5 (拖拽 API)
- JavaScript (原生 ES6+)
- Bootstrap 3.x (UI 框架)
- Chart.js (图表库)
- jQuery (DOM 操作)
```

#### 后端技术
```
- Go 1.x
- httprouter (HTTP 路由)
- go-bindata (资源嵌入)
- encoding/json (JSON 处理)
- multipart/form-data (文件上传)
```

#### 核心包
```
github.com/xueqiu/rdr/decoder     # RDB 解码器
github.com/xueqiu/rdr/dump        # 数据统计
github.com/dongmx/rdb             # RDB 解析
github.com/dustin/go-humanize     # 数字格式化
github.com/julienschmidt/httprouter # 路由
```

## 文件结构

```
rdr/
├── main.go                    # 主程序入口
├── dump/
│   ├── show.go               # Web 服务器
│   ├── upload.go             # 文件上传处理(新增)
│   ├── render.go             # 页面渲染
│   ├── template.go           # 模板函数
│   ├── counter.go            # 数据统计
│   └── dump.go               # RDB 解析
├── views/
│   ├── upload.html           # 上传页面(新增)
│   ├── enhanced_revel.html   # 增强版结果页面(新增)
│   ├── revel.html            # 原结果页面
│   ├── base.html             # 基础模板
│   └── ...
├── static/
│   ├── bootstrap/            # Bootstrap 框架
│   ├── chartjs/              # Chart.js 库
│   └── js/                   # JavaScript 文件
├── uploads/                  # 上传文件目录(新增)
├── README.md                 # 项目说明(更新)
├── WEB_USAGE.md             # Web 使用指南(新增)
└── FEATURES.md              # 功能特性文档(新增)
```

## 核心代码说明

### 1. 文件上传处理 (upload.go)

```go
// uploadHandler 处理文件上传请求
func uploadHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    // 1. 解析 multipart form
    r.ParseMultipartForm(maxUploadSize)

    // 2. 遍历上传的文件
    for _, fileHeader := range files {
        // 3. 保存文件到 uploads 目录
        // 4. 启动后台解析任务
        go func() {
            decoder := decoder.NewDecoder()
            rdb.Decode(f, decoder)
            counter := NewCounter()
            counter.Count(decoder.Entries)
            counters.Set(filename, counter)
        }()
    }

    // 5. 返回成功响应
    json.NewEncoder(w).Encode(response)
}
```

### 2. Web 服务器启动 (show.go)

```go
// ShowWeb 启动带上传功能的 Web 服务器
func ShowWeb(c *cli.Context) {
    InitHTMLTmpl()
    router := httprouter.New()

    // 路由配置
    router.GET("/", showUploadPage)           // 上传页面
    router.POST("/api/upload", uploadHandler)  // 上传 API
    router.GET("/instance/:path", rdbReveal)   // 结果页面
    router.GET("/list", listInstances)         // 实例列表

    http.ListenAndServe(":"+port, router)
}
```

### 3. 前端上传逻辑 (upload.html)

```javascript
// 拖拽事件处理
uploadArea.addEventListener('drop', (e) => {
    e.preventDefault();
    handleFiles(e.dataTransfer.files);
});

// 文件上传
async function uploadFiles() {
    const formData = new FormData();
    files.forEach(file => formData.append('files', file));

    const response = await fetch('/api/upload', {
        method: 'POST',
        body: formData
    });

    const result = await response.json();
    window.location.href = '/instance/' + result.instances[0];
}
```

## 性能优化

### 1. 并发处理
- 使用 goroutine 并发解析多个 RDB 文件
- 异步处理不阻塞主线程

### 2. 内存管理
- 解析结果缓存在内存中
- 使用 SafeMap 保证并发安全

### 3. 前端优化
- 文件验证在客户端进行
- 使用进度条提供用户反馈
- Chart.js 按需加载

## 安全考虑

### 1. 文件验证
- 检查文件扩展名(.rdb)
- 限制文件大小(默认 10GB)
- 使用 http.MaxBytesReader 防止内存溢出

### 2. 路径安全
- 使用 filepath.Base 防止路径遍历
- 上传目录权限控制

### 3. 并发控制
- 使用 Mutex 锁保护共享资源
- SafeMap 实现线程安全的数据存储

## 扩展性

### 1. 支持的扩展
- 可以添加更多数据可视化图表
- 可以实现文件列表管理(删除、重命名)
- 可以添加用户认证和权限控制
- 可以集成更多 Redis 分析工具

### 2. 配置选项
- 端口配置
- 上传目录配置
- 文件大小限制配置
- 缓存策略配置

## 使用场景

### 1. Redis 运维
- 分析生产环境 RDB 备份
- 识别大 Key 问题
- 内存使用优化

### 2. 开发调试
- 本地 RDB 文件分析
- 数据结构验证
- 性能问题排查

### 3. 团队协作
- 共享分析服务器
- 多人同时分析不同 RDB 文件
- 统一的分析界面

## 总结

本项目成功实现了:
1. ✅ 现代化的文件上传界面(支持拖拽)
2. ✅ 实时 RDB 文件解析
3. ✅ 多样化的数据可视化展示
4. ✅ RESTful API 设计
5. ✅ 响应式 Web 界面
6. ✅ 完整的文档支持

相比原有的命令行工具,Web 版本提供了更友好的用户体验,更直观的数据展示,以及更便捷的操作方式。
