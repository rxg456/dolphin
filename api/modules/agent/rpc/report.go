package rpc

import (
	"context"
	"time"

	"github.com/go-kit/log/level"
	"github.com/toolkits/pkg/logger"

	"github.com/rxg456/dolphin/api/common"
	"github.com/rxg456/dolphin/api/modules/agent/taskjob"
	serverRpc "github.com/rxg456/dolphin/api/modules/server/rpc"
)

// 定义一个周期执行的ticker
func TickerTaskReport(cli *RpcCli, ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	localIp := common.GetLocalIp()
	cli.DoTaskReport(localIp)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Infof("TickerLogJobSync.receive_quit_signal_and_quit")
			return nil
		case <-ticker.C:
			cli.DoTaskReport(localIp)
		}
	}
}

// 构造TaskReport的rpc请求，其中ReportTasks来自本地map的任务收集任务
func (r *RpcCli) DoTaskReport(localIp string) {
	req := serverRpc.TaskReportRequest{
		AgentIp:     localIp,
		ReportTasks: taskjob.Locals.ReportTasks(),
	}
	var resp serverRpc.TaskReportResponse
	err := r.GetCli()
	if err != nil {
		level.Error(r.logger).Log("msg", "get cli error", "serverAddr", r.ServerAddr, "err", err)
		return
	}

	err = r.Cli.Call("Server.TaskReport", req, &resp)
	if err != nil {
		r.CloseCli()
		level.Error(r.logger).Log("msg", "Server.TaskReport.error", "serverAddr", r.ServerAddr, "err", err)
		return
	}

	// 遍历rpc的结果，分配任务
	if resp.AssignTasks != nil {
		count := len(resp.AssignTasks)
		for i := 0; i < count; i++ {
			at := resp.AssignTasks[i]
			taskjob.Locals.AssignTask(at)
		}
	}
}
