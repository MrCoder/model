# Structurizr for Go

This GitHub repository is a non-official client library for the
[Structurizr](https://structurizr.com/) cloud service and on-premises
installation, both of which are web-based publishing platforms for software
architecture models based upon the [C4 model](https://c4model.com).

The repository defines a Go DSL that makes it convenient to describe the
software architecture model so that it can be uploaded to the Structurizr
service.

This library also provides a [Goa](https://github.com/goadesign/goa)
plugin so that the design of APIs and microservices written in Goa can be
augmented with a description of the corresponding software architecture.

## Example

Here is a complete and correct DSL for an architecture model:

```Go
var _ = Workspace("Getting Started", "This is a model of my software system.", func() {
    var System = SoftwareSystem("Software System", "My software system.")

    var User = Person("User", "A user of my software system.", func() {
        Uses(System, "Uses")
    })

    Views(func() {
        SystemContext(MySystem, "SystemContext", "An example of a System Context diagram.", func() {
            AddAll()
            AutoLayout()
        })
        Styles(func() {
            ElementStyle(System, func() {
                Background("#1168bd")
                Color("#ffffff")
             })
            ElementStyle(User, func() {
                Shape(ShapePerson)
                Background("#08427b")
                Color("#ffffff")
            })
        })
    })
})
```

This code creates a model containing elements and relationships, creates a
single view and adds some styling.
![Getting Started Diagram](https://structurizr.com/static/img/getting-started.png)

## Standalone Usage

The [eval](https://github.com/goadesign/structurizr/tree/master/eval) package
makes it convenient to run the DSL above. Here is a complete example that
uploads the workspace described in the DSL to the Structurizr service:

```Go
package main

include "goa.design/structurizr/eval"
include "goa.design/structurizr/client"

// DSL that describes software architecture model.
var _ = Workspace("Getting Started", "This is a model of my software system.", func() {
    var System = SoftwareSystem("Software System", "My software system.")

    var User = Person("User", "A user of my software system.", func() {
        Uses(System, "Uses")
    })

    Views(func() {
        SystemContext(MySystem, "SystemContext", "An example of a System Context diagram.", func() {
            AddAll()
            AutoLayout()
        })
        Styles(func() {
            ElementStyle(System, func() {
                Background("#1168bd")
                Color("#ffffff")
             })
            ElementStyle(User, func() {
                Shape("ShapePerson")
                Background("#08427b")
                Color("#ffffff")
            })
        })
    })
})

// Executes the DSL and uploads the corresponding workspace to Structurizr.
func main() {
    // Run the model DSL
    w, err := eval.RunDSL()
    if err != nil {
        fmt.Fprintf(os.Stderr, "invalid model: %s", err.String())
        os.Exit(1)
    }

    // Upload the model to the Structurizr service. 
    // The API key and secret must be set in the STRUCTURIZR_KEY and
    // STRUCTURIZR_SECRET environment variables respectively. The 
    // workspace ID must be set in STRUCTURIZR_WORKSPACE_ID.
    var (
        key = os.Getenv("STRUCTURIZR_KEY")
        secret = os.Getenv("STRUCTURIZR_SECRET")
        wid = os.Getenv("STRUCTURIZR_WORKSPACE_ID")
    )
    c := service.NewClient(key, secret)
    if err := c.Put(wid, w); err != nil {
        fmt.Fprintf(os.Stderr, "failed to store workspace: %s", err.String())
    }
}
```

## Goa Plugin

This package can also be used as a Goa plugin by including the DSL package in
the Goa design:

```Go
package design

import . "goa.design/goa/v3/dsl"
import . "goa.design/plugins/structurizr/dsl"

// ... DSL describing API, services and architecture model
```

Running `goa gen` creates a `structurizr.json` file in the `gen` folder. This
file follows the
[structurizr JSON schema](https://github.com/structurizr/json) and can be
uploaded to the Structurizr service for example using the
[Structurizr CLI](https://github.com/structurizr/cli).

## DSL Syntax

The code snippet below describes the entire syntax of the DSL.

```Go
// Workspace defines the workspace containing the models and views. Workspace
// must appear exactly once in a given design. A name must be provided if a
// description is.
var _ = Workspace("[name]", "[description]", func() {

    // Version number.
    Version("<version>")

    // Enterprise defines a named "enterprise" (e.g. an organisation). On System
    // Landscape and System Context diagrams, an enterprise is represented as a
    // dashed box. Only a single enterprise can be defined within a model.
    Enterprise("<name>")

    // Person defines a person (user, actor, role or persona).
    var Person = Person("<name>", "[description]", func() { // optional
        Tag("<name>", "[name]") // as many tags as needed

        // URL where more information about this system can be found.
        URL("<url>")

        // External indicates the person is external to the enterprise.
        External()

        // Properties define an arbitrary set of associated key-value pairs.
        Properties(func() {
            Prop("<name>", "<value">)
        })

        // Adds a uni-directional relationship between this person and the given element.
        Uses(Element, "<description>", "[technology]", Synchronous /* or Asynchronous */, func() {
            Tag("<name>", "[name]") // as many tags as needed
        })

        // Adds an interaction between this person and another.
        InteractsWith(Person, "<description>", "[technology]", Synchronous /* or Asynchronous */, func() {
            Tag("<name>", "[name]") // as many tags as needed
        })
    })

    // SoftwareSystem defines a software system.
    var SoftwareSystem = SoftwareSystem("<name>", "[description]", func() { // optional
        Tag("<name>",  "[name]") // as many tags as needed

        // URL where more information about this software system can be
        // found.
        URL("<url>") 

        // External indicates the software system is external to the enterprise.
        External()

        // Properties define an arbitrary set of associated key-value pairs.
        Properties(func() {
            Prop("<name>", "<value">)
        })

        // Adds a uni-directional relationship between this software system and the given element.
        Uses(Element, "<description>", "[technology]", Synchronous /* or Asynchronous */, func() {
            Tag("<name>", "[name]") // as many tags as needed
        })

        // Adds an interaction between this software system and a person.
        Delivers(Person, "<description>", "[technology]", Synchronous /* or Asynchronous */, func() {
            Tag("<name>", "[name]") // as many tags as needed
        })
    })

    // Container defines a container within a software system.
    var Container = Container(SoftwareSystem, "<name>",  "[description]",  "[technology]",  func() { // optional
        Tag("<name>",  "[name]") // as many tags as neede

        // URL where more information about this container can be found.
        URL("<url>")

        // Properties define an arbitrary set of associated key-value pairs.
        Properties(func() {
            Prop("<name>", "<value">)
        })

        // Adds a uni-directional relationship between this container and the given element.
        Uses(Element, "<description>", "[technology]", Synchronous /* or Asynchronous */, func () {
            Tag("<name>", "[name]") // as many tags as needed
        })

        // Adds an interaction between this container and a person.
        Delivers(Person, "<description>", "[technology]", Synchronous /* or Asynchronous */, func() {
            Tag("<name>", "[name]") // as many tags as needed
        })
    })

    // Container may also refer to a Goa service in which case the name
    // and description are taken from the given service definition and
    // the technology is set to "Go and Goa v3".
    var Container = Container(SoftwareSystem, GoaService, func() {
        // ... see above
    })

    // Component defines a component within a container.
    var Component = Component(Container, "<name>",  "[description]",  "[technology]",  func() { // optional
        Tag("<name>",  "[name]") // as many tags as neede

        // URL where more information about this container can be found.
        URL("<url>")

        // Properties define an arbitrary set of associated key-value pairs.
        Properties(func() {
            Prop("<name>", "<value">)
        })

        // Adds a uni-directional relationship between this component and the given element.
        Uses(Element, "<description>", "[technology]", Synchronous /* or Asynchronous */, func() {
            Tag("<name>", "[name]") // as many tags as needed
        })

        // Adds an interaction between this component and a person.
        Delivers(Person, "<description>", "[technology]", Synchronous /* or Asynchronous */, func() {
            Tag("<name>", "[name]") // as many tags as needed
        })
    })

    // DeploymentEnvironment provides a way to define a deployment
    // environment (e.g. development, staging, production, etc).
    DeploymentEnvironment("<name>", func() {

        // DeploymentNode defines a deployment node. Deployment nodes can be
        // nested, so a deployment node can contain other deployment nodes.
        // A deployment node can also contain InfrastructureNode and
        // ContainerInstance elements.
        var DeploymentNode = DeploymentNode("<name>",  "[description]",  "[technology]",  func() { // optional
            Tag("<name>",  "[name]") // as many tags as needed

            // Instances sets the number of instances, defaults to 1.
            Instances(2)

            // URL where more information about this deployment node can be
            // found.
            URL("<url>") 

            // Properties define an arbitrary set of associated key-value pairs.
            Properties(func() {
                Prop("<name>", "<value">)
            })
        })

        // InfrastructureNode defines an infrastructure node, typically
        // something like a load balancer, firewall, DNS service, etc.
        var InfrastructureNode = InfrastructureNode(DeploymentNode, "<name>", "[description]", "[technology]", func() { // optional
            Tag("<name>",  "[name]") // as many tags as needed

            // URL where more information about this infrastructure node can be
            // found.
            URL("<url>") 

            // Properties define an arbitrary set of associated key-value pairs.
            Properties(func() {
                Prop("<name>", "<value">)
            })
        })

        // ContainerInstance defines an instance of the specified
        // container that is deployed on the parent deployment node.
        var ContainerInstance = ContainerInstance(identifier, func() { // optional
            Tag("<name>",  "[name]") // as many tags as needed

            // Sets instance number or index.
            InstanceID(1)
            
            // Properties define an arbitrary set of associated key-value pairs.
            Properties(func() {
                Prop("<name>", "<value">)
            })

            // HealthCheck defines a HTTP-based health check for this
            // container instance.
            HealthCheck("<name>", func() {

                // URL is the health check URL/endpoint.
                URL("<url>")
                
                // Interval is the polling interval, in seconds.
                Interval(42)

                // Timeout after which a health check is deemed as failed,
                // in milliseconds.
                Timeout(42)

                // Header defines a header that should be sent with the
                // request.
                Header("<name>", "<value>")
            })
        })
    })

    // Views is optional and defines one or more views.
    Views(func() {
        
        // SystemLandscapeView defines a System Landscape view.
        SystemLandscapeView("[key]", "[description]", func() {

            // Title of this view.
            Title("<title>")

            // AddDefault adds default elements that are relevant for the
            // specific view:
            //
            //    - System landscape view: adds all software systems and people
            //    - System context view: adds softare system and other related
            //      software systems and people.
            //    - Container view: adds all containers in software system as well
            //      as related software systems and people.
            //    - Component view: adds all components in container as well as
            //      related containers, software systems and people.
            AddDefault()

            // Add given person or element to view. If person or element was
            // already added implictely (e.g. via AddAll()) then overrides how
            // the person or element is rendered.
            Add(PersonOrElement, func() {

                // Set explicit coordinates for where to render person or
                // element.
                Coord(X, Y)

                // Do not render relationships when rendering person or element.
                NoRelationships()
            })

            // Add given relationship to view. If relationship was already added
            // implictely (e.g. via AddAll()) then overrides how the
            // relationship is rendered.
            Add(Source, Destination, func() {

                // Vertices lists the x and y coordinate of the vertices used to
                // render the relationship. The number of arguments must be even.
                Vertices(10, 20, 10, 40)
    
                // Routing algorithm used when rendering relationship, one of
                // RoutingDirect, RoutingCurved or RoutingOrthogonal.
                Routing(RoutingOrthogonal)
    
                // Position of annotation along line; 0 (start) to 100 (end).
                Position(50)
            })

            // Add all elements and people  in scope to view.
            AddAll()

            // Add all elements that are directly connected to given person
            // or element.
            AddNeighbors(PersonOrElement)

            // Remove given element or person from view.
            Remove(ElementOrPerson)

            // Remove given relationship from view.
            Remove(Source, Destination)

            // Remove elements and relationships with given tag.
            Remove("<tag>")

            // Remove all elements and people that cannot be reached by
            // traversing the graph of relationships starting with given element
            // or person.
            RemoveUnreachable(ElementOrPerson)

            // Remove all elements that have no relationships to other elements.
            RemoveUnrelated()

            // AutoLayout enables automatic layout mode for the diagram. The
            // first argument indicates the rank direction, it must be one of
            // RankTopBottom, RankBottomTop, RankLeftRight or RankRightLeft.
            AutoLayout(RankTopBottom, func() {

                // Separation between ranks in pixels
                RankSeparation(200)

                // Separation between nodes in the same rank in pixels
                NodeSeparation(200) 

                // Separation between edges in pixels
                EdgeSeparation(10) 

                // Create vertices during automatic layout.
                Vertices()
            }) 

            // AnimationStep defines an animation step consisting of the
            // specified elements.
            AnimationStep(Element, Element/*, ...*/)
            
            // PaperSize defines the paper size that should be used to render
            // the view. The possible values for the argument follow the
            // patterns SizeA[0-6][Portrait|Landscape], SizeLetter[Portrait|Landscape]
            // or SizeLegal[Portrait_Landscape]. Alternatively the argument may be
            // one of SizeSlide4X3, SizeSlide16X9 or SizeSlide16X10.
            PaperSize(SizeSlide4X3)

            // Make enterprise boundary visible to differentiate internal
            // elements from external elements on the resulting diagram.
            EnterpriseBoundaryVisible()
        })

        SystemContextView(SoftwareSystem, "[key]", "[description]", func() {
            // ... same usage as SystemLandscapeView.
        })
        
        ContainerView(SoftwareSystem, "[key]", "[description]", func() {
            // ... same usage as SystemLandscapeView without EnterpriseBoundaryVisible.

            // All all containers in software system to view.
            AddContainers()

            // Make software system boundaries visible for "external" containers
            // (those outside the software system in scope).
            SystemBoundariesVisible()
        })

        ComponentView(Container, "[key]", "[description]", func() {
            // ... same usage as SystemLandscapeView without EnterpriseBoundaryVisible.

            // All all containers in software system to view.
            AddContainers()

            // All all components in container to view.
            AddComponents()

            // Make container boundaries visible for "external" components
            // (those outside the container in scope).
            ContainerBoundariesVisible()
        })

        // FilteredView defines a Filtered view on top of the specified view.
        // The given view must be a System Landscape, System Context, Container,
        // or Component view on which this filtered view should be based.
        FilteredView(View, func() {
            // Set of tags to include or exclude (if Exclude() is used)
            // elements/relationships when rendering this filtered view.
            FilterTag("<tag>", "[tag]") // as many as needed

            // Exclude elements and relationships with the given tags instead of
            // including.
            Exclude()
        }) 

        // DynamicView defines a Dynamic view for the specified scope. The
        // first argument defines the scope of the view, and therefore what can
        // be added to the view, as follows: 
        //
        //   * Global scope: People and software systems.
        //   * Software system scope: People, other software systems, and
        //     containers belonging to the software system.
        //   * Container scope: People, other software systems, other
        //     containers, and components belonging to the container.
        DynamicView(Global, "[key]", "[description]", func() {

            // Title of this view.
            Title("<title>")

            // AutoLayout enables automatic layout mode for the diagram. The
            // first argument indicates the rank direction, it must be one of
            // RankTopBottom, RankBottomTop, RankLeftRight or RankRightLeft.
            AutoLayout(RankTopBottom, func() {

                // Separation between ranks in pixels
                RankSeparation(200)

                // Separation between nodes in the same rank in pixels
                NodeSeparation(200) 

                // Separation between edges in pixels
                EdgeSeparation(10) 

                // Create vertices during automatic layout.
                Vertices()
            }) 

            // PaperSize defines the paper size that should be used to render
            // the view. The possible values for the argument follow the
            // patterns SizeA[0-6][Portrait|Landscape], SizeLetter[Portrait|Landscape]
            // or SizeLegal[Portrait_Landscape]. Alternatively the argument may be
            // one of SizeSlide4X3, SizeSlide16X9 or SizeSlide16X10.
            PaperSize(SizeSlide4X3)

            // Set of relationships that make up dynamic diagram.
            Add(Source, Destination, func() {

                // Vertices lists the x and y coordinate of the vertices used to
                // render the relationship. The number of arguments must be even.
                Vertices(10, 20, 10, 40)
    
                // Routing algorithm used when rendering relationship, one of
                // RoutingDirect, RoutingCurved or RoutingOrthogonal.
                Routing(RoutingOrthogonal)
    
                // Position of annotation along line; 0 (start) to 100 (end).
                Position(50)
    
                // Description used in dynamic views.
                Description("<description>")
    
                // Order of relationship in dynamic views, e.g. 1.0, 1.1, 2.0
                Order("<order>")
            })
        })
          
        // DynamicView on software system or container uses the corresponding
        // identifier as first argument.
        DynamicView(SoftwareSystemOrContainer, "[key]", "[description]", func() {
            // see usage above
        })

        // DeploymentView defines a Deployment view for the specified scope and
        // deployment environment. The first argument defines the scope of the
        // view, and the second property defines the deployment environment. The
        // combination of these two arguments determines what can be added to
        // the view, as follows: 
        //  * Global scope: All deployment nodes, infrastructure nodes, and
        //    container instances within the deployment environment.
        //  * Software system scope: All deployment nodes and infrastructure
        //    nodes within the deployment environment. Container instances within
        //    the deployment environment that belong to the software system.
        DeploymentView(Global, "<environment name>", "[key]", "[description]", func() {
            // ... same usage as SystemLandscape without EnterpriseBoundaryVisible.
        })

        // DeploymentView on a software system uses the software system as first
        // argument.
        DeploymentView(SoftwareSystem, "<environment name>", "[key]", "[description]", func() {
            // see usage above
        })

        // Styles is a wrapper for one or more element/relationship styles,
        // which are used when rendering diagrams.
        Styles(func() {

            // ElementStyle defines an element style. All nested properties
            // (shape, icon, etc) are optional, see Structurizr - Notation for
            // more details.
            ElementStyle("<tag>", func() {
                Shape(ShapeBox) // ShapeBox, ShapeRoundedBox, ShapeCircle, ShapeEllipse, ShapeHexagon, ShapeCylinder, ShapePipe, ShapePerson
                                // ShapeRobot, ShapeFolder, ShapeWebBrowser, ShapeMobileDevicePortrait, ShapeMobileDeviceLandscape, ShapeComponent
                Icon("<file>")
                Width(42)
                Height(42)
                Background("#<rrggbb>")
                Color("#<rrggbb>")
                Stroke("#<rrggbb>")
                FontSize(42)
                Boder(BorderSolid) // BorderSolid, BorderDashed, BorderDotted
                Opacity(42) // Between 0 and 100
                Metadata(true)
                Description(true)
            })

            // RelationshipStyle defines a relationship style. All nested
            // properties (thickness, color, etc) are optional, see Structurizr
            // - Notation for more details.
            RelationshipStyle("<tag>", func() {
                Thickness(42)
                Color("#<rrggbb>")
                Dashed(true)
                Routing(RoutingDirect) // RoutingDirect, RoutingOrthogonal, RoutingCurved
                FontSize(42)
                Width(42)
                Position(42) // Between 0 and 100
                Opacity(42)  // Between 0 and 100
            })
        })

        // Theme specifies one or more themes that should be used when
        // rendering diagrams. See Structurizr - Themes for more details.
        Theme("<theme URL>", "[theme URL]") // as many theme URLs as needed

        // Branding defines custom branding that should be used when rendering
        // diagrams and documentation. See Structurizr - Branding for more
        // details.
        Branding(func() {
            Logo("<file>")
            Font("<name>", "[url]")
        })
    })
})
```
