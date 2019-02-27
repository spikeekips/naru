package config

import (
	"fmt"
	"os"

	sebakcommon "boscoin.io/sebak/lib/common"
	logging "github.com/inconshreveable/log15"
	isatty "github.com/mattn/go-isatty"
	"github.com/spikeekips/cvc"

	restv1 "github.com/spikeekips/naru/api/rest/v1"
	"github.com/spikeekips/naru/common"
	"github.com/spikeekips/naru/digest"
	"github.com/spikeekips/naru/sebak"
	"github.com/spikeekips/naru/storage"
)

type Log struct {
	cvc.BaseGroup
	File   string      `flag-help:"set log file"`
	Format string      `flag-help:"log format"`
	Level  logging.Lvl `flag-help:"log level {debug error warn info crit}"`
}

func NewLog() *Log {
	return &Log{
		Format: common.DefaultLogFormat,
		Level:  common.DefaultLogLevel,
	}
}

func (l *Log) FlagValueLevel() string {
	s := l.Level.String()
	switch l.Level {
	case logging.LvlError:
		s = "error"
	case logging.LvlDebug:
		s = "debug"
	}

	return s
}

func (l Log) ParseLevel(v string) (logging.Lvl, error) {
	return logging.LvlFromString(v)
}

func (l Log) ParseFormat(v string) (string, error) {
	switch v {
	case "terminal":
	case "json":
	default:
		return "", fmt.Errorf("invalid log format, '%s'", v)
	}

	return v, nil
}

func (l *Log) Formatter() (logging.Format, error) {
	var logFormatter logging.Format
	switch l.Format {
	case "terminal":
		if isatty.IsTerminal(os.Stdout.Fd()) {
			logFormatter = logging.TerminalFormat()
		} else {
			logFormatter = logging.LogfmtFormat()
		}
	case "json":
		logFormatter = sebakcommon.JsonFormatEx(false, true)
	default:
		err := fmt.Errorf("invalid log format, '%s'", l.Format)
		return nil, err
	}

	return logFormatter, nil
}

func (l *Log) Handler() (logging.Handler, error) {
	var handler logging.Handler
	formatter, err := l.Formatter()
	if err != nil {
		return nil, err
	}

	if len(l.File) < 1 {
		handler = logging.StreamHandler(os.Stdout, formatter)
	} else {
		lh, err := logging.FileHandler(l.File, formatter)
		if err != nil {
			return nil, err
		}
		handler = lh
	}

	if l.Level == logging.LvlDebug { // only debug produces `caller` data
		handler = logging.CallerFileHandler(handler)
	}

	return handler, nil
}

func (l *Log) SetLogging(f func(logging.Lvl, logging.Handler)) error {
	handler, err := l.Handler()
	if err != nil {
		return err
	}

	f(l.Level, handler)

	return nil
}

func (l *Log) SetLoggingWithDefault(f func(logging.Lvl, logging.Handler), lvl logging.Lvl, handler logging.Handler) error {
	if l == nil {
		f(lvl, handler)
		return nil
	}
	if _, err := l.Formatter(); err != nil {
		f(lvl, handler)
		return nil
	}

	return l.SetLogging(f)
}

type Logs struct {
	cvc.BaseGroup
	Global  *Log
	Package *PackageLog
}

func (l *Logs) SetAllLogging(logger logging.Logger) error {
	lvl := l.Global.Level
	handler, err := l.Global.Handler()
	if err != nil {
		return err
	}

	common.SetLoggingWithLogger(lvl, handler, logger)

	if err := l.Package.Config.SetLoggingWithDefault(SetLogging, lvl, handler); err != nil {
		return err
	}
	if err := l.Package.Common.SetLoggingWithDefault(common.SetLogging, lvl, handler); err != nil {
		return err
	}
	if err := l.Package.Digest.SetLoggingWithDefault(digest.SetLogging, lvl, handler); err != nil {
		return err
	}
	if err := l.Package.Restv1.SetLoggingWithDefault(restv1.SetLogging, lvl, handler); err != nil {
		return err
	}
	if err := l.Package.SEBAK.SetLoggingWithDefault(sebak.SetLogging, lvl, handler); err != nil {
		return err
	}
	if err := l.Package.Storage.SetLoggingWithDefault(storage.SetLogging, lvl, handler); err != nil {
		return err
	}

	return nil
}

type PackageLog struct {
	cvc.BaseGroup
	Config  *Log
	Common  *Log
	Digest  *Log
	Restv1  *Log
	Storage *Log
	SEBAK   *Log
}

func NewLogs() *Logs {
	return &Logs{
		Global:  NewLog(),
		Package: &PackageLog{},
	}
}
