package host2host

import (
	"github.com/Deutsche-Boerse/edt-sftp/client"
	"github.com/Deutsche-Boerse/edt-sftp/client/structs"
	"github.com/Deutsche-Boerse/edt-sftp/conf"
	"github.com/pkg/errors"
)

// error messages
const (
	errConfig    string = "failed to get config "
	errDownload  string = "failed downloading "
	errUnzip     string = "failed unzipping "
	errResponse  string = "failed sending response "
	errClean     string = "failed cleaning "
	errNilConfig string = "config is nil "
)

//Download files from remote and POST them to endpoint specified by config
func Download(config *conf.SftpConfig) ([]*structs.DownloadInfo, error) {
	var downloads []*structs.DownloadInfo
	if config == nil {
		return downloads, errors.New(errNilConfig)
	}
	c, err := client.NewFactory(client.ClientOptions{SftpConfig: config}).Get()
	if err != nil {
		return nil, errors.New(errConfig + err.Error())
	}
	if downloads, err = c.Download(); err != nil {
		return downloads, errors.New(errDownload + err.Error())
	}
	//nothing has been downloaded
	if downloads == nil {
		return nil, nil
	}
	if err = c.Unzip(downloads); err != nil {
		return downloads, errors.New(errUnzip + err.Error())
	}
	if err = c.SendToEdt(downloads); err != nil {
		return downloads, errors.New(errResponse + err.Error())
	}
	if err = c.SendResponses(downloads); err != nil {
		return downloads, errors.New(errResponse + err.Error())
	}
	if err = c.Clean(downloads); err != nil {
		return downloads, errors.New(errClean + err.Error())
	}
	return downloads, nil
}
