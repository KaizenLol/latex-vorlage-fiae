type Config struct {
	DryRun         bool
	HasAzureCreds  bool
	EmailSource    string
	SqsUrl         string
	MsTeamsGroupId string
}

type Kickstarter struct {
	*box.Box
	sqsConsumer    *consumer.Consumer
	sesClient      *ses.Client
	emailTemplate  *template.Template
	graphClient    *msgraphsdk.GraphServiceClient
	dryRun         bool
	hasAzureCreds  bool
	emailSource    string
	sqsUrl         string
	msTeamsGroupId string
}

func New(config *Config) (*Kickstarter, error) {
	kickstarter := &Kickstarter{
		Box: box.New(
			box.WithWebServer(),
		),
		dryRun:         config.DryRun,
		hasAzureCreds:  config.HasAzureCreds,
		emailSource:    config.EmailSource,
		sqsUrl:         config.SqsUrl,
		msTeamsGroupId: config.MsTeamsGroupId,
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	if err := kickstarter.initSQSConsumer(awsCfg); err != nil {
		return nil, fmt.Errorf("failed to initialize SQS consumer: %w", err)
	}

	kickstarter.sesClient = ses.NewFromConfig(awsCfg)

	if kickstarter.hasAzureCreds {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to load azure credentials %w", err)
		}

		kickstarter.graphClient, err = msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
		if err != nil {
			return nil, fmt.Errorf("failed to create microsoft graph client %w", err)
		}
	}

	kickstarter.emailTemplate, err = template.ParseFiles("assets/email.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %w", err)
	}

	return kickstarter, nil
}