# Custom Scheduler 项目架构文档

> 版本：0.0.1 | Kubernetes：v1.24.1 | 更新日期：2026-04-04

---

## 目录

- [项目简介](#项目简介)
- [目录结构](#目录结构)
- [核心组件](#核心组件)
- [数据流](#数据流)
- [配置说明](#配置说明)
- [部署方式](#部署方式)
- [构建脚本](#构建脚本)
- [已知问题与局限](#已知问题与局限)

---

## 项目简介

`custom-scheduler` 是一个基于 **Kubernetes Scheduler Framework** 的自定义调度器扩展，通过插件机制在标准调度流程中注入自定义业务逻辑，实现更精细化的 Pod 调度策略。

**核心调度目标：**

| 约束 | 说明 |
|------|------|
| CPU 资源限制 | 拒绝将 Pod 调度到 CPU 使用率超过 80% 的节点 |
| 应用分散策略 | 同一应用（相同 `app` label）在单个节点上的副本数不超过 2 |

---

## 目录结构

```
custom-scheduler/
├── main.go                     # 程序入口，注册并启动自定义调度器
├── build.sh                    # 编译 & 构建 Docker 镜像脚本
├── replace.sh                  # 更新 Kubernetes 依赖版本脚本
├── Dockerfile                  # Docker 镜像定义（基于 centos7）
├── go.mod / go.sum             # Go 模块依赖
├── README.md                   # 项目说明
│
├── kinds/
│   └── kinds.go                # 插件配置参数结构体定义
│
├── pkg/
│   ├── kube-client.go          # Kubernetes clientset 初始化
│   └── etcd.go                 # etcd client 初始化
│
├── plugins/
│   └── plugin.go               # 调度插件核心实现（PreFilter/Filter/PreBind）
│
├── stradegy/                   # 调度策略层（注：目录名拼写有误，应为 strategy）
│   ├── metrics/
│   │   ├── metrics.go          # 基于节点 CPU 指标的调度策略
│   │   └── metrics_test.go     # 单元测试
│   └── ipstatic/
│       └── ipstatic.go         # IP 静态分配策略（未实现，预留模块）
│
└── yaml/                       # Kubernetes 部署清单
    ├── configmap.yaml          # 调度器参数配置
    ├── deployment.yaml         # 调度器 Deployment
    ├── clusterRole.yaml        # RBAC ClusterRole
    ├── clusterRoleBind.yaml    # RBAC ClusterRoleBinding
    └── serviceAccount.yaml     # ServiceAccount
```

---

## 核心组件

### 1. 调度插件 `plugins/plugin.go`

实现 Kubernetes Scheduler Framework 的插件接口，插件名称为 `custom-plugin`。

**实现的生命周期钩子：**

```
PreFilter  →  Filter  →  PreBind
```

| 钩子 | 职责 |
|------|------|
| `PreFilter` | 记录 Pod 信息日志，不做拦截 |
| `Filter` | 核心调度逻辑：判断节点是否满足 CPU 和应用分散约束 |
| `PreBind` | 绑定前校验节点信息 |

**Filter 逻辑流程：**

```
获取目标节点信息
    │
    ├─→ 查询节点 CPU 使用率（通过 metrics 策略）
    │       └─→ 超过 usage_limit？→ Unschedulable
    │
    └─→ 统计节点上相同 app label 的 Pod 数量
            └─→ 超过 same_app_count？→ Unschedulable
    
均满足 → Success
```

---

### 2. 指标调度策略 `stradegy/metrics/metrics.go`

负责评估节点的资源负载，决定是否允许调度。

**工作原理：**

1. 调用 Kubernetes Metrics API（`/apis/metrics.k8s.io/v1beta1/nodes/<node>`）获取 CPU 实时用量（纳秒）
2. 从节点 `Allocatable` 中取出 CPU 总量
3. 计算使用率：`usage = cpuUsageNano / (allocatableCPU * 1e9)`
4. 与配置阈值 `usage_limit` 比较

> **注意**：当前 Metrics API 地址硬编码为 `http://docker.for.mac.host.internal:54030`，仅适用于开发环境。

---

### 3. 配置参数 `kinds/kinds.go`

```go
type Args struct {
    SameAppCount int      // 同一 app 在单节点的最大副本数，默认 2
    WebAppCount  int      // Web 应用数量限制，默认 3（当前未使用）
    UsageLimit   float64  // 节点 CPU 使用率阈值，默认 0.8（80%）
    EtcdUrl      []string // etcd 端点地址列表
}
```

---

### 4. 客户端工具 `pkg/`

| 文件 | 说明 |
|------|------|
| `kube-client.go` | 支持集群内（ServiceAccount）和集群外（kubeconfig）两种认证方式 |
| `etcd.go` | 初始化 etcd 客户端，连接超时 5 秒 |

---

## 数据流

```
[Kubernetes API Server]
        │
        │  Pod 等待调度
        ▼
[main.go] ──注册──▶ [custom-plugin]
                          │
                    ┌─────▼─────┐
                    │ PreFilter │  记录日志
                    └─────┬─────┘
                          │
                    ┌─────▼─────┐
                    │  Filter   │
                    └─────┬─────┘
                          │
              ┌───────────┴───────────┐
              │                       │
    ┌─────────▼──────────┐  ┌────────▼────────┐
    │  metrics 策略       │  │  app label 统计  │
    │  查询节点 CPU 使用率 │  │  统计节点副本数   │
    └─────────┬──────────┘  └────────┬────────┘
              │                       │
              └───────────┬───────────┘
                          │
                    ┌─────▼─────┐
                    │  PreBind  │  校验节点信息
                    └─────┬─────┘
                          │
                    [Pod 绑定到节点]
```

**外部依赖：**

- `k8s.io/kubernetes v1.24.1` — Kubernetes 核心
- `k8s.io/kube-scheduler` — 调度器框架
- `k8s.io/client-go` — API 客户端
- `go.etcd.io/etcd/client/v3` — etcd 分布式协调
- `k8s.io/klog/v2` — 结构化日志

---

## 配置说明

### ConfigMap（`yaml/configmap.yaml`）

调度器 Profile 名称：`custom-scheduler`

**启用的插件阶段：**

```yaml
plugins:
  preFilter:
    enabled: [custom-plugin]
  filter:
    enabled: [custom-plugin]
  preBind:
    enabled: [custom-plugin]
```

**插件参数：**

```yaml
pluginConfig:
  - name: custom-plugin
    args:
      web_app_count: 3      # 未使用
      same_app_count: 2     # 同节点最大同类副本数
      usage_limit: 0.8      # CPU 使用率上限
      etcd_url:
        - docker.for.mac.host.internal:2379
```

---

## 部署方式

### RBAC 权限

| 资源 | 所需权限 |
|------|---------|
| `nodes` | get, list, watch |
| `pods` | get, list, watch, update, delete |
| `pods/binding` | create |
| `pods/status` | patch, update |
| `events` | create, update, get |
| `endpoints` | create, update, get |
| `namespaces` | get, list, watch |
| `storageclasses`, `persistentvolumes` | get, list, watch |
| `leases` | get, list, update, create, delete |

### Deployment

```
命名空间：kube-system
副本数：1
镜像：gd306.cn/kubernetes/custom-scheduler:0.0.1
配置挂载：/etc/kubernetes/custom-scheduler-config.yaml
ServiceAccount：custom-scheduler-sa
```

> 调度器启动后支持 **Leader Election**，可扩展为高可用部署。

---

## 构建脚本

### `build.sh`

```bash
# 交叉编译 Linux/AMD64 二进制
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o custom-scheduler main.go

# 构建并推送 Docker 镜像
docker build -t custom-scheduler:0.0.1 .
docker tag custom-scheduler:0.0.1 gd306.cn/kubernetes/custom-scheduler:0.0.1
docker push gd306.cn/kubernetes/custom-scheduler:0.0.1
```

### `replace.sh`

自动拉取指定版本的 `k8s.io` 模块 replace 映射，用于锁定与目标 Kubernetes 版本匹配的依赖。

---

## 已知问题与局限

| 问题 | 说明 |
|------|------|
| 硬编码地址 | `metrics.go` 中 Metrics API 地址硬编码，生产环境需修改 |
| `PreFilterExtensions()` panic | 方法直接 panic，当前不影响功能但存在风险 |
| `ipstatic` 未实现 | IP 静态分配策略模块为空，功能未完成 |
| `web_app_count` 未使用 | 配置项已定义但调度逻辑中未引用 |
| PreFilter 无实际逻辑 | 始终返回成功，无实际过滤作用 |
| 无错误处理 | 部分路径缺少错误处理，Metrics 请求失败时行为不明确 |
| 固定 K8s 版本 | 依赖锁定在 v1.24.1，升级需手动执行 `replace.sh` |

---

## 设计模式总结

| 模式 | 应用位置 |
|------|---------|
| 插件模式 | Scheduler Framework 扩展点 |
| 策略模式 | `stradegy/` 目录下可替换的调度策略 |
| 依赖注入 | 通过 ConfigMap 注入插件参数 |
| 适配器模式 | `pkg/kube-client.go` 统一集群内外认证方式 |
