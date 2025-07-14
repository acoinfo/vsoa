# GO-VSOA: A Pure GO Implementation of VSOA SDK Framework

GO-VSOA is a development library for VSOA (Vision Service Oriented Architecture) completely implemented in Go and Go assembly, without any dependencies on C libraries and CGO.

[VSOA](https://www.acoinfo.com/product/5330/) is a mission-critical distributed microservices architecture designed by Yihui. This project is the official Go language SDK introduced by Yihui based on the original architecture.

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

Create two folders, `go-vsoa-server` and `go-vsoa-client`, in the src folder for the client and server, respectively.
Go into each folder and save the following as go.mod.

For the `go-vsoa-server` folder:

```mod  
module go-vsoa-server

go 1.20

require github.com/acoinfo/go-vsoa v1.0.5
```

For the `go-vsoa-client` folder:

```mod  
module go-vsoa-client

go 1.20

require github.com/acoinfo/go-vsoa v1.0.5
```

### Writing the Server

File name: `server.go`

```go  
package main

import (
    "encoding/json"
    "time"

    "github.com/acoinfo/go-vsoa/protocol"
    "github.com/acoinfo/go-vsoa/server"
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

```  

### Writing the Client

File name: `client.go`

```go  
package main

import (
    "encoding/json"
    "errors"
    "fmt"

    "github.com/acoinfo/go-vsoa/client"
    "github.com/acoinfo/go-vsoa/protocol"
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
