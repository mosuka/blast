package cmd

var (
	configFile         string
	id                 string
	raftAddress        string
	grpcAddress        string
	httpAddress        string
	dataDirectory      string
	peerGrpcAddress    string
	mappingFile        string
	certificateFile    string
	keyFile            string
	commonName         string
	corsAllowedMethods []string
	corsAllowedOrigins []string
	corsAllowedHeaders []string
	file               string
	logLevel           string
	logFile            string
	logMaxSize         int
	logMaxBackups      int
	logMaxAge          int
	logCompress        bool
)
