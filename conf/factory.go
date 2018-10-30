package conf

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/tkanos/gonfig"
)

const envPath = "EDT_SFTP_CONFIG"

type (
	ConfigFactory interface {
		Get() (*SftpConfig, error)
	}

	configFactoryImpl struct {
	}
)

type SftpConfig struct {
	Type              string
	Host              string
	User              string
	PrivateKeyFile    string
	SrcPath           string
	DstPath           string
	FileMask          string
	ZeroLenFileSuffix string
	SShClientConfig   ssh.ClientConfig
	ApiGatewayHost    string
	Cron              string
}

// NewFactory is the Factory Method that returns our implementation
func NewFactory() ConfigFactory {
	return &configFactoryImpl{}
}

//Get reads environment variable `edt_environment`
// and returns correct config name
func (configFactoryImpl) Get() (*SftpConfig, error) {

	envPath, exists := os.LookupEnv(envPath)
	if !exists {
		return nil, errors.New(envPath + " variable must be set")
	}
	var config SftpConfig
	if err := gonfig.GetConf(envPath, &config); err != nil {
		return &config, errors.Wrapf(err, "can not read configuration from %s", envPath)
	}

	pkPath := filepath.Join(path.Dir(envPath), config.PrivateKeyFile)
	buffer, err := ioutil.ReadFile(pkPath)
	if err != nil {
		return &config, errors.Wrapf(err, "can not read private key from %s", pkPath)
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return &config, err
	}
	config.SShClientConfig = ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return &config, nil
}
