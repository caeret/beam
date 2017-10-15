# Beam

Beam is a server-side library which supports the redis protocol.

# Usage

Simple echo server:

```
handler := beam.HandleFunc(func(request *beam.Request) (beam.Reply, error) {
    if request.CommandStr() == "echo" && request.Len() == 1 {
        return beam.NewSimpleStringsReply(request.ArgStr(0)), nil
    }
    return beam.NewErrorsReply("invalid command."), nil
})
config := beam.Config{
    Logger: logging.NewSimpleLogger(),
    Addr:   ":8090",
}
server := beam.NewServer(handler, config)
panic(server.Serve())
```

Simple key-value storage supports SET, GET and DEL commands:

```
var storage sync.Map
const passwd = "foobar"

// create a MappedHandler
mappedHandler := beam.NewMappedHandler()
// add GET handler
mappedHandler.SetFunc("GET", func(request *beam.Request) (beam.Reply, error) {
    if request.Len() != 1 {
        return beam.NewErrorsReply("ERR invalid command."), nil
    }
    v, ok := storage.Load(request.ArgStr(0))
    if !ok {
        return beam.NewBulkStringsReplyRaw(nil), nil
    }
    return beam.NewBulkStringsReplyRaw(v.([]byte)), nil
})
// add SET handler
mappedHandler.SetFunc("SET", func(request *beam.Request) (beam.Reply, error) {
    if request.Len() != 2 {
        return beam.NewErrorsReply("ERR invalid command."), nil
    }
    storage.Store(request.ArgStr(0), request.Arg(1))
    return beam.NewSimpleStringsReply("OK"), nil
})
// add DEL handler
mappedHandler.SetFunc("DEL", func(request *beam.Request) (beam.Reply, error) {
    if request.Len() != 1 {
        return beam.NewErrorsReply("ERR invalid command."), nil
    }
    _, exist := storage.Load(request.ArgStr(0))
    if !exist {
        return beam.NewIntegersReply(0), nil
    }
    storage.Delete(request.ArgStr(0))
    return beam.NewIntegersReply(1), nil
})
// add authetication handler
mappedHandler.SetFunc("AUTH", func(request *beam.Request) (beam.Reply, error) {
    // password check
    if request.Len() == 1 && request.ArgStr(0) == passwd {
        request.SetAttr("auth", struct{}{})
        return beam.NewSimpleStringsReply("OK"), nil
    }
    return beam.NewErrorsReply("AUTH invalid password."), nil
})

// create a handler chain
handler := beam.NewHandlerChain(mappedHandler)
// add authetication middleware
handler.AddFunc(func(req *beam.Request, next beam.Handler) (beam.Reply, error) {
    fmt.Println(req)
    if !req.HasAttr("auth") && strings.ToUpper(req.CommandStr()) != "AUTH" {
        return beam.NewErrorsReply("NOAUTH Authentication required."), nil
    }
    return next.Handle(req)
})

s := beam.NewServer(handler, beam.Config{
    Logger:    logging.NewSimpleLogger(),
    RWTimeout: time.Second * 5,
    Addr:      ":6390",
})

fmt.Println("serve:", s.Serve())
```