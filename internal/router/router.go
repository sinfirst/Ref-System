package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/sinfirst/Ref-System/internal/app"
	"github.com/sinfirst/Ref-System/internal/middleware/compress"
	"github.com/sinfirst/Ref-System/internal/middleware/logging"
)

func NewRouter(a *app.App) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logging.WithLogging)
	router.With(compress.DecompressHandle).Post("/api/user/register", a.Register)
	router.With(compress.DecompressHandle).Post("/api/user/login", a.Login)
	router.With(compress.DecompressHandle).Post("/api/user/orders", a.OrdersIn)
	router.With(compress.DecompressHandle).Post("/api/user/balance/withdraw", a.WithDraw)

	router.With(compress.CompressHandle).Get("/api/user/orders", a.OrdersOut)
	router.With(compress.CompressHandle).Get("/api/user/balance", a.GetBalance)
	router.With(compress.CompressHandle).Get("/api/user/withdrawals", a.WithDrawInfo)

	return router
}
