package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopLevelTypical(t *testing.T) {
	assert := assert.New(t)

	tr := Trace{
		Span{TraceID: 1, SpanID: 1, ParentID: 0, Service: "mcnulty", Type: "web"},
		Span{TraceID: 1, SpanID: 2, ParentID: 1, Service: "mcnulty", Type: "sql"},
		Span{TraceID: 1, SpanID: 3, ParentID: 2, Service: "master-db", Type: "sql"},
		Span{TraceID: 1, SpanID: 4, ParentID: 1, Service: "redis", Type: "redis"},
		Span{TraceID: 1, SpanID: 5, ParentID: 1, Service: "mcnulty", Type: ""},
	}

	tr.ComputeTopLevel()

	assert.True(tr[0].TopLevel(), "root span should be top-level")
	assert.False(tr[1].TopLevel(), "main service, and not a root span, not top-level")
	assert.True(tr[2].TopLevel(), "only 1 span for this service, should be top-level")
	assert.True(tr[3].TopLevel(), "only 1 span for this service, should be top-level")
	assert.False(tr[4].TopLevel(), "yet another sup span, not top-level")
}

func TestTopLevelSingle(t *testing.T) {
	assert := assert.New(t)

	tr := Trace{
		Span{TraceID: 1, SpanID: 1, ParentID: 0, Service: "mcnulty", Type: "web"},
	}

	tr.ComputeTopLevel()

	assert.True(tr[0].TopLevel(), "root span should be top-level")
}

func TestTopLevelEmpty(t *testing.T) {
	assert := assert.New(t)

	tr := Trace{}

	tr.ComputeTopLevel()

	assert.Equal(0, len(tr), "trace should still be empty")
}

func TestTopLevelOneService(t *testing.T) {
	assert := assert.New(t)

	tr := Trace{
		Span{TraceID: 1, SpanID: 2, ParentID: 1, Service: "mcnulty", Type: "web"},
		Span{TraceID: 1, SpanID: 3, ParentID: 2, Service: "mcnulty", Type: "web"},
		Span{TraceID: 1, SpanID: 1, ParentID: 0, Service: "mcnulty", Type: "web"},
		Span{TraceID: 1, SpanID: 4, ParentID: 1, Service: "mcnulty", Type: "web"},
		Span{TraceID: 1, SpanID: 5, ParentID: 1, Service: "mcnulty", Type: "web"},
	}

	tr.ComputeTopLevel()

	assert.False(tr[0].TopLevel(), "just a sub-span, not top-level")
	assert.False(tr[1].TopLevel(), "just a sub-span, not top-level")
	assert.True(tr[2].TopLevel(), "root span should be top-level")
	assert.False(tr[3].TopLevel(), "just a sub-span, not top-level")
	assert.False(tr[4].TopLevel(), "just a sub-span, not top-level")
}

func TestTopLevelLocalRoot(t *testing.T) {
	assert := assert.New(t)

	tr := Trace{
		Span{TraceID: 1, SpanID: 1, ParentID: 0, Service: "mcnulty", Type: "web"},
		Span{TraceID: 1, SpanID: 2, ParentID: 1, Service: "mcnulty", Type: "sql"},
		Span{TraceID: 1, SpanID: 3, ParentID: 2, Service: "master-db", Type: "sql"},
		Span{TraceID: 1, SpanID: 4, ParentID: 1, Service: "redis", Type: "redis"},
		Span{TraceID: 1, SpanID: 5, ParentID: 1, Service: "mcnulty", Type: ""},
		Span{TraceID: 1, SpanID: 6, ParentID: 4, Service: "redis", Type: "redis"},
		Span{TraceID: 1, SpanID: 7, ParentID: 4, Service: "redis", Type: "redis"},
	}

	tr.ComputeTopLevel()

	assert.True(tr[0].TopLevel(), "root span should be top-level")
	assert.False(tr[1].TopLevel(), "main service, and not a root span, not top-level")
	assert.True(tr[2].TopLevel(), "only 1 span for this service, should be top-level")
	assert.True(tr[3].TopLevel(), "top-level but not root")
	assert.False(tr[4].TopLevel(), "yet another sup span, not top-level")
	assert.False(tr[5].TopLevel(), "yet another sup span, not top-level")
	assert.False(tr[6].TopLevel(), "yet another sup span, not top-level")
}

func TestTopLevelWithTag(t *testing.T) {
	assert := assert.New(t)

	tr := Trace{
		Span{TraceID: 1, SpanID: 1, ParentID: 0, Service: "mcnulty", Type: "web", Meta: map[string]string{"env": "prod"}},
		Span{TraceID: 1, SpanID: 2, ParentID: 1, Service: "mcnulty", Type: "web", Meta: map[string]string{"env": "prod"}},
	}

	tr.ComputeTopLevel()

	t.Logf("%v\n", tr[1].Meta)

	assert.True(tr[0].TopLevel(), "root span should be top-level")
	assert.Equal("prod", tr[0].Meta["env"], "env tag should still be here")
	assert.False(tr[1].TopLevel(), "not a top-level span")
	assert.Equal("prod", tr[1].Meta["env"], "env tag should still be here")
}

func TestTopLevelGetSetBlackBox(t *testing.T) {
	assert := assert.New(t)

	span := Span{}

	assert.True(span.TopLevel(), "by default, all spans are considered top-level")
	span.setTopLevel(false)
	assert.False(span.TopLevel(), "marked as non top-level (AKA subname)")
	span.setTopLevel(true)
	assert.True(span.TopLevel(), "top-level again")

	span.Meta = map[string]string{"env": "staging"}

	assert.True(span.TopLevel(), "by default, all spans are considered top-level")
	span.setTopLevel(false)
	assert.False(span.TopLevel(), "marked as non top-level (AKA subname)")
	span.setTopLevel(true)
	assert.True(span.TopLevel(), "top-level again")
}

func TestTopLevelGetSetMeta(t *testing.T) {
	assert := assert.New(t)

	span := Span{}

	span.setTopLevel(false)
	assert.Equal("true", span.Meta["_sub_name"], "should have a _sub_name:true flag")
	span.setTopLevel(true)
	assert.Nil(span.Meta, "no meta at all")

	span.Meta = map[string]string{"env": "staging"}

	span.setTopLevel(false)
	assert.Equal("true", span.Meta["_sub_name"], "should have a _sub_name:true flag")
	assert.Equal("staging", span.Meta["env"], "former tags should still be here")
	assert.False(span.TopLevel(), "marked as non top-level (AKA subname)")
	span.setTopLevel(true)
	assert.True(span.TopLevel(), "top-level again")
}

func TestForceMetrics(t *testing.T) {
	assert := assert.New(t)

	span := Span{}

	assert.False(span.ForceMetrics(), "by default, metrics are not enforced for sub name spans")
	span.Meta = map[string]string{"datadog.trace_metrics": "true"}
	assert.True(span.ForceMetrics(), "metrics should be enforced because tag is present")
	span.Meta = map[string]string{"env": "dev"}
	assert.False(span.ForceMetrics(), "there's a tag, but metrics should not be enforced anyway")
}
