package plugins

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	runtime2 "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

// Name 插件名称
const Name = "custom-plugin"

type Args struct {
	SameAppCount int `json:"same_app_count"`
	WebAppCount  int `json:"web_app_count"`
}

type Custom struct {
	args   *Args
	handle framework.Handle
}

func (c *Custom) Name() string {
	return Name
}

func (c *Custom) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	klog.Infof("prefilter pod: %v", p.Name)
	return nil, framework.NewStatus(framework.Success, "")
}

func (c *Custom) PreFilterExtensions() framework.PreFilterExtensions {
	//TODO implement me
	panic("implement me")
}

func (c *Custom) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	klog.Infof("filter pod: %v, node: %v", pod.Name, nodeInfo.Node().Name)
	
	sameAppCount := 0
	for _, pods := range nodeInfo.Pods {
		if pods.Pod.Labels["app"] == pod.Labels["app"] {
			sameAppCount += 1
		}
	}

	if sameAppCount >= c.args.SameAppCount {
		klog.Errorf("too many same app on node %v，count: %v, no more than %v！", nodeInfo.Node().Name, sameAppCount, c.args.SameAppCount)
		return framework.NewStatus(framework.Unschedulable, "too many same app on node")
	}
	return framework.NewStatus(framework.Success, "")
}

func (c *Custom) PreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	if nodeInfo, err := c.handle.SnapshotSharedLister().NodeInfos().Get(nodeName); err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("prebind get node info error: %+v", nodeName))
	} else {
		klog.Infof("prebind node info: %+v", nodeInfo.Node())
		return framework.NewStatus(framework.Success, "")
	}
}

func New(config runtime.Object, f framework.Handle) (framework.Plugin, error) {
	args := &Args{}
	if err := runtime2.DecodeInto(config, args); err != nil {
		klog.Errorf(err.Error())
	}
	klog.Infof("get plugin config args: %+v", args)
	return &Custom{
		args:   args,
		handle: f,
	}, nil
}
