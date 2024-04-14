package echohttp

import (
	"log/slog"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct{}

func NewServer() *Server {
	return new(Server)
}

func (s *Server) EchoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	message := r.URL.Query().Get("message")

	slog.InfoContext(ctx, "got message", slog.Any("request_body", message))
	_, _ = w.Write([]byte(message))
	slog.InfoContext(ctx, "sent message", slog.Any("response_body", message))
}

func NewRouter(s *Server) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/echo", otelhttp.WithRouteTag("/echo", http.HandlerFunc(s.EchoHandler)))
	return otelhttp.NewHandler(
		mux, "echo-server",
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		//otelhttp.WithPropagators(...) could be passed here to use custom propagators, defaults to global propagators
		//otelhttp.WithSpanNameFormatter(...), defaults to otelhttp.DefaultSpanNameFormatter
	)
}
