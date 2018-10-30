package client

import (
	"errors"
	"strings"

	"github.com/Deutsche-Boerse/edt-sftp/client/sftp"
	"github.com/Deutsche-Boerse/edt-sftp/client/structs"
	"github.com/Deutsche-Boerse/edt-sftp/conf"
)

//Client
type Client interface {
	Download() ([]*structs.DownloadInfo, error)
	Unzip(info []*structs.DownloadInfo) error
	SendResponses(info []*structs.DownloadInfo) error
	SendToEdt(info []*structs.DownloadInfo) error
	Clean(info []*structs.DownloadInfo) error
}

type (
	//ClientFactory returns remote client (SFTP or SCP)
	ClientFactory interface {
		Get() (Client, error)
	}
	//ClientOptions pass parameters for factory
	ClientOptions struct {
		SftpConfig *conf.SftpConfig
	}

	clientImpl struct {
		options ClientOptions
	}
)

//NewFactory creates remote client
func NewFactory(options ClientOptions) ClientFactory {
	return &clientImpl{
		options: options,
	}
}

func (c *clientImpl) Get() (Client, error) {
	if strings.ToLower(c.options.SftpConfig.Type) == "sftp" {
		return &sftp.Config{Config: c.options.SftpConfig}, nil
	}
	return nil, errors.New("not implemented client")
}
