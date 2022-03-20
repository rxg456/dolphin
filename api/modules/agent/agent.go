package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	"github.com/prometheus/common/promlog"
	promlogflag "github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/rxg456/dolphin/api/modules/agent/config"
	"github.com/rxg456/dolphin/api/modules/agent/rpc"
	"github.com/rxg456/dolphin/api/modules/agent/taskjob"
)

var (
	// 命令行解析
	app = kingpin.New(filepath.Base(os.Args[0]), "The dolphin-agent")
	// 指定配置文件
	configFile = app.Flag("config.file", "dolphin-agent configuration file path").Short('c').Default("dolphin-agent.yml").String()
)

func main() {
	// 版本信息
	app.Version(version.Print("dolphin-agent"))
	// 帮助信息
	app.HelpFlag.Short('h')

	promlogConfig := promlog.Config{}

	promlogflag.AddFlags(app, &promlogConfig)
	// 强制解析
	kingpin.MustParse(app.Parse(os.Args[1:]))
	// 设置logger
	var logger log.Logger
	logger = func(config *promlog.Config) log.Logger {
		var (
			l  log.Logger
			le level.Option
		)
		if config.Format.String() == "logfmt" {
			l = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		} else {
			l = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		}

		switch config.Level.String() {
		case "debug":
			le = level.AllowDebug()
		case "info":
			le = level.AllowInfo()
		case "warn":
			le = level.AllowWarn()
		case "error":
			le = level.AllowError()
		}
		l = level.NewFilter(l, le)
		l = log.With(l, "ts", log.TimestampFormat(
			func() time.Time { return time.Now().Local() },
			"2006-01-02 15:04:05.000 ",
		), "caller", log.DefaultCaller)
		return l
	}(&promlogConfig)

	level.Debug(logger).Log("debug.msg", "using config.file", "file.path", *configFile)

	sConfig, err := config.LoadFile(*configFile)
	if err != nil {
		level.Error(logger).Log("msg", "config.LoadFile Error,Exiting ...", "error", err)
		return
	}
	level.Info(logger).Log("msg", "load.config.success", "file.path", *configFile, "rpc_server_addr", sConfig.RpcServerAddr)
	// 初始化rpc client
	rpcCli := rpc.InitRpcCli(sConfig.RpcServerAddr, logger)

	// 初始化任务缓存
	taskjob.InitLocals(sConfig.Job.MetaDir)

	// 编排开始
	var g run.Group
	ctxAll, cancelAll := context.WithCancel(context.Background())
	fmt.Println(ctxAll)
	{

		// 处理信号退出的handler
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)
		cancelC := make(chan struct{})
		g.Add(
			func() error {
				select {
				case <-term:
					level.Warn(logger).Log("msg", "Receive SIGTERM ,exiting gracefully....")
					cancelAll()
					return nil
				case <-cancelC:
					level.Warn(logger).Log("msg", "other cancel exiting")
					return nil
				}
			},
			func(err error) {
				close(cancelC)
			},
		)
	}

	{
		// 同步任务
		g.Add(func() error {
			err := rpc.TickerTaskReport(rpcCli, ctxAll)
			if err != nil {
				level.Error(logger).Log("msg", "taskjob.TickerTaskReport.error", "err", err)
				return err
			}
			return err
		}, func(err error) {
			cancelAll()
		})
	}

	g.Run()

}
