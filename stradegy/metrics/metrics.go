package metrics

import (
	"context"
	"custom-scheduler/kinds"
	"custom-scheduler/pkg"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"strconv"
	"strings"
	"time"
)

type Metrics struct {
	args     *kinds.Args
	handle   framework.Handle
	pod      *v1.Pod
	nodeInfo *framework.NodeInfo
}

// MetricsConfig {"kind":"NodeMetrics","apiVersion":"metrics.k8s.io/v1beta1","metadata":{"name":"docker-desktop","creationTimestamp":"2022-10-31T10:09:00Z","labels":{"beta.kubernetes.io/arch":"amd64","beta.kubernetes.io/os":"linux","kubernetes.io/arch":"amd64","kubernetes.io/hostname":"docker-desktop","kubernetes.io/os":"linux","node-role.kubernetes.io/control-plane":"","node.kubernetes.io/exclude-from-external-load-balancers":""}},"timestamp":"2022-10-31T10:08:48Z","window":"12.255s","usage":{"cpu":"339779681n","memory":"1776008Ki"}}
type MetricsConfig struct {
	Kinds       string                 `json:"kinds"`
	ApiVersions string                 `json:"apiVersions"`
	MetaData    map[string]interface{} `json:"metadata"`
	TimeStamp   string                 `json:"timestamp"`
	Window      string                 `json:"window"`
	Usage       map[string]string      `json:"usage"`
}

func NewMetricsStradegy(args *kinds.Args, handle framework.Handle, pod *v1.Pod, nodeInfo *framework.NodeInfo) *Metrics {
	return &Metrics{args: args, handle: handle, pod: pod, nodeInfo: nodeInfo}
}

func (m *Metrics) ScheduleByNodeRequests() (*framework.Status, bool) {
	cli, err := pkg.InitClientSet("http://docker.for.mac.host.internal:54030", "")
	if err != nil {
		return framework.NewStatus(framework.Unschedulable, "connect kubernetes cluster failed!"), false
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	requestUri := fmt.Sprintf("/apis/metrics.k8s.io/v1beta1/nodes/%v", m.nodeInfo.Node().Name)
	var mc MetricsConfig
	currentRequest := cli.RESTClient().Get().RequestURI(requestUri).Do(ctx)
	requestRaw, _ := currentRequest.Raw()
	err = json.Unmarshal(requestRaw, &mc)
	if err != nil {
		klog.Error(err)
	}

	// node cpu当前使用量
	currentCpuUsed, _ := strconv.Atoi(strings.Split(mc.Usage["cpu"], "n")[0])
	// node cpu总量
	allocatable := m.nodeInfo.Node().Status.Allocatable.Cpu().Value() * 1000 * 1000000
	// node cpu usage
	nodeCpuUsage, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(currentCpuUsed)/float64(allocatable)), 64)
	klog.Info(nodeCpuUsage)

	// 测试规则
	// node节点cpu使用率(nodeCpuUsage)大于m.args.UsageLimit的话，就不往上调度了
	if nodeCpuUsage > m.args.UsageLimit {
		klog.Errorf("%s cpu usage too high, value: %v", m.nodeInfo.Node().Name, nodeCpuUsage)
		return framework.NewStatus(framework.Unschedulable, "node cpu usage too high!"), false
	}
	return framework.NewStatus(framework.Success, ""), true
}
