package monitor

const (
	preKey = "rpcsf/monitor/"
)

// targets.json 文件数据
type TargetGroup struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}
