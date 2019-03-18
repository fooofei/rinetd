# rinetd

This is a golang version port of linux tool [`rientd`](https://github.com/samhocevar/rinetd).


## Compile

```shell
go build -v
```
will generate `rinetd` file.

## Use

Same like the c version of rinetd, this rinetd use addr pairs write in `rinetd.conf`.

The addr pairs format look like
```
0.0.0.0 12345 127.0.0.1 2345
```

rinetd will listen on `0.0.0.0:12345` and the TCP connection from this port will pipe read/write to `127.0.0.1:2345`.

You can also write commnet line begin with `#` or `//`.

WARN:The `deny` and `allow` not supported.



## TODO

The `parse.peg` also young, not strong.

