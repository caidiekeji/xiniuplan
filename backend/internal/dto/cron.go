package dto

// CreateCronRequest 创建计划任务请求
type CreateCronRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required,oneof=shell backup url app_restart"`
	Schedule string `json:"schedule" binding:"required"`
	Command  string `json:"command"`
	Target   string `json:"target"`
	Remark   string `json:"remark"`
}

// UpdateCronRequest 更新计划任务请求（所有字段可选）
type UpdateCronRequest struct {
	Name     *string `json:"name"`
	Type     *string `json:"type"`
	Schedule *string `json:"schedule"`
	Command  *string `json:"command"`
	Target   *string `json:"target"`
	Remark   *string `json:"remark"`
}

// CronTaskResponse 计划任务响应
type CronTaskResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Schedule    string `json:"schedule"`
	Command     string `json:"command"`
	Target      string `json:"target"`
	Remark      string `json:"remark"`
	Status      string `json:"status"`
	LastRunAt   string `json:"last_run_at"`
	NextRunAt   string `json:"next_run_at"`
	NextRunTime string `json:"next_run_time"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CronLogResponse 执行日志响应
type CronLogResponse struct {
	ID        uint   `json:"id"`
	TaskID    uint   `json:"task_id"`
	TaskName  string `json:"task_name"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Output    string `json:"output"`
	Result    string `json:"result"`
	Duration  string `json:"duration"`
	CreatedAt string `json:"created_at"`
}
