package kong

import "time"

// JSONResponse representation of a response from mockbin service
type JSONResponse struct {
	StartedDateTime time.Time `json:"startedDateTime"`
	ClientIPAddress string    `json:"clientIPAddress"`
	Method          string    `json:"method"`
	URL             string    `json:"url"`
	HTTPVersion     string    `json:"httpVersion"`
	Cookies         struct {
	} `json:"cookies"`
	Headers struct {
		Host            string `json:"host"`
		Connection      string `json:"connection"`
		AcceptEncoding  string `json:"accept-encoding"`
		XForwardedFor   string `json:"x-forwarded-for"`
		CfRay           string `json:"cf-ray"`
		XForwardedProto string `json:"x-forwarded-proto"`
		CfVisitor       string `json:"cf-visitor"`
		XForwardedHost  string `json:"x-forwarded-host"`
		XForwardedPort  string `json:"x-forwarded-port"`
		XForwardedPath  string `json:"x-forwarded-path"`
		UserAgent       string `json:"user-agent"`
		CfConnectingIP  string `json:"cf-connecting-ip"`
		CdnLoop         string `json:"cdn-loop"`
		XRequestID      string `json:"x-request-id"`
		Via             string `json:"via"`
		ConnectTime     string `json:"connect-time"`
		XRequestStart   string `json:"x-request-start"`
		TotalRouteTime  string `json:"total-route-time"`
	} `json:"headers"`
	QueryString struct {
	} `json:"queryString"`
	PostData struct {
		MimeType string        `json:"mimeType"`
		Text     string        `json:"text"`
		Params   []interface{} `json:"params"`
	} `json:"postData"`
	HeadersSize int `json:"headersSize"`
	BodySize    int `json:"bodySize"`
}
