// handler is passed to the sqs consumer
// if error returned, message will not be deleted from SQS
func (k *Kickstarter) handler(ctx context.Context, message types.Message) error {
	if message.Body == nil {
		k.Logger.Error("received message with empty body")
	}
	userEmail, err := k.extractEmail(*message.Body)
	if err != nil {
		k.Logger.Error("failed to extract email", slog.String("error", err.Error()))
		return nil
	}

	if userEmail == "" {
		k.Logger.Error("email is invalid or empty", slog.String("email", userEmail))
		return nil
	}

	k.Logger.Info("sending mail to", slog.String("Email:", userEmail))
	err = k.sendMail(ctx, userEmail)
	if err != nil {
		k.Logger.Error("failed to send email", slog.String("error", err.Error()))
	}

	err = k.addUserToMSTeamsGroup(ctx, userEmail, k.msTeamsGroupId)
	if err != nil {
		k.Logger.Error("failed to add user to microsoft group", slog.String("error", err.Error()))
	}

	return nil
}

// SendEmail sends an email using AWS SES
func (k *Kickstarter) sendMail(ctx context.Context, recipient string) error {
	EmailSource := k.emailSource

	subject := "Let's Get Started in the TUI Engineering World!"

	firstName, lastName := k.getName(recipient)

	var buf bytes.Buffer

	err := k.emailTemplate.Execute(&buf, templatedata{FirstName: firstName, LastName: lastName})
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	input := &ses.SendEmailInput{
		Destination: &sesTypes.Destination{
			ToAddresses: []string{recipient},
		},
		Message: &sesTypes.Message{
			Body: &sesTypes.Body{
				Html: &sesTypes.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(buf.String()),
				},
			},
			Subject: &sesTypes.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(EmailSource),
	}

	if k.dryRun {
		k.Logger.Info("Dry run mode, skipping sending email")
		k.Logger.Info("Recipient:", slog.String("recipient", recipient))
		k.Logger.Info("Subject", slog.String("subject", subject))
		k.Logger.Info("Source", slog.String("source", EmailSource))
		k.Logger.Debug("Email HTML Body:", slog.String("htmlBody", buf.String()))
		return nil
	}

	_, err = k.sesClient.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	k.Logger.Info("Email sent successfully", slog.String("recipient", recipient))
	k.Logger.Info("Source", slog.String("source", EmailSource))

	return nil
}

func (k *Kickstarter) addUserToMSTeamsGroup(ctx context.Context, email string, groupId string) error {

	userId, err := k.getUserIdByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get userId: %w", err)
	}

	if k.dryRun {
		k.Logger.Info("Dry run mode, skipping adding user to Microsoft Teams group")
		k.Logger.Info("Resolved User ID", slog.String("userId", *userId))
		k.Logger.Info("User Email", slog.String("email", email))
		k.Logger.Info("Microsoft Teams Group ID", slog.String("groupId", groupId))
		return nil
	}

	requestBody := graphmodels.NewReferenceCreate()
	odataId := fmt.Sprintf("https://graph.microsoft.com/v1.0/directoryObjects/{%s}", *userId)
	requestBody.SetOdataId(&odataId)

	err = k.graphClient.Groups().ByGroupId(groupId).Members().Ref().Post(ctx, requestBody, nil)
	if err != nil {
		return fmt.Errorf("failed to add user to microsoft group: %w", err)
	}

	k.Logger.Info("successfully added user to the microsoft teams group", slog.String("user", email))
	return nil
}
