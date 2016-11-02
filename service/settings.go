package service

// Represent the settings file
type Settings struct {
	Server struct {
		URL        string
		Port       string
		DbRootPath string
		OutputPath string
	}
	Slack struct {
		Token   string
		Channel string
	}
}
