package metrics

import (
	"context"
	"custom-scheduler/pkg"
	"encoding/json"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
	"testing"
	"time"
)

type MetricsTest struct {
	Kind        string                 `json:"kind"`
	ApiVersions string                 `json:"apiVersions"`
	MetaData    map[string]interface{} `json:"metadata"`
	TimeStamp   string                 `json:"timestamp"`
	Window      string                 `json:"window"`
	Usage       map[string]string      `json:"usage"`
}

func TestMetrics(t *testing.T) {
	cli, _ := pkg.InitClientSet("http://docker.for.mac.host.internal:54030", "")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	request := cli.RESTClient().Get().RequestURI("/apis/metrics.k8s.io/v1beta1/nodes/docker-desktop").Do(ctx)
	resultByte, _ := request.Raw()
	result := &MetricsTest{}
	err := json.Unmarshal(resultByte, &result)
	if err != nil {
		t.Error(err)
	}

	nodeInfo, err := cli.CoreV1().Nodes().Get(ctx, "docker-desktop", v1.GetOptions{})
	allocatable := nodeInfo.Status.Allocatable.Cpu().Value() * 1000 * 1000000
	t.Log(allocatable)
	cpuUsed, _ := strconv.Atoi(strings.Split(result.Usage["cpu"], "n")[0])
	t.Log(cpuUsed)
	usage := fmt.Sprintf("%.2f", float64(cpuUsed)/float64(allocatable))
	t.Log(usage)
}
