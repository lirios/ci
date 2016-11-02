package service

// Represent the settings file
type Settings struct {
	Server struct {
		Port       string
		DbRootPath string
		OutputPath string
	}
}
