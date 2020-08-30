package expr

import (
	"fmt"

	"goa.design/model/design"
)

type (
	// Relationship describes a uni-directional relationship between two elements.
	Relationship struct {
		ID               string
		Source           *Element
		Description      string
		Technology       string
		InteractionStyle design.InteractionStyleKind
		Tags             string
		URL              string

		// DestinationPath is used to compute the destination after all DSL has
		// completed execution.
		DestinationPath string

		// Destination is only guaranteed to be initialized after the DSL has
		// been executed. It can be used in validations and finalizers.
		Destination *Element

		// LinkedRelationshipID is the ID of the relationship pointing to the
		// container corresponding to the container instance with this
		// relationship.
		LinkedRelationshipID string
	}
)

// EvalName is the qualified name of the expression.
func (r *Relationship) EvalName() string {
	var src, dest = "<unknown source>", "<unknown destination>"
	if r.Source != nil {
		src = r.Source.Name
	}
	if r.Destination != nil {
		dest = r.Destination.Name
	}
	return fmt.Sprintf("relationship %q [%s -> %s]", r.Description, src, dest)
}

// Finalize computes the destination and adds the "Relationship" tag.
func (r *Relationship) Finalize() {
	r.MergeTags("Relationship")
}

// Dup creates a new relationship with identical description, tags, URL,
// technology and interaction style as r. Dup also creates a new ID for the
// result.
func (r *Relationship) Dup(newSrc, newDest *Element) *Relationship {
	dup := &Relationship{
		Source:           newSrc,
		InteractionStyle: r.InteractionStyle,
		Tags:             r.Tags,
		URL:              r.URL,
		Destination:      newDest,
		Description:      r.Description,
		Technology:       r.Technology,
	}
	Identify(dup)
	return dup
}

// MergeTags adds the given tags. It skips tags already present in e.Tags.
func (r *Relationship) MergeTags(tags ...string) {
	r.Tags = mergeTags(r.Tags, tags)
}
