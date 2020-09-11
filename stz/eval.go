package stz

import (
	"goa.design/goa/v3/eval"
	"goa.design/model/expr"
)

// RunDSL runs the DSL defined in a global variable and returns the corresponding
// Structurize workspace.
func RunDSL() (*Workspace, error) {
	if err := eval.RunDSL(); err != nil {
		return nil, err
	}
	return WorkspaceFromDesign(expr.Root), nil
}

// WorkspaceFromDesign returns a Structurizr workspace initialized from the
// given design.
func WorkspaceFromDesign(d *expr.Design) *Workspace {
	model := &Model{}
	m := d.Model
	if name := m.Enterprise; name != "" {
		model.Enterprise = &Enterprise{Name: name}
	}
	model.People = make([]*Person, len(m.People))
	for i, p := range m.People {
		model.People[i] = modelizePerson(p)
	}
	model.Systems = make([]*SoftwareSystem, len(m.Systems))
	for i, sys := range m.Systems {
		model.Systems[i] = modelizeSystem(sys)
	}
	model.DeploymentNodes = modelizeDeploymentNodes(m.DeploymentNodes)

	views := &Views{}
	v := d.Views
	views.LandscapeViews = make([]*LandscapeView, len(v.LandscapeViews))
	for i, lv := range v.LandscapeViews {
		views.LandscapeViews[i] = &LandscapeView{
			ViewProps:                 modelizeProps(lv.Props()),
			EnterpriseBoundaryVisible: lv.EnterpriseBoundaryVisible,
		}
	}
	views.ContextViews = make([]*ContextView, len(v.ContextViews))
	for i, lv := range v.ContextViews {
		views.ContextViews[i] = &ContextView{
			ViewProps:                 modelizeProps(lv.Props()),
			EnterpriseBoundaryVisible: lv.EnterpriseBoundaryVisible,
			SoftwareSystemID:          lv.SoftwareSystemID,
		}
	}
	views.ContainerViews = make([]*ContainerView, len(v.ContainerViews))
	for i, cv := range v.ContainerViews {
		views.ContainerViews[i] = &ContainerView{
			ViewProps:               modelizeProps(cv.Props()),
			SystemBoundariesVisible: cv.SystemBoundariesVisible,
			SoftwareSystemID:        cv.SoftwareSystemID,
		}
	}
	views.ComponentViews = make([]*ComponentView, len(v.ComponentViews))
	for i, cv := range v.ComponentViews {
		views.ComponentViews[i] = &ComponentView{
			ViewProps:                  modelizeProps(cv.Props()),
			ContainerBoundariesVisible: cv.ContainerBoundariesVisible,
			ContainerID:                cv.ContainerID,
		}
	}
	views.DynamicViews = make([]*DynamicView, len(v.DynamicViews))
	for i, dv := range v.DynamicViews {
		views.DynamicViews[i] = &DynamicView{
			ViewProps: modelizeProps(dv.Props()),
			ElementID: dv.ElementID,
		}
	}
	views.FilteredViews = make([]*FilteredView, len(v.FilteredViews))
	for i, lv := range v.FilteredViews {
		mode := "Include"
		if lv.Exclude {
			mode = "Exclude"
		}
		views.FilteredViews[i] = &FilteredView{
			Title:       lv.Title,
			Description: lv.Description,
			Key:         lv.Key,
			BaseKey:     lv.BaseKey,
			Mode:        mode,
			Tags:        lv.FilterTags,
		}
	}
	views.Configuration = &Configuration{Styles: modelizeStyles(v.Styles)}

	return &Workspace{
		Name:        d.Name,
		Description: d.Description,
		Version:     d.Version,
		Model:       model,
		Views:       views,
	}
}

func modelizePerson(p *expr.Person) *Person {
	return &Person{
		ID:            p.Element.ID,
		Name:          p.Element.Name,
		Description:   p.Element.Description,
		Technology:    p.Element.Technology,
		Tags:          p.Element.Tags,
		URL:           p.Element.URL,
		Properties:    p.Element.Properties,
		Relationships: modelizeRelationships(p.Relationships),
		Location:      LocationKind(p.Location),
	}
}

func modelizeRelationships(rels []*expr.Relationship) []*Relationship {
	res := make([]*Relationship, len(rels))
	for i, r := range rels {
		res[i] = &Relationship{
			ID:                   r.ID,
			Description:          r.Description,
			Tags:                 r.Tags,
			URL:                  r.URL,
			SourceID:             r.Source.ID,
			DestinationID:        r.Destination.ID,
			Technology:           r.Technology,
			InteractionStyle:     InteractionStyleKind(r.InteractionStyle),
			LinkedRelationshipID: r.LinkedRelationshipID,
		}
	}
	return res
}

func modelizeSystem(sys *expr.SoftwareSystem) *SoftwareSystem {
	return &SoftwareSystem{
		ID:            sys.ID,
		Name:          sys.Name,
		Description:   sys.Description,
		Technology:    sys.Technology,
		Tags:          sys.Tags,
		URL:           sys.URL,
		Properties:    sys.Properties,
		Relationships: modelizeRelationships(sys.Relationships),
		Location:      LocationKind(sys.Location),
		Containers:    modelizeContainers(sys.Containers),
	}
}

func modelizeContainers(cs []*expr.Container) []*Container {
	res := make([]*Container, len(cs))
	for i, c := range cs {
		res[i] = &Container{
			ID:            c.ID,
			Name:          c.Name,
			Description:   c.Description,
			Technology:    c.Technology,
			Tags:          c.Tags,
			URL:           c.URL,
			Properties:    c.Properties,
			Relationships: modelizeRelationships(c.Relationships),
			Components:    modelizeComponents(c.Components),
		}
	}
	return res
}

func modelizeComponents(cs []*expr.Component) []*Component {
	res := make([]*Component, len(cs))
	for i, c := range cs {
		res[i] = &Component{
			ID:            c.ID,
			Name:          c.Name,
			Description:   c.Description,
			Technology:    c.Technology,
			Tags:          c.Tags,
			URL:           c.URL,
			Properties:    c.Properties,
			Relationships: modelizeRelationships(c.Relationships),
		}
	}
	return res
}

func modelizeDeploymentNodes(dns []*expr.DeploymentNode) []*DeploymentNode {
	res := make([]*DeploymentNode, len(dns))
	for i, dn := range dns {
		res[i] = &DeploymentNode{
			ID:          dn.ID,
			Name:        dn.Name,
			Description: dn.Description,
			Technology:  dn.Technology,
			Environment: dn.Environment,
			Instances:   dn.Instances,
			Tags:        dn.Tags,
			URL:         dn.URL,
		}
	}
	return res
}

func modelizeProps(prop *expr.ViewProps) *ViewProps {
	props := &ViewProps{
		Title:             prop.Title,
		Description:       prop.Description,
		Key:               prop.Key,
		PaperSize:         PaperSizeKind(prop.PaperSize),
		ElementViews:      modelizeElementViews(prop.ElementViews),
		RelationshipViews: modelizeRelationshipViews(prop.RelationshipViews),
		Animations:        modelizeAnimationSteps(prop.AnimationSteps),
	}
	if layout := prop.AutoLayout; layout != nil {
		props.AutoLayout = &AutoLayout{
			RankDirection: RankDirectionKind(layout.RankDirection),
			RankSep:       layout.RankSep,
			NodeSep:       layout.NodeSep,
			EdgeSep:       layout.EdgeSep,
			Vertices:      layout.Vertices,
		}
	}
	return props
}

func modelizeElementViews(evs []*expr.ElementView) []*ElementView {
	res := make([]*ElementView, len(evs))
	for i, ev := range evs {
		res[i] = &ElementView{
			ID: ev.Element.ID,
			X:  ev.X,
			Y:  ev.Y,
		}
	}
	return res
}

func modelizeRelationshipViews(rvs []*expr.RelationshipView) []*RelationshipView {
	res := make([]*RelationshipView, len(rvs))
	for i, rv := range rvs {
		vertices := make([]*Vertex, len(rv.Vertices))
		for i, v := range rv.Vertices {
			vertices[i] = &Vertex{v.X, v.Y}
		}
		res[i] = &RelationshipView{
			ID:          rv.RelationshipID,
			Description: rv.Description,
			Order:       rv.Order,
			Vertices:    vertices,
			Routing:     RoutingKind(rv.Routing),
			Position:    rv.Position,
		}
	}
	return res
}

func modelizeAnimationSteps(as []*expr.AnimationStep) []*AnimationStep {
	res := make([]*AnimationStep, len(as))
	for i, s := range as {
		elems := make([]string, len(s.Elements))
		for i, e := range s.Elements {
			elems[i] = e.GetElement().ID
		}
		res[i] = &AnimationStep{
			Order:         s.Order,
			Elements:      elems,
			Relationships: s.RelationshipIDs,
		}
	}
	return res
}

func modelizeStyles(s *expr.Styles) *Styles {
	elems := make([]*ElementStyle, len(s.Elements))
	for i, es := range s.Elements {
		elems[i] = &ElementStyle{
			Tag:         es.Tag,
			Background:  es.Background,
			Stroke:      es.Stroke,
			Color:       es.Color,
			Shape:       ShapeKind(es.Shape),
			Metadata:    es.Metadata,
			Description: es.Description,
			Opacity:     es.Opacity,
		}
	}
loopelem:
	for _, ses := range s.StructurizrElements {
		for _, es := range elems {
			if ses.Tag == es.Tag {
				es.Width = ses.Width
				es.Height = ses.Height
				es.FontSize = ses.FontSize
				es.Icon = ses.Icon
				es.Border = BorderKind(ses.Border)
				continue loopelem
			}
		}
		elems = append(elems, &ElementStyle{
			Tag:      ses.Tag,
			Width:    ses.Width,
			Height:   ses.Height,
			FontSize: ses.FontSize,
			Icon:     ses.Icon,
			Border:   BorderKind(ses.Border),
		})
	}
	rels := make([]*RelationshipStyle, len(s.Relationships))
	for i, rs := range s.Relationships {
		rels[i] = &RelationshipStyle{
			Tag:     rs.Tag,
			Color:   rs.Color,
			Dashed:  rs.Dashed,
			Routing: RoutingKind(rs.Routing),
			Opacity: rs.Opacity,
		}
		if rs.Thick != nil && *rs.Thick {
			six := 6
			rels[i].Thickness = &six
		}
	}
looprel:
	for _, res := range s.StructurizrRelationships {
		for _, rs := range rels {
			if res.Tag == rs.Tag {
				rs.Thickness = res.Thickness
				rs.Width = res.Width
				rs.FontSize = res.FontSize
				rs.Position = res.Position
				continue looprel
			}
		}
		rels = append(rels, &RelationshipStyle{
			Tag:       res.Tag,
			Thickness: res.Thickness,
			Width:     res.Width,
			FontSize:  res.FontSize,
			Position:  res.Position,
		})
	}
	return &Styles{
		Elements:      elems,
		Relationships: rels,
	}
}
