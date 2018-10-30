package utils

import (
	"github.com/Deutsche-Boerse/edt-sftp/conf"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
)

type SftpClient struct {
	Connection *sftp.Client
}

//CreateClient creates sftp connection
func CreateClient() *SftpClient {
	config, _ := conf.NewFactory().Get()
	client, _ := ssh.Dial("tcp", config.Host, &config.SShClientConfig)
	sftpClient, _ := sftp.NewClient(client)
	return &SftpClient{sftpClient}
}

//CreateDirIfNotExists creates remote directory
func (client *SftpClient) CreateDirIfNotExists(dir string) *SftpClient {
	_ = client.Connection.Mkdir(dir)
	return client
}

//RemoveDir for remote directories
func (client *SftpClient) RemoveDir(path string) *SftpClient {
	fileInfoWalker := client.Connection.Walk(path)
	for {
		if processed := !fileInfoWalker.Step(); processed {
			break
		}
		fi := fileInfoWalker.Stat()
		if fi == nil {
			return client
		}
		if fi.IsDir() {
			continue
		}
		_ = client.Connection.Remove(fileInfoWalker.Path())
	}
	client.Connection.Remove(path)
	return client
}

func (client *SftpClient) Close() {
	defer client.Connection.Close()
}

//Exists tests whether remote file exists
func (client *SftpClient) Exists(path string, b *bool) *SftpClient {
	if _, err := client.Connection.Stat(path); err == nil {
		*b = true
		return client
	}
	*b = false
	return client
}

//LinkFromLocalToRemote copy files from remote to local folder
func (client *SftpClient) LinkFromRemoteToLocal(from string, to string) *SftpClient {
	var srcFile *sftp.File
	var dstFile *os.File
	var err error
	if srcFile, err = client.Connection.Open(from); err != nil {
		return client
	}
	if dstFile, err = os.Create(to); err != nil {
		return client
	}
	if _, err = srcFile.WriteTo(dstFile); err != nil {
		return client
	}
	dstFile.Close()
	return client
}

//LinkFromLocalToRemote copy files from local to remote folder
func (client *SftpClient) LinkFromLocalToRemote(from string, to string) *SftpClient {
	var bytes []byte
	var err error
	var f *sftp.File
	if bytes, err = ioutil.ReadFile(from); err != nil {
		log.Println("unable read file " + from + " " + err.Error())
		return client
	}
	if f, err = client.Connection.Create(to); err != nil {
		log.Println("unable create file " + to + " " + err.Error())
		return client
	}
	f.Write(bytes)
	f.Close()
	return client
}
