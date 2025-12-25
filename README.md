# GO-VSOA 使用Pure GO实现的VSOA SDK框架

完全使用 Go 的 VSOA 开发库，不需要任何 C 库、汇编和 CGO 的依赖。
[VSOA](https://www.acoinfo.com/product/5330/)是翼辉自主设计的任务关键型分布式微服务架构。此项目是翼辉在原有基础上推出的 Go 语言官方 SDK。

## ChangeLog

### 2025/12/25 V1.1.10

- Fix client auto reconnect set error.

### 2025/08/07 V1.1.9

- Add IsSubscribed API.

### 2025/07/29 V1.1.8

- Add auto reconnect check.

### 2025/07/29 V1.1.7

- Add onconnect/ondisconnect default implementation.

### 2025/07/28 V1.1.6

- Fix defaultOnClientHandler error.

### 2025/07/28 V1.1.5

- Add client delete API.

### 2025/07/28 V1.1.4

- Add client onconnect implementation.

### 2025/07/26 V1.1.3

- Add client reconnect implementation.

### 2025/07/23 V1.1.2

- Fix server connection push error.
- Fix server connection receive error.

### 2025/07/21 V1.1.1

- Add client default option.

### 2025/07/16 V1.1.0

- Fix some bugs on Pub/Sub
- Drop Assembly
- Optimized message processing
- Change module name to vsoa

### 2024/11/26 V1.0.5

- Fix some bugs
- Add Server NEW Feature: OnClient func
- Add Server NEW Feature: RAW Publish by trigger

### 2023/12/25 v1.0.4

- Happy Christmas Day!
- Server add Count API
- Better stream test example with `client_file_transfer_test.go`
- Add Client NEW Feature: Regulator

### 2023/10/12 V1.0.3

- Change Server API!
- Server now can handle widecard match
- Change module URL to gitee.com

### 2023/10/08 V1.0.1

- Release V1.0.1 ver
- Support RPC/SUB/UNSUB/PUB/DATAGRAM
- Not Supoort QoS

## 目前可以使用的平台

| 系统 | 支持 |
-------- | -----
| Windows | YES |
| MacOS | YES |
| Linux | YES |
| FreeBSD | YES |
| SylixOS | YES* |

SylixOS 需要使用支持编译 SylixOS 系统的 Golang 编译器编译。发布在 [GO SylixOS SDK](https://github.com/acoinfo/go/releases)

## 简单RPC示例代码

本范例假设使用者已经学习过 Go 语言基础，同时能够理解 go module 的工作逻辑。
我们将写一个可以开关控制灯光，且能够读取当前灯光状态的C/S例程。
其他示例代码在配套发布的example文件夹中可以找到。

### 预处理（本地离线模式）

Golang编译器需要设置为启用go module模式。

~~~bash  
go env -w GO111MODULE=on
~~~  

在src文件夹下创建两个文件夹`vsoa-server`、`vsoa-client`，分别放置客户端和服务端
分别进入两个文件夹并将以下文件保存为go.mod

`vsoa-server`文件夹下：

~~~mod  
module vsoa-server

go 1.24

require github.com/acoinfo/vsoa v1.0.5
~~~

`vsoa-client`文件夹下：

~~~mod
module vsoa-client

go 1.24

require github.com/acoinfo/vsoa v1.0.5
~~~  

### 编写服务端

文件名：`server.go`

~~~go
package main

import (
    "encoding/json"
    "time"

    "github.com/acoinfo/vsoa/protocol"
    "github.com/acoinfo/vsoa/server"
)

type RpcLightParam struct {
    LightStatus bool `json:"Light On"`
}

var lightstatus = true

func getLight(req, resp *protocol.Message) {
	status, _ := json.Marshal(lightstatus)
	resp.Param, _ = json.RawMessage(`{"Light On":` + string(status) + `}`).MarshalJSON()
}

func setLight(req, resp *protocol.Message) {
	var p RpcLightParam
	if err := json.Unmarshal(req.Param, &p); err != nil {
		return
	}
	lightstatus = p.LightStatus
	status, _ := json.Marshal(lightstatus)
	resp.Param, _ = json.RawMessage(`{"Light On":` + string(status) + `}`).MarshalJSON()
}

func main() {
	s := server.NewServer("golang VSOA server",
		server.Option{Password: "123456"})

	s.On("/light", protocol.RpcMethodGet, getLight)
	s.On("/light", protocol.RpcMethodSet, setLight)

	if err := s.Serve("localhost:3001"); err != nil {
		log.Fatal(err)
	}
}

~~~  

### 编写客户端

文件名：`client.go`

~~~go
package main

import (
	"encoding/json"
	"fmt"
	"vsoa-examples/rpc"
	"log"

	"github.com/acoinfo/vsoa/client"
	"github.com/acoinfo/vsoa/protocol"
)

func vsoaRpcCall(c *client.Client, method protocol.RpcMessageType, param json.RawMessage) bool {
	req := protocol.NewMessage()
	req.Param = param

	reply, err := c.Call(rpc.RpcExampleURL, protocol.TypeRPC, method, req)
	if err != nil {
		fmt.Println("call error:", err)
		return false
	}

	var dst rpc.RpcLightParam
	if err := json.Unmarshal(reply.Param, &dst); err != nil {
		log.Println("unmarshal error:", err)
		return false
	}
	fmt.Printf("Seq:%d RPC %s Light On:%v\n", reply.SeqNo(), protocol.RpcMethodText(method), dst.LightStatus)
	return dst.LightStatus
}

func main() {
	c := client.NewClient(client.Option{Password: rpc.Password})
	if _, err := c.Connect("vsoa", rpc.ServerAddr); err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	lightOn := vsoaRpcCall(c, protocol.RpcMethodGet, nil)

	cmd, _ := json.Marshal(map[string]bool{"Light On": !lightOn})
	fmt.Println("cmd:", string(cmd))
	vsoaRpcCall(c, protocol.RpcMethodSet, cmd)

	vsoaRpcCall(c, protocol.RpcMethodGet, nil)
}
~~~

### 运行示例程序

通过命令行/终端程序分别进入`vsoa-server`、`vsoa-client`文件夹
执行 `go build` 指令，后运行生成的可执行文件:

~~~bash
cd vsoa-server
go build
./vsoa-server

# Windows下
.\vsoa-server.exe
~~~

~~~bash
cd vsoa-client
go build
./vsoa-client

# Windows下
.\vsoa-client.exe
~~~

### 预期结果

服务器程序不显示任何输出。
客户端程序输出类似如下打印：

~~~bash
Seq: 1 RPC Get  Light On: true
Seq: 2 RPC Set  Light On: false
Seq: 3 RPC Get  Light On: false
~~~
