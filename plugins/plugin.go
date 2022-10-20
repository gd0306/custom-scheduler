package plugins

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

// 插件名称
const Name = "custom-scheduler"

var (
	_ framework.PreFilterPlugin = &Custom{}
	_ framework.FilterPlugin    = &Custom{}
	_ framework.PreBindPlugin   = &Custom{}
)

type Args struct {
	FavoriteColor  string `json:"favorite_color,omitempty"`
	FavoriteNumber int    `json:"favorite_number,omitempty"`
	ThanksTo       string `json:"thanks_to,omitempty"`
}

type Custom struct {
	args   *Args
	handle framework.Handle
}

func (c *Custom) Name() string {
	return Name
}

func (c *Custom) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	klog.V(3).Infof("prefilter pod: %v", p.Name)
	return nil, framework.NewStatus(framework.Success, "")
}

func (c *Custom) PreFilterExtensions() framework.PreFilterExtensions {
	//TODO implement me
	panic("implement me")
}

func (c *Custom) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	klog.V(3).Infof("filter pod: %v, node: %v", pod.Name, nodeInfo.Node().Name)
	return framework.NewStatus(framework.Success, "")
}

//func (c *Custom) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
//	klog.V(3).Infof("prefilter pod: %v", pod.Name)
//	return framework.NewStatus(framework.Success, "")
//}

//func (c *Custom) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
//	klog.V(3).Infof("filter pod: %v, node: %v", pod.Name, nodeName)
//	return framework.NewStatus(framework.Success, "")
//}

func (c *Custom) PreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	if nodeInfo, err := c.handle.SnapshotSharedLister().NodeInfos().Get(nodeName); err != nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("prebind get node info error: %+v", nodeName))
	} else {
		klog.V(3).Infof("prebind node info: %+v", nodeInfo.Node())
		return framework.NewStatus(framework.Success, "")
	}
}

//type PluginFactory = func(configuration *runtime.Unknown, f FrameworkHandle) (Plugin, error)
func New(_ runtime.Object, f framework.Handle) (framework.Plugin, error) {
	args := &Args{}
	klog.V(3).Infof("get plugin config args: %+v", args)
	return &Custom{
		args:   args,
		handle: f,
	}, nil
}
