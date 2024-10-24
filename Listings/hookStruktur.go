type Hook struct {
	*box.Box
	gitlab    *gitlab.Client
	snsClient *sns.Client
}

func New() (*Hook, error) {
	hook := &Hook{
		Box: box.New(
			box.WithWebServer(),
		),
	}

	token, ok := os.LookupEnv("GITLAB_TOKEN")
	if !ok {
		return nil, errors.New("GITLAB_TOKEN must be set")
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	var err error
	hook.gitlab, err = gitlab.NewClient(token, gitlab.WithBaseURL(os.Getenv("GITLAB_URL")), gitlab.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab client %w", err)
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(hook.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	hook.snsClient = sns.NewFromConfig(awsCfg)

	hook.WebServer.POST("/api/v1/hook", hook.handleHook)

	return hook, nil
}

