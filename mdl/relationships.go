package mdl

import (
	"goa.design/goa/v3/codegen"
	"goa.design/model/expr"
)

type (
	relationshipData struct {
		// Relationship source and destination element IDs.
		SourceID, DestinationID string
		// Description of relationship
		Description string
		// Start and End link mermaid symbols (e.g. "--", "->")
		Start, End string
		// Technology used for relationship if any
		Technology string
	}
)

func relationships(rvs []*expr.RelationshipView) *codegen.SectionTemplate {
	data := make([]*relationshipData, len(rvs))
	for i, rv := range rvs {
		rel := expr.Registry[rv.RelationshipID].(*expr.Relationship)
		start, end := lineStartEnd(rv)
		data[i] = &relationshipData{
			SourceID:      rv.Source.ID,
			DestinationID: rv.Destination.ID,
			Description:   rv.Description,
			Start:         start,
			End:           end,
			Technology:    rel.Technology,
		}
	}
	funcs := map[string]interface{}{"wrap": wrap, "indent": indent}
	return &codegen.SectionTemplate{Name: "relationships", Source: relationshipT, Data: data, FuncMap: funcs}
}

func lineStartEnd(rv *expr.RelationshipView) (string, string) {
	rs := relStyle(rv)
	if rs.Dashed == nil || *rs.Dashed {
		return "-.", ".->"
	}
	if rs.Thick != nil && *rs.Thick {
		return "==", "==>"
	}
	return "--", "-->"
}

const relationshipT = `{{ range . -}}
{{ indent 1 }}{{ .SourceID }} {{ .Start }}"<div class='relationship'><div class='relationship-label'>{{ wrap .Description }}</div>
{{- if .Technology }}<div class='relationship-technology'>[{{ .Technology }}]</div></div>
{{- end }}"{{ .End }}{{ .DestinationID }}
{{ end }}`
