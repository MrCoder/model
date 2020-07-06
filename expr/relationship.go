package expr

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type (
	// Relationship describes a uni-directional relationship between two elements.
	Relationship struct {
		// ID of relationship.
		ID string `json:"id"`
		// Description of relationship if any.
		Description string `json:"description"`
		// Tags attached to relationship as comma separated list if any.
		Tags string `json:"tags"`
		// URL where more information can be found.
		URL string `json:"url"`
		// SourceID is the ID of the source element.
		SourceID string `json:"sourceId"`
		// DestinationID is ID the destination element.
		DestinationID string `json:"destinationId"`
		// Technology associated with relationship.
		Technology string `json:"technology"`
		// InteractionStyle describes whether the interaction is synchronous or asynchronous
		InteractionStyle InteractionStyleKind
		// ID of container-container relationship upon which this container
		// instance-container instance relationship is based.
		LinkedRelationshipID string `json:"linkedRelationshipId"`
		// Source Element.
		Source *Element `json:"-"`
		// Destination Element.
		Destination *Element `json:"-"`
	}

	// InteractionStyleKind is the enum for possible interaction styles.
	InteractionStyleKind int
)

const (
	// InteractionUndefined means no interaction style specified in design.
	InteractionUndefined InteractionStyleKind = iota
	// InteractionSynchronous describes a synchronous interaction.
	InteractionSynchronous
	// InteractionAsynchronous describes an asynchronous interaction.
	InteractionAsynchronous
)

// EvalName is the qualified name of the expression.
func (r *Relationship) EvalName() string {
	return fmt.Sprintf("%s [%s -> %s]", r.Description, r.SourceID, r.DestinationID)
}

// MarshalJSON replaces the constant value with the proper structurizr schema
// string value.
func (i InteractionStyleKind) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBufferString(`"`)
	switch i {
	case InteractionSynchronous:
		buf.WriteString("Synchronous")
	case InteractionAsynchronous:
		buf.WriteString("Asynchronous")
	case InteractionUndefined:
		buf.WriteString("Undefined")
	}
	buf.WriteString(`"`)
	return buf.Bytes(), nil
}

// UnmarshalJSON sets the constant from its JSON representation.
func (i *InteractionStyleKind) UnmarshalJSON(data []byte) error {
	var val string
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	switch val {
	case "Synchronous":
		*i = InteractionSynchronous
	case "Asynchronous":
		*i = InteractionAsynchronous
	case "Undefined":
		*i = InteractionUndefined
	}
	return nil
}
