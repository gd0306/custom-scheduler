package kinds

type Args struct {
	SameAppCount int      `json:"same_app_count"`
	WebAppCount  int      `json:"web_app_count"`
	UsageLimit   float64  `json:"usage_limit"`
	EtcdUrl      []string `json:"etcd_url"`
}
