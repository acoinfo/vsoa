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

Create two folders, `vsoa-server` and `vsoa-client`, in the src folder for the client and server, respectively.
Go into each folder and save the following as go.mod.

For the `vsoa-server` folder:

```mod
module vsoa-server

go 1.24

require github.com/acoinfo/vsoa v1.0.5
```

For the `vsoa-client` folder:

```mod
module vsoa-client

go 1.24

require github.com/acoinfo/vsoa v1.0.5
```

### Writing the Server

File name: `server.go`

```go
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

```

### Writing the Client

File name: `client.go`

```go
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
```  

### Running the Example Program

Navigate to the vsoa-server and vsoa-client folders separately using the command line/terminal.
Execute the go build command and then run the generated executable:

For `vsoa-server`:

```bash
cd vsoa-server
go build
./vsoa-server

# On Windows
.\vsoa-server.exe
```  

For `vsoa-client`:

```bash
cd vsoa-client
go build
./vsoa-client

# On Windows
.\vsoa-client.exe
```

### Expected Results

The server program does not display any output.
The client program outputs something similar to the following:

```bash
Seq: 1 RPC Get  Light On: true
Seq: 2 RPC Set  Light On: false
Seq: 3 RPC Get  Light On: false
```
