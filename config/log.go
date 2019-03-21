package config

import (
	"encoding/json"
	"fmt"
	"os"

	sebakcommon "boscoin.io/sebak/lib/common"
	logging "github.com/inconshreveable/log15"
	isatty "github.com/mattn/go-isatty"
	"github.com/spikeekips/cvc"

	"github.com/spikeekips/naru/common"
)

type LogConfig struct {
	cvc.BaseGroup
	File        string `flag-help:"set log file"`
	Format      string `flag-help:"log format"`
	LevelString string `flag:"level" flag-help:"log level {debug error warn info crit}"`
}

func NewLog() *LogConfig {
	return &LogConfig{
		Format:      common.DefaultLogFormat,
		LevelString: common.DefaultLogLevel.String(),
	}
}

func (l LogConfig) ParseLevelString(v string) (string, error) {
	_, err := logging.LvlFromString(v)
	return v, err
}

func formatLogLevel(s string) string {
	switch s {
	case "eror":
		return "error"
	}

	return s
}

func (l LogConfig) ParseFormat(v string) (string, error) {
	switch v {
	case "terminal", "json":
	default:
		return "", fmt.Errorf("invalid log format, '%s'", v)
	}

	return v, nil
}

func (l *LogConfig) FlagValueFormat() string {
	if len(l.Format) < 1 {
		return common.DefaultLogFormat
	}

	return l.Format
}

func (l *LogConfig) FlagValueLevelString() string {
	if len(l.LevelString) < 1 {
		return formatLogLevel(common.DefaultLogLevel.String())
	}

	return l.LevelString
}

func (l *LogConfig) Level() logging.Lvl {
	if len(l.LevelString) < 1 {
		return common.DefaultLogLevel
	}

	lvl, err := logging.LvlFromString(l.LevelString)
	if err != nil {
		log.Warn("invalid log level found", "level", l.LevelString)

		return logging.LvlCrit
	}

	return lvl
}

func (l LogConfig) String() string {
	b, _ := json.Marshal(l)
	return string(b)
}

func (l *LogConfig) Validate() error {
	if len(l.LevelString) < 1 {
		return nil
	}

	if len(l.File) > 0 {
		if _, err := logging.FileHandler(l.File, logging.TerminalFormat()); err != nil {
			return err
		}
	}

	return nil
}

func (l *LogConfig) ValidateFomatter() error {
	switch l.Format {
	case "", "terminal", "json":
		return nil
	}

	return fmt.Errorf("invalid log format, '%s'", l.Format)
}

func (l *LogConfig) Formatter() logging.Format {
	var logFormatter logging.Format
	switch l.Format {
	case "terminal":
		if isatty.IsTerminal(os.Stdout.Fd()) {
			logFormatter = logging.TerminalFormat()
		} else {
			logFormatter = logging.LogfmtFormat()
		}
	case "", "json":
		logFormatter = sebakcommon.JsonFormatEx(false, true)
	}

	return logFormatter
}

func (l *LogConfig) Handler() logging.Handler {
	if l == nil {
		return common.DefaultLogHandler
	}

	var handler logging.Handler
	formatter := l.Formatter()

	if len(l.File) < 1 {
		handler = logging.StreamHandler(os.Stdout, formatter)
	} else {
		handler, _ = logging.FileHandler(l.File, formatter)
	}

	if l.Level() == logging.LvlDebug { // only debug produces `caller` data
		handler = logging.CallerFileHandler(handler)
	}

	return handler
}

func (l *LogConfig) SetLogger(logger logging.Logger) {
	logger.SetHandler(logging.LvlFilterHandler(l.Level(), l.Handler()))
}

func (l *LogConfig) Combine(c *LogConfig) {
	if l == nil {
		l = &LogConfig{}
	}

	file := l.File
	format := l.Format
	levelString := l.LevelString

	if len(file) < 1 {
		l.File = c.File
	}
	if len(format) < 1 {
		l.Format = c.Format
	}
	if len(levelString) < 1 {
		l.LevelString = c.LevelString
	}
}

type Logs struct {
	cvc.BaseGroup
	Global  *LogConfig
	Package *PackageLog
}

type PackageLog struct {
	cvc.BaseGroup
	Config  *LogConfig
	Common  *LogConfig
	Digest  *LogConfig
	Restv1  *LogConfig
	Storage *LogConfig
	SEBAK   *LogConfig
	Query   *LogConfig
}

func NewLogs() *Logs {
	return &Logs{
		Global:  NewLog(),
		Package: &PackageLog{},
	}
}

func (l *Logs) Merge() error {
	if l.Global == nil {
		l.Global = NewLog()
	}

	l.Package.Config.Combine(l.Global)
	l.Package.Common.Combine(l.Global)
	l.Package.Digest.Combine(l.Global)
	l.Package.Restv1.Combine(l.Global)
	l.Package.Storage.Combine(l.Global)
	l.Package.SEBAK.Combine(l.Global)
	l.Package.Query.Combine(l.Global)

	return nil
}
