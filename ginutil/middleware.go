package ginutil

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/general-backend/goctx"
	"gitlab.com/general-backend/gopkg"
	"gitlab.com/general-backend/gotrace"
	"gitlab.com/general-backend/m800log"
)

func AccessMiddleware(timeout time.Duration, errTimeout gopkg.CodeError) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := goctx.GetContextFromGetHeader(c)
		ctx.Set(goctx.LogKeyHTTPMethod, c.Request.Method)
		ctx.Set(goctx.LogKeyURI, c.Request.URL.RequestURI())

		cid, _ := ctx.GetCID()

		finish := ctx.SetTimeout(timeout)
		start := time.Now().UTC()

		c.Set(goctx.ContextKey, ctx)

		sp, isNew := gotrace.ExtractSpanFromContext(c.HandlerName(), ctx)
		if isNew {
			defer sp.Finish()
			ctx.Set(goctx.LogKeyTrace, sp)
		}

		defer m800log.Access(ctx, start)

		go func() {
			c.Writer.Header().Set(goctx.HTTPHeaderCID, cid)
			c.Next()
			finish()
		}()
		<-ctx.Done()
		switch ctx.Err() {
		// fast return to release http resource, actual handler is still running
		case context.DeadlineExceeded:
			GinError(c, errTimeout)
			return
		default:
			// do nothing, common path
		}
	}
}
