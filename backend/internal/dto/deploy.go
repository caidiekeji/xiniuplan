package dto

// DeployAppUpdateRequest 部署应用更新请求
//
// 与 model.DeployApp 的关键差异：Enabled 用 *bool。
//
// 用普通 bool 时无法区分「前端没传这个字段」和「前端传了 false」——
// 编辑表单不包含 enabled 开关，绑定后零值 false 会被无条件写库，
// 导致用户只是改了个名字就把应用静默禁用了。指针类型让「未传」保持 nil，
// 从而跳过该字段的更新。
type DeployAppUpdateRequest struct {
	Name          string `json:"name"`
	RepoURL       string `json:"repo_url"`
	Branch        string `json:"branch"`
	BuildType     string `json:"build_type"`
	OutDir        string `json:"out_dir"`
	DestPath      string `json:"dest_path"`
	ContainerPort int    `json:"container_port"`
	Enabled       *bool  `json:"enabled"`
	Secret        string `json:"secret"`
}
