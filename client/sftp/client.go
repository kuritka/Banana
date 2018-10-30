package sftp

import (
	"bytes"
	"github.com/Deutsche-Boerse/edt-sftp/constants"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Deutsche-Boerse/edt-sftp/client/structs"
	"github.com/Deutsche-Boerse/edt-sftp/conf"
	"github.com/Deutsche-Boerse/edt-sftp/response"
	"github.com/Deutsche-Boerse/edt-sftp/unzip"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	ErrDownloadsIsNil = "downloads is nil "
	ErrEmptyZipFile   = "empty zip file "
)

const (
	ident1 = " "
	ident2 = "    + "
	ident3 = "       * "
)

type Config struct {
	Config *conf.SftpConfig
}

func (config *Config) getConnection() (*sftp.Client, error) {
	sshClient, err := ssh.Dial("tcp", config.Config.Host, &config.Config.SShClientConfig)
	if err != nil {
		return nil, err
	}
	// open an SFTP session over an existing ssh connection.
	var connection *sftp.Client
	if connection, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}
	return connection, nil
}

//Download - downloads files from remote to local
func (config *Config) Download() ([]*structs.DownloadInfo, error) {

	var downloads []*structs.DownloadInfo
	var connection *sftp.Client
	var err error

	if connection, err = config.getConnection(); err != nil {
		log.Error().Err(err).Msg("cannot establish connection")
		return []*structs.DownloadInfo{}, err
	}
	defer connection.Close()

	fileInfoWalker := connection.Walk(config.Config.SrcPath)
	for {
		if processed := !fileInfoWalker.Step(); processed {
			break
		}
		var info os.FileInfo
		if info = fileInfoWalker.Stat(); info.IsDir() {
			continue
		}

		//We cannot download file in the middle of uploading so zero len file <filename>_0 must exists
		if ok, err := path.Match(strings.ToLower(config.Config.FileMask), strings.ToLower(info.Name())); err != nil {
			log.Error().Err(err).Msgf("cannot match '%s' with '%s'", strings.ToLower(info.Name()), config.Config.FileMask)
			return downloads, err
		} else if !ok {
			continue
		}
		currentFile := fileInfoWalker.Path()
		// _0 file doesn't exist
		if _, err := connection.Stat(currentFile + config.Config.ZeroLenFileSuffix); err != nil {
			continue
		}

		var downloadInfo structs.DownloadInfo
		downloadInfo, err = config.processDownload(connection, currentFile)
		downloads = append(downloads, &downloadInfo)
		if err != nil {
			downloadInfo.Error = err
			log.Error().Err(err).Msgf("failed downloading %s", currentFile)
			continue
		}
		log.Info().Msgf("%s copied %s", ident1, downloadInfo.SourcePathOriginal)
	}
	return downloads, nil
}

func (config *Config) processDownload(connection *sftp.Client, currentFile string) (structs.DownloadInfo, error) {
	downloadInfo := structs.DownloadInfo{}
	downloadInfo.SourcePathOriginal = currentFile
	var err error
	//Renaming source file. When something breaks, we don't want to repeatedly grab that file
	//instead of that, file stays in the source until issue is resolved
	//the main reason is to prevent loosing files
	downloadInfo.SourcePath = currentFile + constants.EDT
	if err = connection.Rename(currentFile, downloadInfo.SourcePath); err != nil {
		return downloadInfo, errors.Wrapf(err, "cannot rename %s to %s", currentFile, downloadInfo.SourcePath)
	}

	// Copy
	var srcFile *sftp.File
	var dstFile *os.File
	if srcFile, err = connection.Open(downloadInfo.SourcePath); err != nil {
		return downloadInfo, errors.Wrapf(err, "cannot open connection for %s", downloadInfo.SourcePath)
	}
	defer srcFile.Close()

	// Create the destination file
	downloadInfo.DestinationPath = config.Config.DstPath + filepath.Base(currentFile)
	if dstFile, err = os.Create(downloadInfo.DestinationPath); err != nil {
		return downloadInfo, errors.Wrapf(err, "cannot create destination file %s", downloadInfo.DestinationPath)
	}
	defer dstFile.Close()

	// Copy the file
	if _, err = srcFile.WriteTo(dstFile); err != nil {
		return downloadInfo, errors.Wrapf(err, "cannot write %s to destination %s", srcFile.Name(), dstFile.Name())
	}

	//remove zero len file from source
	if err = connection.Remove(downloadInfo.SourcePathOriginal + config.Config.ZeroLenFileSuffix); err != nil {
		return downloadInfo, errors.Wrapf(err, "cannot remove %s ", downloadInfo.SourcePathOriginal+config.Config.ZeroLenFileSuffix)
	}
	return downloadInfo, nil
}

//Unzip source file locally
func (config *Config) Unzip(downloads []*structs.DownloadInfo) error {
	if downloads == nil {
		return errors.New(ErrDownloadsIsNil)
	}
	for _, download := range downloads {
		var unzipped []string
		var err error
		if download.Error != nil {
			continue
		}
		if strings.ToLower(path.Ext(download.DestinationPath)) != constants.ZIP {
			download.Error = errors.New("invalid extension")
			log.Error().Msgf("%s %s", download.Error.Error(), download.DestinationPath)
			continue
		}
		if unzipped, err = unzip.Unzip(download.DestinationPath, config.Config.DstPath); err != nil {
			download.Error = err
			log.Error().Err(err).Msgf("cannot unzip file %s", download.DestinationPath)
			continue
		}
		download.Unzipped = unzipped
		if len(download.Unzipped) == 0 {
			download.Error = errors.New(ErrEmptyZipFile)
			log.Error().Msgf("%s %s", ErrEmptyZipFile, download.DestinationPath)
			continue
		}
		log.Info().Msgf("%s unzipped %s", ident2, path.Base(download.DestinationPath))
		for _, unzippedFile := range unzipped {
			log.Info().Msgf("%s%s", ident3, path.Base(unzippedFile))
		}
	}
	return nil
}

//SendResponses sends response to source
func (config *Config) SendResponses(downloads []*structs.DownloadInfo) error {
	var connection *sftp.Client
	var err error
	if downloads == nil {
		return errors.New(ErrDownloadsIsNil)
	}
	if connection, err = config.getConnection(); err != nil {
		return err
	}
	defer connection.Close()

	for _, download := range downloads {
		if download.Error != nil {
			continue
		}
		resp := response.GetAcknowledge(download.DestinationPath)
		download.ResponsePath = filepath.Join(filepath.Dir(download.SourcePath), resp.Name)

		var remoteResponse *sftp.File
		if remoteResponse, err = connection.Create(download.ResponsePath); err != nil {
			download.Error = err
			log.Error().Msgf("cannot create .response at %s", download.ResponsePath)
			continue
		}

		if _, err := remoteResponse.Write(resp.Content); err != nil {
			download.Error = err
			log.Error().Err(err).Msgf("cannot write response to %s", download.ResponsePath)
			continue
		}

		if err = remoteResponse.Close(); err != nil {
			download.Error = err
			log.Error().Err(err).Msgf("cannot close %s", download.ResponsePath)
			continue
		}

		log.Info().Msgf("%s response %s", ident2, path.Base(download.ResponsePath))
	}
	return nil
}

//SendToEdt sends files to ApiGateway
func (config *Config) SendToEdt(downloads []*structs.DownloadInfo) error {
	timeout := time.Duration(20 * time.Second)
	httpClient := http.Client{Timeout: timeout}
	for _, download := range downloads {
		var resp *http.Response
		var err error

		if download.Error != nil {
			continue
		}
		if resp, err = postMultipart(config.Config.ApiGatewayHost, download.Unzipped, httpClient); err != nil {
			log.Error().Err(err).Msgf("failed to request api-gateway %s", config.Config.ApiGatewayHost)
			return err
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Error().Err(err).Msgf("failed to read response api-gateway %s", config.Config.ApiGatewayHost)
			download.Error = err
			continue
		}
		if resp.StatusCode >= http.StatusBadRequest {
			message := string(body)
			download.Error = errors.New(string(resp.StatusCode))
			log.Error().Msgf("failed to request api-gateway %s", message)
		}
		log.Info().Msgf("%s sent files from %s", ident2, path.Base(download.DestinationPath))
	}
	return nil
}

//Clean removes .zip file from Source an destination. If something breaks it cleans all except .edt file
func (config *Config) Clean(downloads []*structs.DownloadInfo) error {

	var connection *sftp.Client
	var err error
	if downloads == nil {
		return errors.New(ErrDownloadsIsNil)
	}
	if connection, err = config.getConnection(); err != nil {
		return err
	}
	defer connection.Close()
	for _, download := range downloads {

		//whether downloading passed or not we need remove file from source and zip from destination
		if err = os.Remove(download.DestinationPath); err != nil {
			download.Error = err
			log.Error().Err(err).Msgf("cannot remove %s", download.DestinationPath)
			continue
		}
		log.Info().Msgf("%s clean %s", ident1, path.Base(download.DestinationPath))
		for _, unzipped := range download.Unzipped {
			if err = os.Remove(unzipped); err != nil {
				download.Error = err
				log.Error().Err(err).Msgf("cannot remove %s", unzipped)
				break
			}
			log.Info().Msgf("%s clean %s", ident3, path.Base(unzipped))
		}

		//if there are some errors we clean response file (if exists) and skip removing .edt file
		if download.Error != nil {
			if download.ResponsePath != "" {
				if err = connection.Remove(download.ResponsePath); err != nil {
					download.Error = err
					log.Error().Err(err).Msgf("cannot remove %s", download.ResponsePath)
					continue
				}
				log.Info().Msgf("%s clean %s", ident2, path.Base(download.ResponsePath))
			}
			continue
		}

		//and finally remove .edt file
		if err = connection.Remove(download.SourcePath); err != nil {
			download.Error = err
			log.Error().Err(err).Msgf("cannot remove %s", download.SourcePath)
			continue
		}
		log.Info().Msgf("%s clean %s", ident2, path.Base(download.SourcePath))
	}
	return nil
}

// http://polyglot.ninja/golang-making-http-requests/
func postMultipart(url string, files []string, client http.Client) (*http.Response, error) {
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)
	for index, file := range files {
		err := createFormFile(index, file, multipartWriter)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create multipart post request")
		}
	}
	err := multipartWriter.Close()
	if err != nil {
		return nil, err
	}
	return client.Post(url, multipartWriter.FormDataContentType(), &requestBody)
}

func createFormFile(index int, file string, writer *multipart.Writer) error {
	fileWriter, err := writer.CreateFormFile("file_field"+string(index), filepath.Base(file))
	if err != nil {
		return err
	}
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(fileWriter, f)
	if err != nil {
		return err
	}
	return nil
}
