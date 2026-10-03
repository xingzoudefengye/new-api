package openai

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/relaykit/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 上游只给流式（例如 ChatGPT 的 Codex 订阅后端），客户端要非流式时，
// 由 OaiResponsesBufferedHandler 把 SSE 聚合成一个完整的 Responses JSON。
func TestOaiResponsesBufferedHandlerAggregatesSSEIntoJSON(t *testing.T) {
	oldMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(oldMode) })

	body := strings.Join([]string{
		`data: {"type":"response.created","response":{"id":"resp_1","object":"response","model":"gpt-test","created_at":1710000000}}`,
		`data: {"type":"response.output_text.delta","delta":"buffered text"}`,
		`data: {"type":"response.completed","response":{"id":"resp_1","object":"response","model":"gpt-test","created_at":1710000000,"status":"completed","usage":{"input_tokens":2,"output_tokens":3,"total_tokens":5}}}`,
		`data: [DONE]`,
		``,
	}, "\n")

	c, recorder, resp, info := newResponsesChatTestContext(t, body, false)
	info.RelayFormat = types.RelayFormatOpenAIResponses

	usage, apiErr := OaiResponsesBufferedHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	require.Equal(t, 5, usage.TotalTokens)

	got := recorder.Body.String()
	require.NotContains(t, got, "data:")
	require.Contains(t, got, `"object":"response"`)
	require.Contains(t, got, "buffered text")

	// 聚合后是 JSON，不能把上游的 text/event-stream 带出去。
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
}

func TestOaiResponsesBufferedHandlerSurfacesUpstreamFailure(t *testing.T) {
	oldMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { gin.SetMode(oldMode) })

	body := strings.Join([]string{
		`data: {"type":"response.failed","response":{"status":"failed","error":{"type":"server_error","message":"upstream exploded"}}}`,
		``,
	}, "\n")

	c, _, resp, info := newResponsesChatTestContext(t, body, false)
	info.RelayFormat = types.RelayFormatOpenAIResponses

	_, apiErr := OaiResponsesBufferedHandler(c, info, resp)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Error(), "upstream exploded")
}
