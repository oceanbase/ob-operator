package k8s

import (
	"testing"

	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/stretchr/testify/assert"
)

func TestListEvents(t *testing.T) {
	events, err := ListEvents(&param.QueryEventParam{
		ObjectType: "Pod",
		Type:       "Normal",
		Namespace:  "kube-system",
	})
	assert.Nil(t, err)
	assert.NotNil(t, events)
}

func TestListNamespaces(t *testing.T) {
	namespaces, err := ListNamespaces()
	assert.Nil(t, err)
	assert.NotNil(t, namespaces)
}

func TestListNodes(t *testing.T) {
	nodes, err := ListNodes()
	assert.Nil(t, err)
	assert.NotNil(t, nodes)
}

func TestListStorageClasses(t *testing.T) {
	scs, err := ListStorageClasses()
	assert.Nil(t, err)
	assert.NotNil(t, scs)
}
