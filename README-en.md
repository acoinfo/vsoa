# GO-VSOA: A Pure GO Implementation of VSOA SDK Framework

GO-VSOA is a development library for VSOA (Vision Service Oriented Architecture) completely implemented in Go and Go assembly, without any dependencies on C libraries and CGO.

[VSOA](https://www.acoinfo.com/product/5330/) is a mission-critical distributed microservices architecture designed by Yihui. This project is the official Go language SDK introduced by Yihui based on the original architecture.

## ChangeLog

### 2023/10/08

- Released Version 1.0.1
- Added support for RPC/SUB/UNSUB/PUB/DATAGARM
- Note: QoS is not supported

## Currently Supported Platforms

### Architectural Platforms

| Architecture   | Support |
| -------------- | ------- |
| amd64          | YES     |
| x86            | NO      |
| aarch64        | YES     |
| arm            | NO      |

### Operating System Platforms

| System     | Support |
| ---------- | ------- |
| Windows    | YES     |
| MacOS      | YES     |
| Linux      | YES     |
| FreeBSD    | YES     |
| SylixOS    | YES*    |

Note: SylixOS requires the use of a Golang compiler that supports compiling the SylixOS system. For specific information, please contact [AcoInfo Technology Co., Ltd.](https://acoinfo.com)

## Release Mode

Currently, GO-VSOA is released in the form of a local repository. Starting from version V1.1.0, it will be modified to synchronize online and offline releases.

## Simple RPC Example Code

This example assumes that the user has already learned the basics of the Go language and can understand the working logic of go modules. We will write a client/server routine that can control the lighting and read the current lighting status. Other example code can be found in the accompanying "example" folder.

### Preprocessing (Local Offline Mode)

The Golang compiler needs to be set to enable go module mode.

```bash
go env -w GO111MODULE=on
```

Download this SDK to the src/ folder in GOPATH.
Rename the go-vsoa folder to go-vsoa@v1.0.1.
Create two folders, `go-vsoa-server` and `go-vsoa-client`, in the src folder for the client and server, respectively.
Go into each folder and save the following as go.mod.

For the `go-vsoa-server` folder:

```mod  
module go-vsoa-server

go 1.20

require go-vsoa v1.0.1

replace go-vsoa v1.0.1 => ../go-vsoa@v1.0.1
```

For the `go-vsoa-client` folder:

```mod  
module go-vsoa-client

go 1.20

require go-vsoa v1.0.1

replace go-vsoa v1.0.1 => ../go-vsoa@v1.0.1
```

### Writing the Server

File name: `server.go`

```go  
package main

import (
    "encoding/json"
    "go-vsoa/protocol"
    "go-vsoa/server"
    "time"
)

type RpcLightParam struct {
    LightStatus bool `json:"Light On"`
}

var lightstatus = true

func startServer() {
    // Initialize the Go VSOA server. In this example, the server's password is set to "123455".
    // If you don't need a password and have no other requirements, you can leave this part empty and pass it as empty to the server.NewServer function.
    serverOption := server.Option{
        Password: "123456",
    }
    s := server.NewServer("golang VSOA server", serverOption)

    // Register the light URL for the RPC GET method.
    // This allows authorized clients to query the current status of the light.
    handleLightGet := func(req, res *protocol.Message) {
        status, _ := json.Marshal(lightstatus)
        res.Param, _ = json.RawMessage(`{"Light On":` + string(status) + `}`).MarshalJSON()
        res.Data = req.Data
    }
    s.AddRpcHandler("/light", protocol.RpcMethodGet, handleLightGet)

    // Register the light URL for the RPC SET method.
    // This allows authorized clients to control the turning on or off of the light.
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
    s.AddRpcHandler("/light", protocol.RpcMethodSet, handleLightSet)

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

```  

### Writing the Client

File name: `client.go`

```go  
package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "go-vsoa/client"
    "go-vsoa/protocol"
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

    // Query the current status of the light.
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

    // If the light is currently on, turn it off; if it's off, turn it on.
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

    // Query the status of the light after executing the operation.
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
```  

### Running the Example Program

Navigate to the go-vsoa-server and go-vsoa-client folders separately using the command line/terminal.
Execute the go build command and then run the generated executable:

For `go-vsoa-server`:

```bash
cd go-vsoa-server
go build
./go-vsoa-server

# On Windows
.\go-vsoa-server.exe
```  

For `go-vsoa-client`:

```bash
cd go-vsoa-client
go build
./go-vsoa-client

# On Windows
.\go-vsoa-client.exe
```

### Expected Results

The server program does not display any output.
The client program outputs something similar to the following:

```bash
Seq: 1 RPC Get  Light On: true
Seq: 2 RPC Set  Light On: false
Seq: 3 RPC Get  Light On: false
```
