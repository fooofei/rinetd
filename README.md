# rinetd

This is a golang version port of linux tool [`rinetd`](https://github.com/samhocevar/rinetd).


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

