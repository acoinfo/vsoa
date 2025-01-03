# GO-VSOA 使用Pure GO实现的VSOA SDK框架

完全使用Go和Go汇编实现的VSOA开发库，不需要任何C库和CGO的依赖。
[VSOA](https://www.acoinfo.com/product/5330/)是翼辉自主设计的任务关键型分布式微服务架构。此项目是翼辉在原有基础上推出的Go语言官方SDK。

## ChangeLog

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

### 架构平台

| 架构 | 支持 |  
-------- | -----  
| amd64架构 | YES |  
| x86架构 | NO |  
| aarch64架构 | YES |
| arm架构 | NO |  

### 操作系统平台

| 系统 | 支持 |  
-------- | -----  
| Windows | YES |  
| MacOS | YES |  
| Linux | YES |  
| FreeBSD | YES |  
| SylixOS | YES* |  

SylixOS需要使用支持编译SylixOS系统的Golang编译器编译。发布在[GO add sylixos support](https://github.com/go-sylixos/go/releases)

## 发布形态

目前go-vsoa使用本地仓库的型式进行发布，在V1.1.0版本以后将修改为线上与线下同步发布。

## 简单RPC示例代码

本范例假设使用者已经学习过Go语言基础，同时能够理解go module的工作逻辑。  
我们将写一个可以开关控制灯光，且能够读取当前灯光状态的C/S例程。  
其他示例代码在配套发布的example文件夹中可以找到。

### 预处理（本地离线模式）

Golang编译器需要设置为启用go module模式。

~~~bash  
go env -w GO111MODULE=on
~~~  

在src文件夹下创建两个文件夹`go-vsoa-server`、`go-vsoa-client`，分别放置客户端和服务端  
分别进入两个文件夹并将以下文件保存为go.mod

`go-vsoa-server`文件夹下：

~~~mod  
module go-vsoa-server

go 1.20

require github.com/go-sylixos/go-vsoa v1.0.5
~~~  

`go-vsoa-client`文件夹下：

~~~mod  
module go-vsoa-client

go 1.20

require github.com/go-sylixos/go-vsoa v1.0.5
~~~  

### 编写服务端

文件名：`server.go`  

~~~go  
package main

import (
    "encoding/json"
    "time"

    "github.com/go-sylixos/go-vsoa/protocol"
    "github.com/go-sylixos/go-vsoa/server"
)

type RpcLightParam struct {
    LightStatus bool `json:"Light On"`
}

var lightstatus = true

func startServer() {
    // 初始化Go VSOA server，此范例设置了server的密码为123455
    // 如果不需要密码且没有其他的需求，此部分可以不设置，直接传空到server.NewServer函数中
    serverOption := server.Option{
        Password: "123456",
    }
    s := server.NewServer("golang VSOA server", serverOption)

    // 注册 light URL，RPC的GET方法
    // 允许授权的客户端查询灯现在的状态
    handleLightGet := func(req, res *protocol.Message) {
        status, _ := json.Marshal(lightstatus)
        res.Param, _ = json.RawMessage(`{"Light On":` + string(status) + `}`).MarshalJSON()
        res.Data = req.Data
    }
    s.On("/light", protocol.RpcMethodGet, handleLightGet)

    // 注册 light URL，RPC的SET方法
    // 允许授权的客户端操作开灯或关灯
    handleLightSet := func(req, res *protocol.Message) {
        reqParam := new(RpcLightParam)
        err := json.Unmarshal(req.Param, reqParam)

        if err != nil {
            status, _ := json.Marshal(lightstatus)
            res.Param, _ = json.RawMessage(`{"Light On":` + string(status) + `}`).MarshalJSON()
            return
        }

        lightstatus = reqParam.LightStatus
        status, _ := json.Marshal(lightstatus)
        res.Param, _ = json.RawMessage(`{"Light On":` + string(status) + `}`).MarshalJSON()
        res.Data = req.Data
    }
    s.On("/light", protocol.RpcMethodSet, handleLightSet)

    go func() {
        _ = s.Serve("0.0.0.0:3001")
    }()
}

func main() {
    startServer()

    for {
        time.Sleep(1 * time.Second)
    }
}

~~~  

### 编写客户端

文件名：`client.go`  

~~~go  
package main

import (
    "encoding/json"
    "errors"
    "fmt"

    "github.com/go-sylixos/go-vsoa/client"
    "github.com/go-sylixos/go-vsoa/protocol"
)

type RpcLightParam struct {
    LightStatus bool `json:"Light On"`
}

var lightstatus = false

func VsoaRpcCall() {
    clientOption := client.Option{
        Password: "123456",
    }

    c := client.NewClient(clientOption)
    _, err := c.Connect("vsoa", "localhost:3001")
    if err != nil {
        fmt.Println(err)
    }
    defer c.Close()

    req := protocol.NewMessage()

    // 查询现在light的状态
    reply, err := c.Call("/light", protocol.TypeRPC, protocol.RpcMethodGet, req)
    if err != nil {
        if err == errors.New(protocol.StatusText(protocol.StatusInvalidUrl)) {
            fmt.Println("Pass: Invalid URL")
        } else {
            fmt.Println(err)
        }
    } else {
        DstParam := new(RpcLightParam)
        json.Unmarshal(reply.Param, DstParam)
        lightstatus = DstParam.LightStatus
        fmt.Println("Seq:", reply.SeqNo(), "RPC Get ", "Light On:", DstParam.LightStatus)
    }

    // 如果现在灯是亮的就关闭它，如果现在灯是关的就打开它
    if lightstatus {
        req.Param, _ = json.RawMessage(`{"Light On":false}`).MarshalJSON()
    } else {
        req.Param, _ = json.RawMessage(`{"Light On":true}`).MarshalJSON()
    }
    reply, err = c.Call("/light", protocol.TypeRPC, protocol.RpcMethodSet, req)
    if err != nil {
        if err == errors.New(protocol.StatusText(protocol.StatusInvalidUrl)) {
            fmt.Println("Pass: Invalid URL")
        } else {
            fmt.Println(err)
        }
    } else {
        DstParam := new(RpcLightParam)
        json.Unmarshal(reply.Param, DstParam)
        fmt.Println("Seq:", reply.SeqNo(), "RPC Set ", "Light On:", DstParam.LightStatus)
    }

    // 查询执行操作后灯的状态
    reply, err = c.Call("/light", protocol.TypeRPC, protocol.RpcMethodGet, req)
    if err != nil {
        if err == errors.New(protocol.StatusText(protocol.StatusInvalidUrl)) {
            fmt.Println("Pass: Invalid URL")
        } else {
            fmt.Println(err)
        }
    } else {
        DstParam := new(RpcLightParam)
        json.Unmarshal(reply.Param, DstParam)
        fmt.Println("Seq:", reply.SeqNo(), "RPC Get ", "Light On:", DstParam.LightStatus)
    }
}

func main() {
    VsoaRpcCall()
}
~~~  

### 运行示例程序

通过命令行/终端程序分别进入`go-vsoa-server`、`go-vsoa-client`文件夹  
执行 `go build` 指令，后运行生成的可执行文件:

~~~bash
cd go-vsoa-server
go build
./go-vsoa-server

# Windows下 
.\go-vsoa-server.exe
~~~  

~~~bash
cd go-vsoa-client
go build
./go-vsoa-client

# Windows下 
.\go-vsoa-client.exe
~~~  

### 预期结果

服务器程序不显示任何输出。  
客户端程序输出类似如下打印：  

~~~bash
Seq: 1 RPC Get  Light On: true
Seq: 2 RPC Set  Light On: false
Seq: 3 RPC Get  Light On: false
~~~
