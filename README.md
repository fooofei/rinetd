# rinetd

This is a golang version port of linux tool [`rinetd`](https://github.com/samhocevar/rinetd).


## Feature

- config file each line is a listen chain
- each chain is a tcp listen or a udp listen
- support config file reload, file changed will apply for no need to exit app
- support comment, "//" "#" at line begin or comment at line end


## Compile

```shell
go build -v
```
will generate `rinetd` executable file.

## Use

Unlike the c version of rinetd, this rinetd use addr pairs writed in `rinetd.conf`.

The addr pairs format in `rinetd.conf` looks like 
```
0.0.0.0:44444   127.0.0.1:55555     tcp
0.0.0.0:5679    127.0.0.1:8200      udp
```

first line represents rinetd will listen on `0.0.0.0:44444` for TCP, 
pipe read/write from this port to `127.0.0.1:55555`.

second line represents rinetd will listen on `0.0.0.0:5679` for UDP, 
pipe read/write from this port to `127.0.0.1:8200`.

You can also write commnet line begin with `#` or `//`.

WARN:The `deny` and `allow` not supported.


## Others

https://github.com/go-gost/gost 旧版本 https://github.com/ginuerzh/gost

使用体验：配置文件描述太复杂了 单位面积有效信息量没当前 rinetd 配置的每行一个映射简单

而且也没有看到可以支持配置动态刷新功能。（v2 版本根本不支持全局配置文件，而是单个功能的配置文件，
单个功能的配置文件里可以进行 reload 功能开启）

https://github.com/inlets/mixctl 

使用体验：阅读代码发现代码没有我们优雅，没有增加 io.Copy Buffer大小，没有优雅关闭TCP单向通道。
没有UDP。rinetd可以转发UDP。
