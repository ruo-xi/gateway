package middleware

import (
	"net/http"
	"strings"
)

type HandleFunc func(ctx *SliceRouterContext)

type SliceGroup struct {
	handlers []HandleFunc
	Path     string
}

type SliceRouter struct {
	Gourps []*SliceGroup
}

type httpSliceRouterHandler struct {
	sliceRouter *SliceRouter
	coreFunc    func(ctx *SliceRouterContext) http.Handler
}

func (r *SliceRouter) Group(path string) *SliceGroup {
	for _, g := range r.Gourps {
		if g.Path == path {
			return g
		}
	}
	newSliceGroup := &SliceGroup{
		Path: path,
	}
	r.Gourps = append(r.Gourps, newSliceGroup)
	return newSliceGroup
}

func (g *SliceGroup) Use(handleFunc ...HandleFunc) {
	g.handlers = append(g.handlers, handleFunc...)
}

type SliceRouterContext struct {
	*SliceGroup
	Index int
	w     http.ResponseWriter
	r     *http.Request
}

func (c *SliceRouterContext) Next() {
	c.Index++
	if c.Index < len(c.handlers) {
		c.handlers[c.Index](c)
		c.Index++
	}
}

func newSliceRouterContext(w http.ResponseWriter, r *http.Request, router *SliceRouter) *SliceRouterContext {
	var newSliceGroup *SliceGroup
	matchLen := 0
	for _, g := range router.Gourps {
		strings.HasPrefix(r.URL.Path, g.Path)
		if matchLen < len(g.Path) {
			matchLen = len(g.Path)
			newSliceGroup = g
		}
	}

	return &SliceRouterContext{
		SliceGroup: newSliceGroup,
		Index:      -1,
		w:          w,
		r:          r,
	}
}

func (h httpSliceRouterHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	context := newSliceRouterContext(writer, request, h.sliceRouter)
	if h.coreFunc != nil {
		context.handlers = append(context.handlers, func(ctx *SliceRouterContext) {
			h.coreFunc(ctx).ServeHTTP(ctx.w, ctx.r)
		})
	}
	context.Next()
}
