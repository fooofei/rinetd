# rinetd

This is a golang version port of linux tool [`rientd`](https://github.com/samhocevar/rinetd).


## Compile

```shell
go build -v
```
will generate `rinetd` file.

## Use

Unlike the c version of rinetd, this rinetd use addr pairs writed in `rinetd.conf`.

The addr pairs format look like in `rinetd.conf`
```
0.0.0.0:44444   127.0.0.1:55555     tcp
0.0.0.0:5679    127.0.0.1:8200      udp
```

first, rinetd will listen on `0.0.0.0:44444` for TCP and 

will pipe read/write from this port to `127.0.0.1:55555`.

second, rinetd will listen on `0.0.0.0:5679` for UDP and 

will pipe read/write from this port to `127.0.0.1:8200`.


You can also write commnet line begin with `#` or `//`.

WARN:The `deny` and `allow` not supported.

