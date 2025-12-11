RDR: redis data reveal
=================================================

RDR(redis data reveal) is a tool to parse redis rdbfile. Comparing to [redis-rdb-tools](https://github.com/sripathikrishnan/redis-rdb-tools), RDR is implemented by golang, much faster (5GB rdbfile takes about 2mins on my PC).

## New Feature: Web Upload & Analysis

Now you can upload and analyze RDB files directly through a web browser!

### Quick Start

```bash
# Start the web server
./rdr web -p 8080

# Open http://localhost:8080 in your browser
# Drag & drop your .rdb files or click to upload
# View real-time analysis results with charts and tables
```

### Features

- Drag & drop file upload
- Real-time RDB file parsing
- Interactive data visualization with charts
- Top 100 largest keys analysis
- Key prefix statistics
- Memory usage breakdown by type
- Length level distribution

See [WEB_USAGE.md](WEB_USAGE.md) for detailed documentation.

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
