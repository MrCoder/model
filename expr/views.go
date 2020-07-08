package expr

import (
	"fmt"

	"goa.design/goa/v3/eval"
)

type (
	// Views is the container for all views.
	Views struct {
		// LandscapeViewss describe the system landscape views.
		LandscapeViews []*LandscapeView `json:"systemLandscapeViews,omitempty"`
		// ContextViews lists the system context views.
		ContextViews []*ContextView `json:"systemContextViews,omitempty"`
		// ContainerViews lists the container views.
		ContainerViews []*ContainerView `json:"containerViews,omitempty"`
		// ComponentViews lists the component views.
		ComponentViews []*ComponentView `json:"componentViews,omitempty"`
		// DynamicViews lists the dynamic views.
		DynamicViews []*DynamicView `json:"dynamicViews,omitempty"`
		// DeploymentViews lists the deployment views.
		DeploymentViews []*DeploymentView `json:"deploymentViews,omitempty"`
		// FilteredViews lists the filtered views.
		FilteredViews []*FilteredView `json:"filteredViews,omitempty"`
		// DSL to be run once all elements have been evaluated.
		DSL func() `json:"-"`
	}

	// LandscapeView describes a system landscape view.
	LandscapeView struct {
		ViewProps
		// EnterpriseBoundaryVisible specifies whether the enterprise boundary
		// (to differentiate internal elements from external elements) should be
		// visible on the resulting diagram.
		EnterpriseBoundaryVisible bool `json:"enterpriseBoundaryVisible"`
	}

	// ContextView describes a system context view.
	ContextView struct {
		ViewProps
		// EnterpriseBoundaryVisible specifies whether the enterprise boundary
		// (to differentiate internal elements from external elements) should be
		// visible on the resulting diagram.
		EnterpriseBoundaryVisible bool `json:"enterpriseBoundaryVisible"`
		// SoftwareSystemID is the ID of the software system this view with is
		// associated with.
		SoftwareSystemID string `json:"softwareSystemId"`
	}

	// ContainerView describes a container view for a specific software
	// system.
	ContainerView struct {
		ViewProps
		// Specifies whether software system boundaries should be visible for
		// "external" containers (those outside the software system in scope).
		ExternalSoftwareSystemBoundariesVisible bool `json:"externalSoftwareSystemBoundariesVisible"`
		// SoftwareSystemID is the ID of the software system this view with is
		// associated with.
		SoftwareSystemID string `json:"softwareSystemId"`
	}

	// ComponentView describes a component view for a specific container.
	ComponentView struct {
		ViewProps
		// Specifies whether container boundaries should be visible for
		// "external" containers (those outside the container in scope).
		ExternalContainerBoundariesVisible bool `json:"externalContainersBoundariesVisible"`
		// The ID of the container this view is associated with.
		ContainerID string `json:"containerID"`
	}

	// DynamicView describes a dynamic view for a specified scope.
	DynamicView struct {
		ViewProps
		// ElementID is the identifier of the element this view is associated with.
		ElementID string
	}

	// DeploymentView describes a deployment view.
	DeploymentView struct {
		ViewProps
		// SoftwareSystemID is the ID of the software system this view with is
		// associated with.
		SoftwareSystemID string `json:"softwareSystemId"`
		// The name of the environment that this deployment view is for (e.g.
		// "Development", "Live", etc).
		Environment string `json:"environment"`
	}

	// View is the common interface for all views.
	View interface {
		ElementView(string) *ElementView
		RelationshipView(string) *RelationshipView
		AllTagged(tag string) []*Element
		AllUnreachable(e ElementHolder) []*Element
		AllUnrelated() []*Element
		AddRelationships(...*Relationship)
		Remove(id string)
		Props() *ViewProps
	}

	// ViewAdder is the interface implemented by views that allow adding
	// elements and animations explicitly.
	ViewAdder interface {
		AddElements(...ElementHolder) error
		AddAnimation([]ElementHolder) error
	}
)

var (
	// Make sure views implement View.
	_ View = &LandscapeView{}
	_ View = &ContextView{}
	_ View = &ContainerView{}
	_ View = &ComponentView{}
	_ View = &DynamicView{}
	_ View = &DeploymentView{}

	// Make sure static views implement ElementAdder.
	_ ViewAdder = &LandscapeView{}
	_ ViewAdder = &ContextView{}
	_ ViewAdder = &ContainerView{}
	_ ViewAdder = &ComponentView{}
	_ ViewAdder = &DeploymentView{}
)

// EvalName returns the generic expression name used in error messages.
func (vs *Views) EvalName() string {
	return "views"
}

// Validate makes sure the right element are in the right views, it also makes
// sure all animation steps have elements.
func (vs *Views) Validate() error {
	verr := new(eval.ValidationErrors)
	checkElements := func(title string, evs []*ElementView, allowContainers bool) {
		var suffix = " and people"
		if allowContainers {
			suffix = ", people and containers"
		}
		for _, ev := range evs {
			if GetSoftwareSystem(ev.ID) != nil {
				continue
			}
			if GetPerson(ev.ID) != nil {
				continue
			}
			if allowContainers && GetContainer(ev.ID) != nil {
				continue
			}
			verr.Add(vs, fmt.Sprintf("%s can only contain software systems%s", title, suffix))
		}
	}
	for _, lv := range vs.LandscapeViews {
		checkElements("software landscape views", lv.ElementViews, false)
	}
	for _, cv := range vs.ContextViews {
		checkElements("software context views", cv.ElementViews, false)
	}
	for _, cv := range vs.ContainerViews {
		checkElements("container views", cv.ElementViews, true)
	}

	for _, v := range vs.all() {
		for i, s := range v.Animations {
			if len(s.ElementIDs) == 0 {
				var suf string
				if v.Key != "" {
					suf = fmt.Sprintf(" with key %q", v.Key)
				}
				verr.AddError(v, fmt.Errorf("animation at index %d in view%s introduces no new elements", i, suf))
			}
		}
	}

	return verr
}

// all returns all the views in a single slice.
func (vs Views) all() (vps []*ViewProps) {
	for _, lv := range vs.LandscapeViews {
		vps = append(vps, &lv.ViewProps)
	}
	for _, cv := range vs.ContextViews {
		vps = append(vps, &cv.ViewProps)
	}
	for _, cv := range vs.ContainerViews {
		vps = append(vps, &cv.ViewProps)
	}
	for _, cv := range vs.ComponentViews {
		vps = append(vps, &cv.ViewProps)
	}
	for _, dv := range vs.DynamicViews {
		vps = append(vps, &dv.ViewProps)
	}
	for _, dv := range vs.DeploymentViews {
		vps = append(vps, &dv.ViewProps)
	}
	return
}

// AddElements adds the given elements to the view if not already present.
func (cv *LandscapeView) AddElements(ehs ...ElementHolder) error {
	for _, eh := range ehs {
		if !isPS(eh) {
			return fmt.Errorf("elements of type %T cannot be added to landscape view", eh)
		}
	}
	addElements(&cv.ViewProps, ehs...)
	return nil
}

// AddRelationships adds the given relationships to the view if not already
// present. It does nothing if the relationship source and destination are not
// already in the view.
func (cv *LandscapeView) AddRelationships(rels ...*Relationship) {
	addRelationships(&cv.ViewProps, rels)
}

// AddAnimation adds the given animation steps to the view.
func (cv *LandscapeView) AddAnimation(ehs []ElementHolder) error {
	return addAnimation(&cv.ViewProps, ehs)
}

// Remove given element from view.
func (cv *LandscapeView) Remove(id string) {
	remove(&cv.ViewProps, id)
}

// AddElements adds the given elements to the view if not already present.
func (cv *ContextView) AddElements(ehs ...ElementHolder) error {
	for _, eh := range ehs {
		if !isPS(eh) {
			return fmt.Errorf("elements of type %T cannot be added to context view", eh)
		}
	}
	addElements(&cv.ViewProps, ehs...)
	return nil
}

// AddRelationships adds the given relationships to the view if not already
// present. It does nothing if the relationship source and destination are not
// already in the view.
func (cv *ContextView) AddRelationships(rels ...*Relationship) {
	addRelationships(&cv.ViewProps, rels)
}

// AddAnimation adds the given animation steps to the view.
func (cv *ContextView) AddAnimation(ehs []ElementHolder) error {
	return addAnimation(&cv.ViewProps, ehs)
}

// Remove given element from view if not software system this view is for.
func (cv *ContextView) Remove(id string) {
	if id == cv.SoftwareSystemID {
		return
	}
	remove(&cv.ViewProps, id)
}

// AddElements adds the given elements to the view if not already present.
func (cv *ContainerView) AddElements(ehs ...ElementHolder) error {
	for _, eh := range ehs {
		if !isPSC(eh) {
			return fmt.Errorf("elements of type %T cannot be added to container view", eh)
		}
	}
	addElements(&cv.ViewProps, ehs...)
	return nil
}

// AddRelationships adds the given relationships to the view if not already
// present. It does nothing if the relationship source and destination are not
// already in the view.
func (cv *ContainerView) AddRelationships(rels ...*Relationship) {
	addRelationships(&cv.ViewProps, rels)
}

// AddAnimation adds the given animation steps to the view.
func (cv *ContainerView) AddAnimation(ehs []ElementHolder) error {
	return addAnimation(&cv.ViewProps, ehs)
}

// Remove given element from view if not software system this view is for.
func (cv *ContainerView) Remove(id string) {
	if id == cv.SoftwareSystemID {
		return
	}
	remove(&cv.ViewProps, id)
}

// AddElements adds the given elements to the view if not already present.
func (cv *ComponentView) AddElements(ehs ...ElementHolder) error {
	for _, eh := range ehs {
		if !isPSCC(eh) {
			return fmt.Errorf("elements of type %T cannot be added to component view", eh)
		}
	}
	addElements(&cv.ViewProps, ehs...)
	return nil
}

// AddRelationships adds the given relationships to the view if not already
// present. It does nothing if the relationship source and destination are not
// already in the view.
func (cv *ComponentView) AddRelationships(rels ...*Relationship) {
	addRelationships(&cv.ViewProps, rels)
}

// AddAnimation adds the given animation steps to the view.
func (cv *ComponentView) AddAnimation(ehs []ElementHolder) error {
	return addAnimation(&cv.ViewProps, ehs)
}

// Remove given element from view if not software system this view is for.
func (cv *ComponentView) Remove(id string) {
	if id == cv.ContainerID || id == GetContainer(cv.ContainerID).System.ID {
		return
	}
	remove(&cv.ViewProps, id)
}

// AddRelationships adds the given relationships to the view if not already
// present. It does nothing if the relationship source and destination are not
// already in the view.
func (cv *DynamicView) AddRelationships(rels ...*Relationship) {
	addRelationships(&cv.ViewProps, rels)
}

// Remove given element from view.
func (cv *DynamicView) Remove(id string) {
	remove(&cv.ViewProps, id)
}

// AddElements adds the given elements to the view if not already present.
func (dv *DeploymentView) AddElements(ehs ...ElementHolder) error {
	var nodes []*DeploymentNode
	for _, eh := range ehs {
		n, ok := eh.(*DeploymentNode)
		if !ok {
			return fmt.Errorf("elements of type %T cannot be added to deployment views, only deployment nodes can", eh)
		}
		nodes = append(nodes, n)
	}

	for _, n := range nodes {
		if addDeploymentNode(dv, n) {
			p := n.Parent
			for p != nil {
				addElements(&dv.ViewProps, p)
				p = p.Parent
			}
		}
	}

	return nil
}

func addDeploymentNode(dv *DeploymentView, n *DeploymentNode) bool {
	var nested bool
	for _, inst := range n.ContainerInstances {
		if dv.SoftwareSystemID == "" || GetContainer(inst.ContainerID).System.ID == dv.SoftwareSystemID {
			addElements(&dv.ViewProps, inst)
			nested = true
		}
	}
	for _, inf := range n.InfrastructureNodes {
		addElements(&dv.ViewProps, inf)
		nested = true
	}
	for _, c := range n.Children {
		nested = nested || addDeploymentNode(dv, c)
	}
	if nested {
		addElements(&dv.ViewProps, n)
	}
	return nested
}

// AddRelationships adds the given relationships to the view if not already
// present. It does nothing if the relationship source and destination are not
// already in the view.
func (dv *DeploymentView) AddRelationships(rels ...*Relationship) {
	addRelationships(&dv.ViewProps, rels)
}

// AddAnimation adds the given animation steps to the view.
func (dv *DeploymentView) AddAnimation(ehs []ElementHolder) error {
	return addAnimation(&dv.ViewProps, ehs)
}

// Remove given deployment node from view.
func (dv *DeploymentView) Remove(id string) {
	n := GetDeploymentNode(id)
	for _, ci := range n.ContainerInstances {
		remove(&dv.ViewProps, ci.ID)
	}
	for _, inf := range n.InfrastructureNodes {
		remove(&dv.ViewProps, inf.ID)
	}
	for _, c := range n.Children {
		dv.Remove(c.ID)
	}
	remove(&dv.ViewProps, id)
}

// addElements adds the given elements to the view if not already present.
func addElements(v *ViewProps, ehs ...ElementHolder) {
loop:
	for _, eh := range ehs {
		id := eh.GetElement().ID
		for _, e := range v.ElementViews {
			if e.ID == id {
				continue loop
			}
		}
		v.ElementViews = append(v.ElementViews, &ElementView{ID: id, Element: eh.GetElement()})

		// Complete relationships
		var rels []*Relationship
		for _, r := range AllRelationships() {
			if r.SourceID == id {
				if v.ElementView(r.DestinationID) != nil {
					rels = append(rels, r)
				}
			} else if r.DestinationID == id {
				if v.ElementView(r.SourceID) != nil {
					rels = append(rels, r)
				}
			}
		}
		addRelationships(v, rels)
	}
}

// addRelationships adds the given relationships to the view if not already
// present. It does nothing if the relationship source and destination are not
// already in the view.
func addRelationships(v *ViewProps, rels []*Relationship) {
loop:
	for _, r := range rels {
		var src, dest bool
		for _, ev := range v.ElementViews {
			if ev.ID == r.SourceID {
				src = true
				if dest {
					break
				}
			}
			if ev.ID == r.DestinationID {
				dest = true
				if src {
					break
				}
			}
		}
		if !src || !dest {
			continue loop
		}
		for _, rv := range v.RelationshipViews {
			if rv.ID == r.ID {
				continue loop
			}
		}
		v.RelationshipViews = append(v.RelationshipViews, &RelationshipView{ID: r.ID, Relationship: r})
	}
}

// addAnimation adds the animations to the view after normalizing their content.
// It makes sure that all the elements are in the view, that each element is
// only included in one animation step, that the relationships are initialized
// and that any dependent deployment nodes is added.
func addAnimation(v *ViewProps, ehs []ElementHolder) error {
	var known []string
	for _, s := range v.Animations {
		for _, id := range s.ElementIDs {
			known = append(known, id)
		}
	}
	n := &Animation{Order: len(v.Animations) + 1}
loop:
	for _, e := range ehs {
		id := e.GetElement().ID
		if v.ElementView(id) == nil {
			continue loop // item not in view, skip
		}
		for _, k := range known {
			if k == id {
				continue loop // item already in a step, skip
			}
		}

		known = append(known, id)
		n.ElementIDs = append(n.ElementIDs, id)
		n.Elements = append(n.Elements, e)

		// Add parent deployment nodes for infrastructure nodes
		// and container instances.
		var node *DeploymentNode
		if inf, ok := e.(*InfrastructureNode); ok {
			node = inf.Parent
		} else if ci, ok := e.(*ContainerInstance); ok {
			node = ci.Parent
		}
		for node != nil {
			known = append(known, node.ID)
			n.ElementIDs = append(n.ElementIDs, node.ID)
			n.Elements = append(n.Elements, node)
			node = node.Parent
		}
	}
	if len(n.ElementIDs) == 0 {
		return fmt.Errorf("None of the specified elements exist in this view or do not appear in previous animation steps")
	}

	// Add relationships between new elements and elements in
	// previous steps.
	for _, rv := range v.RelationshipViews {
		var newSrc, newDest, oldSrc, oldDest bool
		for _, s := range v.Animations {
			for _, id := range s.ElementIDs {
				if id == rv.Relationship.SourceID {
					oldSrc = true
				} else if id == rv.Relationship.DestinationID {
					oldDest = true
				}
				if oldSrc && oldDest {
					break
				}
			}
			if oldSrc && oldDest {
				break
			}
		}
		for _, id := range n.ElementIDs {
			if id == rv.Relationship.SourceID {
				newSrc = true
			} else if id == rv.Relationship.DestinationID {
				newDest = true
			}
			if newSrc && newDest {
				break
			}
		}
		if newSrc && oldDest || oldSrc && newDest {
			n.Relationships = append(n.Relationships, rv.ID)
		}
	}

	v.Animations = append(v.Animations, n)
	return nil
}

// removes the element with the given ID from the view if present.
func remove(v *ViewProps, id string) {
	idx := v.index(id)
	if idx == -1 {
		return
	}
	v.ElementViews = append(v.ElementViews[:idx], v.ElementViews[idx+1:]...)

	// Remove corresponding relationships.
	var ids []string
	for _, r := range v.RelationshipViews {
		if r.Relationship.SourceID == id {
			ids = append(ids, id)
		} else if r.Relationship.DestinationID == id {
			ids = append(ids, id)
		}
	}
	rvs := v.RelationshipViews
	tmp := rvs[:0]
	for _, r := range rvs {
		remove := false
		for _, id := range ids {
			if r.ID == id {
				remove = true
				break
			}
		}
		if !remove {
			tmp = append(tmp, r)
		}
	}
	v.RelationshipViews = tmp
}

// isPS returns true if element is a person or software system, false otherwise.
func isPS(eh ElementHolder) bool {
	switch eh.(type) {
	case *Person, *SoftwareSystem:
		return true
	}
	return false
}

// isPSC returns true if element is a person, a software system or a container,
// false otherwise.
func isPSC(eh ElementHolder) bool {
	if isPS(eh) {
		return true
	}
	_, ok := eh.(*Container)
	return ok
}

// isPSCC returns true if element is a person, a software system, a container or
// a component, false otherwise.
func isPSCC(eh ElementHolder) bool {
	if isPSC(eh) {
		return true
	}
	_, ok := eh.(*Component)
	return ok
}
