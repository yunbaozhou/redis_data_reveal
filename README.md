RDR: redis data reveal
=================================================

RDR(redis data reveal) is a tool to parse redis rdbfile. Comparing to [redis-rdb-tools](https://github.com/sripathikrishnan/redis-rdb-tools), RDR is implemented by golang, much faster (5GB rdbfile takes about 2mins on my PC).

## New Feature: Enhanced Web Interface with History Tracking

Now you can upload and analyze RDB files directly through a modern web interface!

### Quick Start

```bash
# Start the web server
./rdr.exe web -p 8080

# Open http://localhost:8080 in your browser
# Drag & drop your .rdb files or click to upload
# View real-time analysis results with charts and tables
```

**📖 快速上手**: 查看 [快速开始指南](QUICK_START.md) 了解详细使用步骤

**🔧 运维分析**: 查看 [运维功能文档](OPS_FEATURES.md) 了解所有运维分析功能

### New Features

- **Modern UI**: Clean, gradient-based interface with modal upload dialog
- **Integrated Upload**: Upload button directly on main page and analysis pages
- **Analysis History**: Automatic tracking of all analyzed files with persistent storage
- **History Sidebar**: Quick access to previously analyzed files from any page
- **Drag & Drop**: Easy file upload with visual feedback
- **Real-time Progress**: Live upload and parsing progress tracking
- **Auto-persistence**: All analysis history saved to `history.json` for future sessions

### Features

#### 基础分析
- Drag & drop file upload with modern UI
- Real-time RDB file parsing with progress tracking
- Interactive data visualization with Chart.js
- Top 100 largest keys analysis
- Key prefix statistics and grouping
- Memory usage breakdown by type
- Length level distribution
- Analysis history persistence
- Quick navigation between analyzed files

#### 🆕 运维增强分析
- **健康评分系统**: 0-100 分综合评价 Redis 健康状况
- **异常检测**: 自动识别超大键、内存热点、键爆炸等问题
- **内存热点分析**: 按前缀和类型多维度分析内存分布
- **键模式分析**: 识别常见命名模式和业务模块占用
- **类型效率评估**: 评估数据结构使用效率
- **集群槽位分析**: 检测集群负载均衡问题
- **智能优化建议**: 基于分析结果提供可操作的优化建议
- **多级告警**: Critical/Warning/Info 三级异常分类

详见 [运维功能文档](OPS_FEATURES.md)

## Usage

```
NAME:
   rdr - a tool to parse redis rdbfile

USAGE:
   rdr [global options] command [command options] [arguments...]

VERSION:
   v0.0.1

COMMANDS:
     dump     dump statistical information of rdbfile to STDOUT
     show     show statistical information of rdbfile by webpage
     web      start web server with upload capability for analyzing RDB files
     keys     get all keys from rdbfile
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

```
NAME:
   rdr show - show statistical information of rdbfile by webpage

USAGE:
   rdr show [command options] FILE1 [FILE2] [FILE3]...

OPTIONS:
   --port value, -p value  Port for rdr to listen (default: 8080)
```

```
NAME:
   rdr keys - get all keys from rdbfile

USAGE:
   rdr keys FILE1 [FILE2] [FILE3]...
```

[Linux amd64 Download](https://github.com/xueqiu/rdr/releases/download/v0.0.1/rdr-linux)

[OSX Download](https://github.com/xueqiu/rdr/releases/download/v0.0.1/rdr-darwin)

[Windows Download](https://github.com/xueqiu/rdr/releases/download/v0.0.1/rdr-windows.exe)

After downloading maybe need add permisson to execute.

```
$ chmod a+x ./rdr*
```

## Exapmle
```
$ ./rdr show -p 8080 *.rdb
```
Note that the memory usage is approximate.
![show example](https://yqfile.alicdn.com/img_9bc93fc3a6b976fdf862c8314e34f454.png)

```
$ ./rdr keys example.rdb
portfolio:stock_follower_count:ZH314136
portfolio:stock_follower_count:ZH654106
portfolio:stock_follower:ZH617824
portfolio:stock_follower_count:ZH001019
portfolio:stock_follower_count:ZH346349
portfolio:stock_follower_count:ZH951803
portfolio:stock_follower:ZH924804
portfolio:stock_follower_count:INS104806
```

## License

This project is under Apache v2 License. See the [LICENSE](LICENSE) file for the full license text.
