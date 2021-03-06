package workflow

import (
	"sort"
	"strings"

	"github.com/ovh/cds/sdk"
)

// Sort sorts all the workflow tree
func Sort(w *sdk.Workflow) {
	SortNode(w.Root)
}

// SortNode sort the content of a node
func SortNode(n *sdk.WorkflowNode) {
	sortNodeHooks(&n.Hooks)
	sortNodeTriggers(&n.Triggers)
}

func sortNodeHooks(hooks *[]sdk.WorkflowNodeHook) {
	for i := range *hooks {
		sortConditions(&(*hooks)[i].Conditions)
	}

	sort.Slice(*hooks, func(i, j int) bool {
		return (*hooks)[i].UUID < (*hooks)[j].UUID
	})

}

func sortNodeTriggers(triggers *[]sdk.WorkflowNodeTrigger) {
	for i := range *triggers {
		sortConditions(&(*triggers)[i].Conditions)
		sortNodeHooks(&(*triggers)[i].WorkflowDestNode.Hooks)
	}

	sort.Slice(*triggers, func(i, j int) bool {
		t1 := &(*triggers)[i]
		t2 := &(*triggers)[j]

		return strings.Compare(t1.WorkflowDestNode.Name, t2.WorkflowDestNode.Name) < 0
	})
}

func sortConditions(conditions *[]sdk.WorkflowTriggerCondition) {
	sort.Slice(*conditions, func(i, j int) bool {
		return strings.Compare((*conditions)[i].Variable, (*conditions)[j].Variable) < 0
	})
}

func sortEnvironment(c1, c2 *sdk.WorkflowNodeContext) bool {
	if c1.Environment == nil {
		return true
	}
	if c1.Environment != nil && c2.Environment != nil {
		if c1.Environment.Name == c2.Environment.Name {
			return true
		}
		return strings.Compare(c1.Application.Name, c2.Application.Name) < 0
	}
	return false
}
