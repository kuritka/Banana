package main

import (
	"os"
	"os/signal"
	"io/ioutil"

	"github.com/Deutsche-Boerse/edt-sftp/conf"
	"github.com/Deutsche-Boerse/edt-sftp/constants"
	"github.com/Deutsche-Boerse/edt-sftp/host2host"

	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/rs/zerolog/log"
)

var isDownloading = false
var config *conf.SftpConfig

type app struct{}

//Run main function in CRON job
func (app) Run() int {
	con, err := conf.NewFactory().Get()
	if err != nil {
		log.Error().Err(err).Msg("unable to read configuration")
		return constants.ErrorConfiguration
	}
	config = con
	log.Info().Msgf("sftp service started... %s", config.Cron)
	c := cron.New()
	c.AddFunc(config.Cron, downloadJob)
	c.Start()
	c.Run()
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
	return constants.Ok
}

func downloadJob() {
	err := ioutil.WriteFile("alive.txt", []byte("alive\n"), 0644)
	if err != nil {
		log.Error().Err(err).Msg("failed to create working file")
		return
	}
	if isDownloading {
		log.Info().Msg("download process remains...")
		return
	}
	isDownloading = true
	log.Info().Msg("downloading started...")
	if code, err := download(); err != nil {
		log.Error().Err(err).Msg("failed to establish etl client")
		if code != constants.Ok {
			isDownloading = false
			return
		}
	}
	isDownloading = false
}

func download() (int, error) {
	fetched, err := host2host.Download(config)
	if err != nil {
		return constants.ErrorEstablishedConnection, errors.Wrap(err, "failed to establish etl client")
	}
	for _, fetch := range fetched {
		if fetch.Error != nil {
			err = errors.Wrapf(fetch.Error, "error processing file %s \n", fetch.SourcePathOriginal)
		}
	}
	return constants.Ok, err
}

func main() {
	os.Exit(app{}.Run())
}
