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


func (h *Hook) publishMessage(payload []byte) {
	var event event

	err := json.Unmarshal(payload, &event)
	if err != nil {
		h.Logger.Error("failed to unmarshal body", slog.String("error", err.Error()))
		return
	}

	h.Logger.Debug("event", slog.Any("event", event))
	messageAttributes, err := h.createMessageAttributes(event)
	if err != nil {
		h.Logger.Error("failed to create message attributes", slog.String("eventName", event.EventName), slog.String("error", err.Error()))
	}

	input := &sns.PublishInput{
		Message:           aws.String(string(payload)),
		TopicArn:          aws.String(os.Getenv("SNS_TOPIC_ARN")),
		MessageAttributes: messageAttributes,
	}

	message, err := h.snsClient.Publish(context.Background(), input)
	if err != nil {
		h.Logger.Error("error publishing to SNS", slog.String("error", err.Error()))
		return
	}

	h.Logger.Info("successfully published message to sns", slog.String("eventName", event.EventName), slog.String("messageId", *message.MessageId))
	h.Logger.Debug("Published message details",
		slog.String("messageId", *message.MessageId),
		slog.String("eventName", event.EventName),
		slog.String("message", string(payload)),
		slog.String("Message Attributes", formatMessageAttributes(messageAttributes)),
	)
}

func (h *Hook) createMessageAttributes(event event) (map[string]types.MessageAttributeValue, error) {
	messageAttributes := make(map[string]types.MessageAttributeValue)

	messageAttributes["eventName"] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(event.EventName),
	}
	switch strings.ToLower(event.EventName) {
	case "user_create":
		user, _, err := h.gitlab.Users.GetUser(event.UserID, gitlab.GetUsersOptions{}, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get gitlab user: %w", err)
		}
		slog.Info("successfully retrieved gitlab user", slog.String("UserName", user.Name))
		messageAttributes["isBot"] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(strconv.FormatBool(user.Bot)),
		}
		return messageAttributes, nil
	default:
		return messageAttributes, nil
	}
}
