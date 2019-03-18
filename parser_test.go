package main

import (
    "strings"
    "testing"
)

func Test1(t *testing.T) {
    content := `
# 必须换行结尾
# 省略协议默认为 tcp
1.2.3.4 5678/tcp 127.0.0.1 8100/tcp
`

    r,err := ParseReader("1.md", strings.NewReader(content))

    if err != nil {
        t.Error(err)
    }
    if _,ok := r.([]*Unit); !ok {
        t.Errorf("not unit %v", r)
    }

    expect := &Unit{BindAddr:"1.2.3.4", BindPort:5678, BindProto:"tcp",
        ConnectAddr:"127.0.0.1", ConnectPort:8100, ConnectProto:"tcp"}

    v := r.([]*Unit)
    if *v[0] != *expect {
        t.Errorf("%v != %v", v[0], expect)
    }
}