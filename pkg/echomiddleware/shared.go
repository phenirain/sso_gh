package echomiddleware

import (
	"net/http"
	"strings"
)

var (
	RequestIDHeader    = "X-Request-Id"
	TraceParentHeader  = "Traceparent"
)

//TODO: разобраться с этим
// getting request ID from headers
func getRequestID(headers map[string][]string) string {
	var reqID string
	if reqIDs, ok := headers[http.CanonicalHeaderKey(RequestIDHeader)]; ok && len(reqIDs) > 0 {
		reqID = reqIDs[0]
	}

	return reqID
}

// getting trace ID from headers
func getTraceID(headers map[string][]string) string {
	traceID := "0"
	// example traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-1c8b8f8b8f8b8f8b-01
	// getting 4bf92f3577b34da6a3ce929d0e0e4736
	if traceIDs, ok := headers[http.CanonicalHeaderKey(TraceParentHeader)]; ok && len(traceIDs) > 0 {
		parentVal := traceIDs[0]
		if strings.Count(parentVal, "-") == parentSeparatorNumber {
			traceID = strings.Split(parentVal, "-")[1]
		}
	}

	return traceID
}
