func (h *Hook) handleHook(ctx echo.Context) error {
	err := validateToken(ctx.Request().Header)
	if err != nil {
		h.Logger.Warn("unauthorized access", slog.String("remoteAddr", ctx.Request().RemoteAddr))
		return ctx.NoContent(http.StatusUnauthorized)
	}

	payload, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return fmt.Errorf("failed to read all bytes from body: %w", err)
	}

	go h.publishMessage(payload)

	return nil
}