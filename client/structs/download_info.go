package structs

type DownloadInfo struct {
	DestinationPath    string
	SourcePath         string
	SourcePathOriginal string
	Unzipped           []string
	ResponsePath       string
	Error              error
}
