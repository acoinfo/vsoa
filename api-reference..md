# Overview

VSOA is the abbreviation of Vehicle SOA presented by ACOINFO, VSOA provides a reliable, Real-Time SOA (Service Oriented Architecture) framework, this framework has multi-language and multi-environment implementation, developers can use this framework to build a distributed service model.  

VSOA includes the following features:

1. Support resource tagging of unified URL
1. Support URL matching subscribe and publish model
1. Support Real-Time Remote Procedure Call
1. Support parallel multiple command sequences
1. Support reliable and unreliable data publishing and datagram
1. Support multi-channel full-duplex high speed parallel data stream
1. Support network QoS control
1. Easily implement server fault-tolerant design
1. Supports multiple language bindings

VSOA is a dual-channel communication protocol, using both **TCP** and **UDP**, among which the API marked with `quick` uses the **UDP** channel. The quick channel is used for high-frequency data update channels. Due to the high data update frequency, the requirements for communication reliability are not strict. It should be noted that **UDP** channel cannot pass through NAT network, so please do not use quick channel in NAT network.

The total url and payload length of the VSOA data packet cannot exceed **256KBytes - 20Bytes** and **65507Bytes - 20Bytes** on quick channel, so if you need to send a large amount of data, you can use the VSOA data stream.

[Download](https://workdrive.zohopublic.com.cn/external/c537e274b5e4835a55f21c7baf59fee84ae3ebd43ede97aca6606a086142313f)
 this SDK to the `src/` folder in GOPATH.
Unzip `go-vsoa@v1.0.4.zip` to get `go-vsoa@v1.0.4` folder.
Create your work folder in GOPATH for example `go-vsoa-example`.
Go into your work folder and save the following as go.mod.

```mod  
module example.com/go-vsoa-example

go 1.20

require gitee.com/sylixos/go-vsoa v1.0.4
replace gitee.com/sylixos/go-vsoa v1.0.4 => ../go-vsoa@v1.0.4
```

User can use the following code to import the vsoa's sub modules.

> **Module Name should be like folder name or your online repo's URL**

``` golang
import (
    "gitee.com/sylixos/go-vsoa/client"
    "gitee.com/sylixos/go-vsoa/server"
    "gitee.com/sylixos/go-vsoa/position"
    "gitee.com/sylixos/go-vsoa/protocol"
)
```

## Support

The following shows `vsoa` package APIs available.

&nbsp;|Async Method|Callback Method
---|:--:|:--:
server.NewServer||
s.Serve||
s.Close||
s.Count||
s.On|●|●
s.OnDatagram|●|●
s.OnDatagramDefault|●|●
s.Publish|●|●
s.QuickPublish|●|●
s.NewServerStream||
client.NewClient||
c.Connect||
c.Go|●|●
c.Call||●
c.Close||
c.RemoteAddr||
c.Subscribe||
c.UnSubscribe||
c.StartRegulator||
c.StopRegulator||
c.Slot||
c.UnSlot||
c.NewClientStream||
c.IsAuthed||
c.IsClosing||
c.IsShutdown||
position.NewPositionList||
position.NewPosition||
positionList.Add||
positionList.Remove||
positionList.ServePositionListener||
positionList.LookUp||

## VSOA Server package

### NewServer(name string, so Option) \*Server

+ `name` *{string}* Server Name.  
+ `opt` *{Option}* Server option configuration.  
+ Returns: *{\*Server}* VSOA Server struct instance.  

To create a VSOA server, `name` should not be empty string, otherwise it will be set as default string "default GO-VSOA server name".  

If the server requires a connection password, `opt` needs to contain the following member:

+ `Password` *{string}* Connection password. Optional.  

If the server requires TLS encryption to secure the communication connection, `opt` needs to contain the following member:

+ `TLSConfig` *{\*tls.Config}*  Optional.  

> **Example**

``` golang
s := server.NewServer("golang VSOA RPC server", server.Option{})
```

### VSOA Server Struct

+ `Name` *{string}* Used for ServInfo.  

The VSOA Server Struct is responsible for the entire lifecycle of go-vsoa. It's has `On`/`OnDatagram`/`OnDatagramDefault`/`Publish`/`QuickPublish` methods to add processing callbacks for different services.

#### **Close() (err error)**

+ Returns: `err` *{error}* if `err == nil` means success.  

Close VSOA server.

#### **Count() (count int)**

+ Returns: `count` *{int}* Those without successful password verification will also be counted.  

Get the number of connected clients.

#### **Serve(address string) (err error)**

+ `address` *{string}* should be like "IP:Port".  
+ Returns: `err` *{error}* if `err == nil` means success.  

Start the server to serve clients, if the startup fails, returns an error. In general, you can run the server asynchronously in the background using a goroutine.

> **Examples**

``` golang
// blocking process
s.Serve("127.0.0.1:3001")
```

``` golang
// asynchronously
go func() {
    _ = s.Serve("127.0.0.1:3001")
}()
```

#### **On(servicePath string, serviceMethod protocol.RpcMessageType,handler func(\*protocol.Message, \*protocol.Message)) (err error)**

+ `servicePath` *{string}* Request URL.  
+ `serviceMethod` *{protocol.RpcMessageType}* Operation method.  
+ `handler` *{func(\*protocol.Message, \*protocol.Message)}* callback handler.  
+ Returns: `err` *{error}* if `err == nil` means success.  

When a remote client generates an RPC request, the server will receive the corresponding request event, usually the event name is the requested URL matched.

Possible values of `serviceMethod` include `protocol.RpcMethodGet` (`0`) and `protocol.RpcMethodSet` (`1`).

The server can reply to RPC calls through the `handler` function.

> **Example**

``` golang
s := server.NewServer("golang VSOA RPC server", server.Option{})

// Echo payload
s.On("/echo", protocol.RpcMethodGet, func(req, res *protocol.Message) {
    res.Param = req.Param
    res.Data = req.Data
})

// Strictly match '/a/b/c' path
s.On("/a/b/c", protocol.RpcMethodGet, func(req, res *protocol.Message) {
    res.Param = req.Param
    res.Data = req.Data
})

// Match '/a/b/c' and '/a/b/c/...'
s.On("/a/b/c/", protocol.RpcMethodGet, func(req, res *protocol.Message) {
    res.Param = req.Param
    res.Data = req.Data
})

// Default match
s.On("/", protocol.RpcMethodGet, func(req, res *protocol.Message) {
    res.Param = req.Param
    res.Data = req.Data
})

// Delay echo
s.On("/delayecho", protocol.RpcMethodGet, func(req, res *protocol.Message) {
    // This won't block the server main loop
    time.Sleep(1*time.Second)
    res.Param = req.Param
    res.Data = req.Data
})
```

#### **RPC URL match rules**

PATH|RPC match rules
:--|:--
`"/"`|Default URL listener.
`"/a/b/c"`|Only handle `"/a/b/c"` path call.
`"/a/b/c/"`|Handle `"/a/b/c"` and `"/a/b/c/..."` all path calls.

**NOTICE**: If both `"/a/b/c"` and `"/a/b/c/"` RPC handler are present, When the client makes a `"/a/b/c"` RPC call, `"/a/b/c"` handler is matched before `"/a/b/c/"`.

#### **OnDatagram(servicePath string, handler func(\*protocol.Message, \*protocol.Message)) (err error)**

+ `servicePath` *{string}* Request URL.  
+ `handler` *{func(\*protocol.Message, \*protocol.Message)}* callback handler.  
+ Returns: `err` *{error}* if `err == nil` means success.  

The **DATAGRAM** is another transfer type, usually this type data is used to transmit some data that does not require confirmation, for example, VSOA's **DATAGRAM** data packets can be used to build a VPN network.

When a remote client generates an DATAGRAM request, the server will receive the corresponding request event, usually the event name is the requested URL matched.

The server can handle to DATAGRAM calls through the `handler` function.

> **warning：the second param in handler() means nothing here**
>
> **Example**

``` golang
s := server.NewServer("golang VSOA DATAGRAM server", server.Option{})

// read datagram
s.OnDatagram("/datagram", func(req, _ *protocol.Message) {
    fmt.Println(req.Param)
    fmt.Println(req.Data)
})
```

#### **OnDatagramDefault(handler func(\*protocol.Message, \*protocol.Message)) (err error)**

+ `handler` *{func(\*protocol.Message, \*protocol.Message)}* callback handler.  
+ Returns: `err` *{error}* if `err == nil` means success.

When a remote client generates an DATAGRAM request, the server will receive the corresponding request event, when no match URL found even if in wildcard match, do the default callback handler.

The server can handle to DATAGRAM calls dafault through the `handler` function.

> **warning：the second param in handler() means nothing here**
>
> **Example**

``` golang
s := server.NewServer("golang VSOA DATAGRAM server", server.Option{})

// read datagram
s.OnDatagramDefault(func(req, _ *protocol.Message) {
    fmt.Println(req.Param)
    fmt.Println(req.Data)
})
```

#### **Publish(servicePath string, timeDriction any, pubs func(\*protocol.Message, \*protocol.Message)) (err error)**

Publish a message, all clients subscribed to this URL will receive this message. The arguments of this function are as follows:

+ `servicePath` *{string}* publisging URL.  
+ `timeDriction` *{time.Duration|chan struct\{\}}* pusblish interval or manual trigger.  
+ `handler` *{func(\*protocol.Message, \*protocol.Message)}* publish callback handler.  
+ Returns: `err` *{error}* if `err == nil` means success.  

URL matching: URL uses `'/'` as a separator, for example: `'/a/b/c'`, if the client subscribes to `'/a/'`, the server publish `'/a'`, `'/a/b'` or `'/ a/b/c'` message, the client will be received.

> **Example**

``` golang
s := server.NewServer("golang VSOA PUBLISH server", server.Option{})

s.Publish("/publisher", 1*time.Second, func(req, _ *protocol.Message) {
    req.Param, _ = json.RawMessage(`{"publish":"go-vsoa-Publishing"}`).MarshalJSON()
})
```

``` golang
trigger := make(chan struct{}, 100)
i := 1
rawpubs := func(req, _ *protocol.Message) {
    i++
    req.Param, _ = json.RawMessage(`{"publish":"GO-VSOA-RAW-Publishing No. ` + strconv.Itoa(i) + `"}`).MarshalJSON()
}
s.Publish("/raw/publisher", trigger, rawpubs)

go func() {
    _ = s.Serve("127.0.0.1:3001")
}()

go func() {
    for {
        time.Sleep(100 * time.Millisecond)
        if s.TriggerPublisher("/raw/publisher") != nil {
            break
        }
    }
}()
```

#### **QuickPublish(servicePath string, timeDriction time.Duration, pubs func(\*protocol.Message, \*protocol.Message)) (err error)**

QuickPublish is similar to Publish func, but in quick channel.

If a large number of high-frequency publish are required and delivery is not guaranteed, quick publish can be used. But the quick type publish cannot traverse a NAT network, so the quick publish interface is not allowed in a NAT network. quick publish uses a different channel than publish, quick publish does not guarantee the order of arrival.

+ `servicePath` *{string}* publisging URL.  
+ `timeDriction` *{time.Duration}* pusblish interval.  
+ `handler` *{func(\*protocol.Message, \*protocol.Message)}* publish callback handler.  
+ Returns: `err` *{error}* if `err == nil` means success.  

URL matching: URL uses `'/'` as a separator, for example: `'/a/b/c'`, if the client subscribes to `'/a/'`, the server publish `'/a'`, `'/a/b'` or `'/ a/b/c'` message, the client will be received.

> **Example**

``` golang
s := server.NewServer("golang VSOA PUBLISH server", server.Option{})

s.QuickPublish("/publisher/quick", 1*time.Second, func(req, _ *protocol.Message) {
    req.Param, _ = json.RawMessage(`{"publish":"go-vsoa-Publishing in quick channel"}`).MarshalJSON()
})
```

#### **NewServerStream(res \*protocol.Message) (ss \*ServerStream, err error)**

Create a stream to wait for the client stream to connect, this `ServerStream` struct is using when transfer streams.

+ `res` *{\*protocol.Message}* put an empty res for transfer tunnid automatic.  
+ Returns: `ss` *{\*ServerStream}* ServerStream struct instance to start ss.ServeListener.  
+ Returns: `err` *{error}* if `err == nil` means success.  

> **Example**

``` golang
// Create a new server instance
s := server.NewServer("golang VSOA stream server", server.Option{})

// Register URL and handler function
h := func(req, res *protocol.Message) {
    // Create a new server stream
    ss, _ := s.NewServerStream(res)

    // Prepare push and receive buffers
    pushBuf := bytes.NewBufferString("Golang VSOA stream server push Message")
    receiveBuf := bytes.NewBufferString("")

    // Start serving the server stream in a goroutine
    go func() {
      ss.ServeListener(pushBuf, receiveBuf)
      fmt.Println("stream server receiveBuf:", receiveBuf.String())
    }()
}
s.On("/read", protocol.RpcMethodGet, h)
```

## VSOA Client package

### **NewClient(option Option) \*Client**

+ `option` *{Option}* Client option configuration.  
+ Returns: *{\*Client}* VSOA Client struct instance.  

If the server requires a connection password, `opt` needs to contain the following member:

+ `Password` *{string}* Connection password. Optional.  

`option` can also contain the following members:

+ `pingInterval` *{int}* Ping interval time, must be greater than **2** (*time.Second). Optional.  
+ `pingTimeout` *{int}* Ping timeout, must be less than `pingInterval` **default: half of `pingInterval`**. Optional.  
+ `pingLost` *{uint}* How many consecutive ping timeouts will drop the connection. **default: 3**. Optional.  
+ ConnectTimeout *{time.Duration}* timeout for low level connection. **default: 5\*time.Second**. Optional.  

If the server requires TLS encryption to secure the communication connection, `opt` needs to contain the following member:

+ `TLSConfig` *{\*tls.Config}*  Optional.  

``` golang
c := client.NewClient(client.Option{Password: "123456"})
```

### VSOA Client Struct

+ `Conn` *{net.Conn}* Normal RPC/Datagram/Subs/UnSubs goes TCP.  
+ `QConn` *{\*net.UDPConn}* Quick Datagram/Publish goes UDPs.  
+ `SubscribeList` *{map[string]func(m \*protocol.Message)}* used for server publish.  

The VSOA Client Struct is responsible for the entire lifecycle of go-vsoa. It's has `Connect`/`Go`/`Call`/`Subscribe`/`UnSubscribe` methods to processing connections or calls to VSOA server.

#### **Connect(vsoa_or_VSOA_URL, address_or_URL string) (ServerInfo string, err error)**

+ `vsoa_or_VSOA_URL` *{string}* using `VSOA_URL` or not.  
+ `address_or_URL` *{string}* server URL or IP:PORT address.  
+ Returns: ServerInfo *{string}* server name.  
+ Returns: err *{error}* if `err == nil` means success.  

Possible values of `network` include `client.Type_URL` (`VSOA_URL`) and other strings for IPV4 or IPV6 + port address.

> **Examples**

``` golang
// Connect to the VSOA server in IP:PORT mode
servInfo, err := c.Connect("vsoa", "127.0.0.1:3001")
if err != nil {
    fmt.Println(err)
} else {
    fmt.Println(servInfo)
}
```

``` golang
var vsoa_test_server_url = "vsoa://vsoa_test_server"

err := c.SetPosition("127.0.0.1:6001")
if err != nil {
    fmt.Println(err)
    return
}
// Connect to the VSOA server in VSOA_URL mode
servInfo, err := c.Connect(client.Type_URL, vsoa_test_server_url)
if err != nil {
    fmt.Println(err)
} else {
    fmt.Println(servInfo)
}
```

#### **Go(URL string, mt protocol.MessageType, flags any, req \*protocol.Message, reply \*protocol.Message, done chan \*Call) \*Call**

Go method for client is call server asynchronously.

When go a RPC call all params is needed.

+ `URL` *{string}* should be servicePath.  
+ `mt` *{protocol.MessageType}* should be `protocol.TypeRPC`.  
+ `flags` *{any}* should be `protocal.RpcMethodGet` or `protocal.RpcMethodSet`.  
+ `req` *{\*protocol.Message}* should be whole request message.  
+ `reply` *{\*protocol.Message}* should be init for save server reply's message.  
+ Returns: *{\*Call}* for golang to select to finish this call.  

> **Example**

``` golang
req1 := protocol.NewMessage()
req2 := protocol.NewMessage()
reply1 := protocol.NewMessage()
reply2 := protocol.NewMessage()

// Send the first RPC request ("/light") asynchronously and get the channel for waiting the response
Call1 := c.Go("/light", protocol.TypeRPC, protocol.RpcMethodGet, req1, reply1, nil).Done

// Send the second RPC request ("/light") asynchronously and get the channel for waiting the response
Call2 := c.Go("/light", protocol.TypeRPC, protocol.RpcMethodGet, req2, reply2, nil).Done

// Wait for the responses of both RPC calls and log the results using the logAsyncCall function
for i := 0; i < 2; i++ {
    select {
    case call := <-Call1:
        logAsyncCall(call)
    case call := <-Call2:
        logAsyncCall(call)
    }
}

func logAsyncCall(call *client.Call) {
    if call.Error != nil {
        fmt.Println("Error:", call.Error)
        return
    }
    reply := call.Reply
    fmt.Println("Seq:", reply.SeqNo())
}
```

When go a DATAGRAM call all params is needed.

+ `URL` *{string}* should be servicePath.  
+ `mt` *{protocol.MessageType}* should be `protocol.TypeDatagram`.  
+ `flags` *{any}* should be `protocal.ChannelQuick` or `protocal.ChannelNormal`.  
+ `req` *{\*protocol.Message}* should be whole request message.  
+ `reply` *{\*protocol.Message}* should be init for save server reply's message.  
+ Returns: *{\*Call}* for golang to select to finish this call.  

#### **Call(URL string, mt protocol.MessageType, flags any, req \*protocol.Message) (\*protocol.Message, error)**

Call method for client is call server asynchronously.

When calling a RPC call all params is needed.

+ `URL` *{string}* should be servicePath.  
+ `mt` *{protocol.MessageType}* should be `protocol.TypeRPC`.  
+ `flags` *{any}* should be `protocal.RpcMethodGet` or `protocal.RpcMethodSet`.  
+ `req` *{\*protocol.Message}* should be whole request message.  
+ Returns: *{\*protocol.Message}* for server save server reply's message.  
+ Returns: *{error}* if there is no error it should be `nil`.  

> **Example**

``` golang
req := protocol.NewMessage()
req.Param, _ = json.RawMessage(`{"Test Num":123}`).MarshalJSON()
reply, err = c.Call("/a/b/c", protocol.TypeRPC, protocol.RpcMethodGet, req)
if err != nil {
    if err == strErr(protocol.StatusText(protocol.StatusInvalidUrl)) {
        fmt.Println("Pass: Invalid URL")
    } else {
        fmt.Println(err)
    }
} else {
    fmt.Println("Seq:", reply.SeqNo(), "Param:", reply.Param)
}
```

When calling a DATAGRAM call all params is needed.

+ `URL` *{string}* should be servicePath.  
+ `mt` *{protocol.MessageType}* should be `protocol.TypeDatagram`.  
+ `flags` *{any}* should be `protocal.ChannelQuick` or `protocal.ChannelNormal`.  
+ `req` *{\*protocol.Message}* should be whole request message.  
+ Returns: *{\*protocol.Message}* is always empty  .
+ Returns: *{error}* if there is no error it should be `nil`.  

> **Example**

``` golang
req := protocol.NewMessage()
req.Param, _ = json.RawMessage(`{"Test Num":123}`).MarshalJSON()
_, err = c.Call("/datagram", protocol.TypeDatagram, protocol.ChannelNormal, req)
if err != nil {
    fmt.Println(err)
} else {
    fmt.Println("Datagram send done")
}
```

#### **Subscribe(URL string, onPublish func(m \*protocol.Message)) error**

+ `URL` *{string}* should be publishPath.  
+ `onPublish` *{func(\*protocol.Message)}* callback handler.  
+ Returns: *{error}* if there is no error it should be `nil`.  

Subscribe to the specified event (URL), when the server sends the corresponding event, the Client can receive the event.

`url` indicates the event, VSOA event matching is prefix matching, for example, when the client subscribes to the `'/a/b/'` event, then the `'/a/b'` or `'/a/b/c'` event can be received.

> **Example**

``` golang
err = c.Subscribe("/light", func(pubs *protocol.Message) {
    fmt.Println(pubs.Param, pubs.Data)
})
if err != nil {
    fmt.Println(err)
}
```

> **Subscribe URL match rules**

PATH|Subscribe match rules
:--|:--
`"/"`|Catch all publish message.
`"/a/b/c"`|Only catch `"/a/b/c"` publish message.
`"/a/b/c/"`|Catch `"/a/b/c"` and `"/a/b/c/..."` all publish message.

#### **UnSubscribe(URL string) error**

+ `URL` *{string}* should be publishPath.  
+ Returns: *{error}* if there is no error it should be `nil`.  

Unsubscribe the specified event.
If Unsubscribe's URL ends with `/`, it will Unsubscribe all sub URLs.
> **This API calls UnSlot automatically**

PATH|UnSubscribe match rules
:--|:--
`"/"`|Uncatch all publish message.
`"/a/b/c"`|Only Uncatch `"/a/b/c"` publish message.
`"/a/b/c/"`|Uncatch `"/a/b/c"` and `"/a/b/c/..."` all publish message.

#### **StartRegulator(interval time.Duration) error**

+ `interval` *{time.Duration}* should be greater than 1ms.
+ Returns: *{error}* if there is no error it should be `nil`.  

VSOA regulator provides the function of changing the speed of client subscription data. For example, the server publish period is 100ms, and the regulator can slow down the speed to receive once every 1000ms.

One client only allows have one regulator in lifetime. User can change interval by calling `StopRegulator` than `StartRegulator` with new interval.

#### **StopRegulator() error**

+ Returns: *{error}* if there is no error it should be `nil`.  

Stop VSOA regulator. The regulator can be restart by calling `StartRegulator` API again.

#### **Slot(URL string, onPublish func(m *protocol.Message)) error**

+ `URL` *{string}* should be publishPath.  
+ `onPublish` *{func(\*protocol.Message)}* callback handler.  
+ Returns: *{error}* if there is no error it should be `nil`.  

Slot to the specified event (URL), when the server sends the corresponding event, the Client can receive the event in regulator sampling interval.

`url` indicates the event, VSOA event matching is prefix matching, for example, when the client subscribes to the `'/a/b/'` event, then the `'/a/b'` or `'/a/b/c'` event can be received.

#### **UnSlot(URL string) error**

+ `URL` *{string}* should be publishPath.  
+ Returns: *{error}* if there is no error it should be `nil`.  

UnSlot the specified event.
If UnSlot's URL ends with `/`, it will UnSlot all slots' URLs.

PATH|UnSlot match rules
:--|:--
`"/"`|Uncatch all publish message.
`"/a/b/c"`|Only Uncatch `"/a/b/c"` publish message.
`"/a/b/c/"`|Uncatch `"/a/b/c"` and `"/a/b/c/..."` all publish message.

#### **NewClientStream(tunid uint16) (cs \*ClientStream, err error)**

Create a stream to wait to connet the server stream tunnel, this `ClientStream` struct is using when transfer streams.

+ `tunid` *{uint16}* we need to get stream tunnel id form server.  
+ Returns: `cs` *{\*ClientStream}* ClientStream struct instance to start ss.ServeListener.  
+ Returns: `err` *{error}* shoud be `nil` when its success.  

> **Example**

``` golang
req := protocol.NewMessage()

// Send a request to the server using the "/read" method
reply, err := c.Call("/read", protocol.TypeRPC, protocol.RpcMethodGet, req)
if err != nil {
// Check if the error is due to an invalid URL
    if err == errors.New(protocol.StatusText(protocol.StatusInvalidUrl)) {
        fmt.Println("Pass: Invalid URL")
    } else {
        fmt.Println(err)
    }
    return
} else {
    StreamTunID = reply.TunID()
    fmt.Println("Seq:", reply.SeqNo(), "Stream TunID:", StreamTunID)
}

receiveBuf := bytes.NewBufferString("")

// Create a new client stream
cs, err := c.NewClientStream(StreamTunID)
if err != nil {
    fmt.Println(err)
    return
} else {
    go func() {
        buf := make([]byte, 32*1024)
        for {
            n, err := cs.Read(buf)
            if err != nil {
            // EOF means stream closed
            if err == io.EOF {
                break
            } else {
                fmt.Println(err)
            break
            }
        }
        receiveBuf.Write(buf[:n])
        fmt.Println("stream receiveBuf:", receiveBuf.String())

        // Add data to push back to server
        receiveBuf.WriteString(" received & push back to server")
        // Push data back to stream server
        cs.Write(receiveBuf.Bytes())

        // In this example, we just receive little data from server, so we just stop here
        goto STOP
}

STOP:
    cs.StopClientStream()
        streamDone <- 1
    }()
}
// Don't close the stream util the stream goroutine is done
<-streamDone
```

#### **IsAuthed() bool**

+ Returns: *{bool}* when you call Connect methoud without error, it returns `true` otherwise `false`.  

#### **IsClosing() bool**

+ Returns: *{bool}* Check if client is during closing progress.  

#### **IsShutdown() bool**

+ Returns: *{bool}* Check if client is closed.  

## VSOA position package

VSOA Position Server provides the function of querying VSOA server address by service name, similar to DNS server.

### **NewPositionList() \*PositionList**

+ Returns: *{*PositionList}* to hold server url dns.  

### **NewPosition(name string, domain int, ip string, port int, security bool) \*Position**

+ `name` *{string}* VSOA server name. Must set! for position server to lookup.  
+ `domain` *{int}* Address domain. Refer to `tcp` or `socket` module.  
+ `ip` *{string}* Server IP Address.  
+ `port` *{int}* Server Port.  
+ `Security` *{bool}* Server using TLS or not. **default: false**  

### VSOA PositionList struct

#### **Add(p Position)**

+ `p` *{Position}* the server position you want to add to position server.  

Add adds a Position to the PositionList.
It updates the PositionList if the Position already exists.
If Position.IP is not a valid IP address, it does nothing.

#### **Remove(p Position)**

+ `name` *{string}* the name of the server position you want to remove from position server.  

Remove removes an element from the PositionList based on the provided name.  

#### **ServePositionListener(address net.UDPAddr) (err error)**

+ `address` *{net.UDPAddr}* position address.  
+ Returns: `err` {error}* if `err == nil` means success.  

> **Example**

``` golang
// startPosition initializes a new position list and adds a new position to it.
// It then starts a position listener in a separate goroutine.
func startPosition() {
    // Create a new position list
    pl := position.NewPositionList()

    // Add a new position to the list
    pl.Add(*position.NewPosition("vsoa_test_server", 1, "127.0.0.1", 3001, false))

    // Start a position listener in a separate goroutine
    go pl.ServePositionListener(net.UDPAddr{
        IP:   net.ParseIP("127.0.0.1"),
        Port: 6001,
    })
}
```

#### **LookUp(name string, position_addr string, timeout time.Duration) (err error)**

+ `name` *{string}* the name to look up.  
+ `position_addr` *{string}* the address of the position server.  
+ `timeout` *{time.Duration}* the duration to wait for the lookup operation to complete or timeout.  
+ Returns: `err` {error}* if `err == nil` means success.  

When you call Connect in the Client, it will automatically invoke the LookUp method here. Of course, you can also call it manually for other purposes, such as checking if the IP of a known server name has changed.

> **Example**

``` golang
p := new(position.Position)
err := p.LookUp(address_or_URL, client.position, 500*time.Millisecond)
if err != nil {
    fmt.Println(err)
} else {
    fmt.Println(p)
}
```

## VSOA protocol package

This package has all methods you need about VSOA protocol.

### **NewMessage() \*Message**

NewMessage returns a new instance of the Message struct.

+ Returns: *{\*Message}* new instance of the Message struct.  

> **Example**

``` golang
req := protocol.NewMessage()
```

### **TypeText(code MessageType) string**

+ `code` *{MessageType}* should be like above.  
+ Returns: *{string}* code in string like `TYPE_SERVINFO`.  

MassageType|Value|Returns
---|:--:|---
`protocol.TypeServInfo`|0|TYPE_SERVINFO
`protocol.TypeRPC`|1|TYPE_RPC
`protocol.TypeSubscribe`|2|TYPE_SUBSCRIBE
`protocol.TypeUnsubscribe`|3|TYPE_UNSUBSCRIBE
`protocol.TypePublish`|4|TYPE_PUBLISH
`protocol.TypeDatagram`|5|TYPE_DATAGRAM
`protocol.TypeQosSetup`|6|TYPE_QOS_SETUP
`protocol.TypePingEcho`|0xff|TYPE_PING

### **StatusText(code StatusType) string**

+ `code` *{StatusType}* should be like above.  
+ Returns: *{string}* code in string like `Password error`.  

StatusType|Value|Returns
---|:--:|---
`protocol.StatusSuccess`|0|Call succeeded
`protocol.StatusPassword`|1|Wrong password
`protocol.StatusArguments`|2|Parameter error
`protocol.StatusInvalidUrl`|3|Invalid URL
`protocol.StatusNoResponding`|4|Server not responding
`protocol.StatusNoPermissions`|5|No permission
`protocol.StatusNoMemory`|6|Out of memory

You can also define your own status code. The user-defined failure value is recommended to be `128` ~ `254`.

### VSOA Header Struct

#### **MessageType() MessageType**

+ Returns: *{MessageType}* this Header's MessageType.

MassageType|Value|Description
---|:--:|---
`protocol.TypeServInfo`|0|Shack hand between C/S
`protocol.TypeRPC`|1|VSOA RPC call
`protocol.TypeSubscribe`|2|VSOA subscribe
`protocol.TypeUnsubscribe`|3|VSOA cannel subscribe
`protocol.TypePublish`|4|VSOA server Publish data to subscriber
`protocol.TypeDatagram`|5|VSOA Datagram without resp
`protocol.TypeQosSetup`|6|Setup Qos for VSOA
`protocol.TypePingEcho`|0xff|VSOA internel ping call

#### **MessageTypeText() string**

+ Returns: *{string}* the text representation of the message type.  

#### **Version() byte**

+ Returns: *{byte}* version of the message.  

#### **MessageRpcMethod() RpcMessageTyp**

+ Returns: *{RpcMessageTyp}* the RPC method, returns 0xEE if it's not an RPC message of VSOA.  

#### **MessageRpcMethodText() string**

+ Returns: *{string}* the text representation of the RPC method.  

#### **StatusType() StatusType**

+ Returns: *{StatusType}* the status type of the message.  

Status Code|Value|Description
---|:--:|---
`protocol.StatusSuccess`|0|Call succeeded
`protocol.StatusPassword`|1|Wrong password
`protocol.StatusArguments`|2|Parameter error
`protocol.StatusInvalidUrl`|3|Invalid URL
`protocol.StatusNoResponding`|4|Server not responding
`protocol.StatusNoPermissions`|5|No permission
`protocol.StatusNoMemory`|6|Out of memory

You can also define your own status code. The user-defined failure value is recommended to be `128` ~ `254`.

#### **StatusTypeText() string**

+ Returns: *{string}* the text representation of the status type.  

#### **SeqNo() uint32**

+ Returns: *{uint32}* the sequence number of the message.  

For each RPC, this is a very important information to ensure the integrity of RPC transaction, because the RPC can be invoked in parallel asynchronously.

#### **TunID() uint16**

+ Returns: *{uint16}*  port number.  

the tunnel ID of the active stream tunnel port or Client quick channel port.

#### **IsOneway() bool**

+ Returns: *{bool}* whether the message is a one-way(`DATAGRAM`/`PUBLISH`) message.  

#### **IsRPC() bool**

+ Returns: *{bool}* whether the message is an RPC message.  

#### **IsReply() bool**

+ Returns: *{bool}* whether the message is a reply message.  

#### **IsPingEcho() bool**

+ Returns: *{bool}* whether the message is a Ping Echo message.  

#### **IsServInfo() bool**

+ Returns: *{bool}*  whether the message is a service info message.  

> User will not using it.

#### **IsSubscribe() bool**

+ Returns: *{bool}* whether the message is a subscribe message.  

#### **IsUnSubscribe() bool**

+ Returns: *{bool}* whether the message is an unsubscribe message.  

#### **IsValidTunid() bool**

+ Returns: *{bool}* whether the Header is a valid Tunid.  

#### **SetMessageType(mt MessageType)**

+ `mt` *{MessageType}* the VSOA message type of the message.  

#### **SetMessageRpcMethod(t RpcMessageType)**

+ `t` *{RpcMessageType}* the VSOA RPC method of the message.  

RpcMessageType|Value|Description
---|:--:|---
`protocol.RpcMethodGet`|0|Get method **default**
`protocol.RpcMethodSet`|1|Set method

#### **SetReply(r bool)**

+ `r` *{bool}* sets the reply flag.  

> User will not using it.

#### **SetSeqNo(seq uint32)**

+ `seq` *{uint32}* sets the sequence number.  

> User will not using it.

#### **SetTunId(ti uint16)**

+ `ti` *{uint16}* sets the tunnel ID of the VSOA client.  

> User will not using it.

#### **SetValidTunid()**

sets a valid Tunid.

> User will not using it.

#### **SetPingEcho()**

sets the type flag to PingEcho.

> User will not using it.

### VSOA Message Struct

+ `_`     *{\*Header}* To have Header struct methods.  
+ `URL`   *{[]byte}* We call it URL but it more likely to be the server PATH for rpcx users.  
+ `Param` *{json.RawMessage}* It's []byte but have Marshal & UnMarshal method.  
+ `Data`  *{[]byte}*  Raw data in VSOA message.  

The VSOA Message Struct is holding Header's method & has all information like URL, Param, Data.

Param should be a Marshaled json.

#### **Clone() \*Message**

Clone clones from an message.

+ Returns: *{\*Message}* new instance of the Message struct cloned the origin one.  

> **Example**

``` golang
req := protocol.NewMessage()
reqClone := req.Clone()
```

#### **Reset()**

Reset clean data of this message but keep allocated data.

> **Example**

``` golang
req := protocol.NewMessage()
req.Reset()
```
