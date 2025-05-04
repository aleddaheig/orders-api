package application

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func (app *App) loadMiddleware(middlewares ...Middleware) {
    // Apply middlewares in reverse to maintain execution order
    for i := len(middlewares) - 1; i >= 0; i-- {
        app.router = middlewares[i](app.router)
    }
}
