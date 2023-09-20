package main

import (
	"errors"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/tidwall/gjson"
	"strings"
)

type vmContext struct {
	types.DefaultVMContext
}

type pluginContext struct {
	//embed the default plugin context
	types.DefaultPluginContext

	// headerName and headerValue will be added to response.
	// configured in OnPluginStart
	headerName  string
	headerValue string
}

func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

func (ctx *pluginContext) OnPluginStart(confSize int) types.OnPluginStartStatus {
	proxywasm.LogInfof("OnPluginStart from Go!")
	data, err := proxywasm.GetPluginConfiguration()
	if err != nil && !errors.Is(err, types.ErrorStatusNotFound) {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}

	if err != nil {
		proxywasm.LogCriticalf("error reading plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}

	if !gjson.Valid(string(data)) {
		proxywasm.LogCritical(`invalid configuration format; expected {"header_name": "<header name>", "header_value": "<header value>"}`)
		return types.OnPluginStartStatusFailed
	}

	ctx.headerName = strings.TrimSpace(gjson.Get(string(data), "header_name").Str)
	ctx.headerValue = strings.TrimSpace(gjson.Get(string(data), "header_value").Str)

	if ctx.headerName == "" || ctx.headerValue == "" {
		proxywasm.LogCritical(`invalid configuration format; expected {"header_name": "<header name>", "header_value": "<header value>"}`)
		return types.OnPluginStartStatusFailed
	}

	proxywasm.LogInfof("header from config: %s = %s", ctx.headerName, ctx.headerValue)

	return types.OnPluginStartStatusOK
}

type httpHeaders struct {
	types.DefaultHttpContext

	contextID   uint32
	headerName  string
	headerValue string
}

func (ctx *httpHeaders) OnHttpResponseHeaders(_ int, _ bool) types.Action {

	// Add the header passed by arguments
	if ctx.headerName != "" {
		if err := proxywasm.AddHttpResponseHeader(ctx.headerName, ctx.headerValue); err != nil {
			proxywasm.LogCriticalf("failed to set response headers: %v", err)
		}
	}

	// Get and log the headers
	hs, err := proxywasm.GetHttpResponseHeaders()
	if err != nil {
		proxywasm.LogCriticalf("failed to get response headers: %v", err)
	}

	for _, h := range hs {
		proxywasm.LogInfof("response header <-- %s: %s", h[0], h[1])
	}
	return types.ActionContinue

}

func (ctx *httpHeaders) OnHttpStreamDone() {
	proxywasm.LogInfof("%d finished", ctx.contextID)
}

func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpHeaders{
		contextID:   contextID,
		headerName:  ctx.headerName,
		headerValue: ctx.headerValue,
	}
}

func main() {
	proxywasm.SetVMContext(&vmContext{})
}
