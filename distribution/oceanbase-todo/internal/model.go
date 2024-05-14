package internal

import (
	"time"

	"gorm.io/gorm"
)

type Todo struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"deletedAt" gorm:"index"`
	Title       string         `json:"title" binding:"required"`
	Description string         `json:"description"`
	FinishedAt  gorm.DeletedAt `json:"finishedAt"`
}

type EditTodo struct {
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	FinishedAt      *time.Time `json:"finishedAt"`
	ClearFinishedAt bool       `json:"clearFinishedAt"`
}

var InitialTodos = []Todo{
	{
		Title:       "Quick Start",
		Description: "Deploy a single-node OceanBase database for testing. <a href=\"https://oceanbase.github.io/ob-operator/docs/manual/quick-start-of-ob-operator\" target=\"_blank\">Docs</a>",
	},
	{
		Title:       "Advanced",
		Description: "Create an OceanBase database with customized configurations. <a href=\"https://oceanbase.github.io/ob-operator/docs/manual/ob-operator-user-guide/cluster-management-of-ob-operator/create-cluster\" target=\"_blank\">Docs</a>",
	},
	{
		Title:       "Tenants",
		Description: "Create and manage tenants in OceanBase database. <a href=\"https://oceanbase.github.io/ob-operator/docs/manual/ob-operator-user-guide/tenant-management-of-ob-operator/tenant-management-intro\" target=\"_blank\">Docs</a>",
	},
	{
		Title:       "High availability",
		Description: "Enable high availability for OceanBase on K8s. <a href=\"https://oceanbase.github.io/ob-operator/docs/manual/ob-operator-user-guide/high-availability/high-availability-intro\" target=\"_blank\">Docs</a>",
	},
	{
		Title:       "Get help from the community",
		Description: "Feel free to ask questions or report issues on GitHub: <a href=\"https://github.com/oceanbase/ob-operator/issues\" target=\"_blank\">Github</a>. Other ways to get help: <a href=\"https://oceanbase.github.io/ob-operator/#getting-help\" target=\"_blank\">Getting Help</a>",
	},
}
