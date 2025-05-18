package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/Ref-System/internal/app"
	"github.com/sinfirst/Ref-System/internal/middleware/auth"
	"github.com/sinfirst/Ref-System/internal/middleware/compress"
	"github.com/sinfirst/Ref-System/internal/middleware/logging"
)

func NewRouter(a *app.App, logger logging.Logger) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logger.WithLogging)
	router.With(compress.DecompressHandle).Post("/api/user/register", a.Register)
	router.With(compress.DecompressHandle).Post("/api/user/login", a.Login)
	router.With(compress.DecompressHandle, auth.AuthMiddleware).Post("/api/user/orders", a.OrdersIn)
	router.With(compress.DecompressHandle, auth.AuthMiddleware).Post("/api/user/balance/withdraw", a.WithDraw)

	router.With(compress.CompressHandle, auth.AuthMiddleware).Get("/api/user/orders", a.OrdersInfo)
	router.With(compress.CompressHandle, auth.AuthMiddleware).Get("/api/user/balance", a.GetBalance)
	router.With(compress.CompressHandle, auth.AuthMiddleware).Get("/api/user/withdrawals", a.WithDrawInfo)

	return router
}
