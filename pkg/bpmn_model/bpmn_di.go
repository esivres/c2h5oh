package bpmn_model

// --- Constants ---

const (
	// DC elements
	DC_ELEMENT_BOUNDS = "Bounds"
	DC_ELEMENT_POINT  = "Point"

	// DI elements
	DI_ELEMENT_WAYPOINT = "waypoint"

	// BPMNDI elements
	BPMNDI_ELEMENT_DIAGRAM = "BPMNDiagram"
	BPMNDI_ELEMENT_PLANE   = "BPMNPlane"
	BPMNDI_ELEMENT_SHAPE   = "BPMNShape"
	BPMNDI_ELEMENT_EDGE    = "BPMNEdge"
	BPMNDI_ELEMENT_LABEL   = "BPMNLabel"

	// Attributes
	BPMNDI_ATTRIBUTE_BPMN_ELEMENT      = "bpmnElement"
	BPMNDI_ATTRIBUTE_IS_HORIZONTAL     = "isHorizontal"
	BPMNDI_ATTRIBUTE_IS_EXPANDED       = "isExpanded"
	BPMNDI_ATTRIBUTE_IS_MARKER_VISIBLE = "isMarkerVisible"
)

// --- Interfaces ---

// Bounds represents a rectangle (DC namespace).
type Bounds interface {
	ModelElementInstance
	GetX() float64
	SetX(x float64)
	GetY() float64
	SetY(y float64)
	GetWidth() float64
	SetWidth(w float64)
	GetHeight() float64
	SetHeight(h float64)
}

// Waypoint represents a point on an edge (DI namespace).
type Waypoint interface {
	ModelElementInstance
	GetX() float64
	SetX(x float64)
	GetY() float64
	SetY(y float64)
}

// BpmnDiagram is the root DI element.
type BpmnDiagram interface {
	ModelElementInstance
	implBpmnDiagram()
	GetId() string
	SetId(id string)
	GetPlane() BpmnPlane
}

// BpmnPlane contains shapes and edges.
type BpmnPlane interface {
	ModelElementInstance
	implBpmnPlane()
	GetId() string
	SetId(id string)
	GetBpmnElement() string
	SetBpmnElement(ref string)
	GetShapes() []BpmnShape
	GetEdges() []BpmnEdge
}

// BpmnShape represents a visual shape for a BPMN element.
type BpmnShape interface {
	ModelElementInstance
	implBpmnShape()
	GetId() string
	SetId(id string)
	GetBpmnElement() string
	SetBpmnElement(ref string)
	GetBounds() Bounds
	SetBounds(b Bounds)
	IsHorizontal() bool
	SetIsHorizontal(v bool)
	IsExpanded() bool
	SetIsExpanded(v bool)
	IsMarkerVisible() bool
	SetIsMarkerVisible(v bool)
}

// BpmnEdge represents a visual edge (connection) for a BPMN element.
type BpmnEdge interface {
	ModelElementInstance
	implBpmnEdge()
	GetId() string
	SetId(id string)
	GetBpmnElement() string
	SetBpmnElement(ref string)
	GetWaypoints() []Waypoint
	AddWaypoint(x, y float64)
}

// BpmnLabel represents a label position.
type BpmnLabel interface {
	ModelElementInstance
	implBpmnLabel()
	GetBounds() Bounds
}

// --- Attribute descriptors ---

var (
	AttrBoundsX      *Attribute[float64]
	AttrBoundsY      *Attribute[float64]
	AttrBoundsWidth  *Attribute[float64]
	AttrBoundsHeight *Attribute[float64]

	AttrWaypointX *Attribute[float64]
	AttrWaypointY *Attribute[float64]

	AttrDiagramId        *Attribute[string]
	AttrPlaneId          *Attribute[string]
	AttrPlaneBpmnElement *Attribute[string]

	AttrShapeId              *Attribute[string]
	AttrShapeBpmnElement     *Attribute[string]
	AttrShapeIsHorizontal    *Attribute[bool]
	AttrShapeIsExpanded      *Attribute[bool]
	AttrShapeIsMarkerVisible *Attribute[bool]

	AttrEdgeId          *Attribute[string]
	AttrEdgeBpmnElement *Attribute[string]

	AttrLabelId *Attribute[string]
)

// --- Implementations ---

type BoundsImpl struct{ BaseInstance }

func (b *BoundsImpl) GetX() float64       { return AttrBoundsX.Get(b) }
func (b *BoundsImpl) SetX(x float64)      { AttrBoundsX.Set(b, x) }
func (b *BoundsImpl) GetY() float64       { return AttrBoundsY.Get(b) }
func (b *BoundsImpl) SetY(y float64)      { AttrBoundsY.Set(b, y) }
func (b *BoundsImpl) GetWidth() float64   { return AttrBoundsWidth.Get(b) }
func (b *BoundsImpl) SetWidth(w float64)  { AttrBoundsWidth.Set(b, w) }
func (b *BoundsImpl) GetHeight() float64  { return AttrBoundsHeight.Get(b) }
func (b *BoundsImpl) SetHeight(h float64) { AttrBoundsHeight.Set(b, h) }

type WaypointImpl struct{ BaseInstance }

func (w *WaypointImpl) GetX() float64  { return AttrWaypointX.Get(w) }
func (w *WaypointImpl) SetX(x float64) { AttrWaypointX.Set(w, x) }
func (w *WaypointImpl) GetY() float64  { return AttrWaypointY.Get(w) }
func (w *WaypointImpl) SetY(y float64) { AttrWaypointY.Set(w, y) }

type BpmnDiagramImpl struct{ BaseInstance }

func (*BpmnDiagramImpl) implBpmnDiagram()  {}
func (d *BpmnDiagramImpl) GetId() string   { return AttrDiagramId.Get(d) }
func (d *BpmnDiagramImpl) SetId(id string) { AttrDiagramId.Set(d, id) }
func (d *BpmnDiagramImpl) GetPlane() BpmnPlane {
	mi := d.GetModelInstance()
	if mi == nil {
		return nil
	}
	children := d.DomElement.GetChildElementsByNS(BPMNDI_NS, BPMNDI_ELEMENT_PLANE)
	if len(children) == 0 {
		return nil
	}
	if inst := mi.GetElementByDom(children[0]); inst != nil {
		if p, ok := inst.(BpmnPlane); ok {
			return p
		}
	}
	return nil
}

type BpmnPlaneImpl struct{ BaseInstance }

func (*BpmnPlaneImpl) implBpmnPlane()              {}
func (p *BpmnPlaneImpl) GetId() string             { return AttrPlaneId.Get(p) }
func (p *BpmnPlaneImpl) SetId(id string)           { AttrPlaneId.Set(p, id) }
func (p *BpmnPlaneImpl) GetBpmnElement() string    { return AttrPlaneBpmnElement.Get(p) }
func (p *BpmnPlaneImpl) SetBpmnElement(ref string) { AttrPlaneBpmnElement.Set(p, ref) }
func (p *BpmnPlaneImpl) GetShapes() []BpmnShape {
	return getDiChildren[BpmnShape](p, BPMNDI_NS, BPMNDI_ELEMENT_SHAPE)
}
func (p *BpmnPlaneImpl) GetEdges() []BpmnEdge {
	return getDiChildren[BpmnEdge](p, BPMNDI_NS, BPMNDI_ELEMENT_EDGE)
}

type BpmnShapeImpl struct{ BaseInstance }

func (*BpmnShapeImpl) implBpmnShape()              {}
func (s *BpmnShapeImpl) GetId() string             { return AttrShapeId.Get(s) }
func (s *BpmnShapeImpl) SetId(id string)           { AttrShapeId.Set(s, id) }
func (s *BpmnShapeImpl) GetBpmnElement() string    { return AttrShapeBpmnElement.Get(s) }
func (s *BpmnShapeImpl) SetBpmnElement(ref string) { AttrShapeBpmnElement.Set(s, ref) }
func (s *BpmnShapeImpl) IsHorizontal() bool        { return AttrShapeIsHorizontal.Get(s) }
func (s *BpmnShapeImpl) SetIsHorizontal(v bool)    { AttrShapeIsHorizontal.Set(s, v) }
func (s *BpmnShapeImpl) IsExpanded() bool          { return AttrShapeIsExpanded.Get(s) }
func (s *BpmnShapeImpl) SetIsExpanded(v bool)      { AttrShapeIsExpanded.Set(s, v) }
func (s *BpmnShapeImpl) IsMarkerVisible() bool     { return AttrShapeIsMarkerVisible.Get(s) }
func (s *BpmnShapeImpl) SetIsMarkerVisible(v bool) { AttrShapeIsMarkerVisible.Set(s, v) }
func (s *BpmnShapeImpl) GetBounds() Bounds {
	mi := s.GetModelInstance()
	if mi == nil {
		return nil
	}
	children := s.DomElement.GetChildElementsByNS(DC_NS, DC_ELEMENT_BOUNDS)
	if len(children) == 0 {
		return nil
	}
	if inst := mi.GetElementByDom(children[0]); inst != nil {
		if b, ok := inst.(Bounds); ok {
			return b
		}
	}
	return nil
}
func (s *BpmnShapeImpl) SetBounds(b Bounds) {
	// Remove existing
	for _, child := range s.DomElement.GetChildElementsByNS(DC_NS, DC_ELEMENT_BOUNDS) {
		s.DomElement.RemoveChild(child)
	}
	if b != nil {
		s.DomElement.AppendChild(b.GetDomElement())
	}
}

type BpmnEdgeImpl struct{ BaseInstance }

func (*BpmnEdgeImpl) implBpmnEdge()               {}
func (e *BpmnEdgeImpl) GetId() string             { return AttrEdgeId.Get(e) }
func (e *BpmnEdgeImpl) SetId(id string)           { AttrEdgeId.Set(e, id) }
func (e *BpmnEdgeImpl) GetBpmnElement() string    { return AttrEdgeBpmnElement.Get(e) }
func (e *BpmnEdgeImpl) SetBpmnElement(ref string) { AttrEdgeBpmnElement.Set(e, ref) }
func (e *BpmnEdgeImpl) GetWaypoints() []Waypoint {
	return getDiChildren[Waypoint](e, DI_NS, DI_ELEMENT_WAYPOINT)
}
func (e *BpmnEdgeImpl) AddWaypoint(x, y float64) {
	mi := e.GetModelInstance()
	wpType := bpmnModel.GetTypeByQName(DI_NS, DI_ELEMENT_WAYPOINT)
	inst, _ := mi.NewInstance(wpType)
	wp := inst.(Waypoint)
	wp.SetX(x)
	wp.SetY(y)
	e.DomElement.AppendChild(wp.GetDomElement())
}

type BpmnLabelImpl struct{ BaseInstance }

func (*BpmnLabelImpl) implBpmnLabel() {}
func (l *BpmnLabelImpl) GetBounds() Bounds {
	mi := l.GetModelInstance()
	if mi == nil {
		return nil
	}
	children := l.DomElement.GetChildElementsByNS(DC_NS, DC_ELEMENT_BOUNDS)
	if len(children) == 0 {
		return nil
	}
	if inst := mi.GetElementByDom(children[0]); inst != nil {
		if b, ok := inst.(Bounds); ok {
			return b
		}
	}
	return nil
}

// --- Helper ---

func getDiChildren[T ModelElementInstance](parent ModelElementInstance, ns, localName string) []T {
	mi := parent.GetModelInstance()
	if mi == nil {
		return nil
	}
	var result []T
	for _, child := range parent.GetDomElement().GetChildElementsByNS(ns, localName) {
		if inst := mi.GetElementByDom(child); inst != nil {
			if typed, ok := inst.(T); ok {
				result = append(result, typed)
			}
		}
	}
	return result
}

// --- Registration ---

func registerDiTypes(mb *ModelBuilder) {
	registerBoundsType(mb)
	registerWaypointType(mb)
	registerBpmnDiagramType(mb)
	registerBpmnPlaneType(mb)
	registerBpmnShapeType(mb)
	registerBpmnEdgeType(mb)
	registerBpmnLabelType(mb)
}

func registerBoundsType(mb *ModelBuilder) {
	tb := mb.DefineType((*Bounds)(nil), DC_ELEMENT_BOUNDS).
		Namespace(DC_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &BoundsImpl{NewBaseInstance(ctx)}
		})
	AttrBoundsX = tb.Float64Attribute("x").Build()
	AttrBoundsY = tb.Float64Attribute("y").Build()
	AttrBoundsWidth = tb.Float64Attribute("width").Build()
	AttrBoundsHeight = tb.Float64Attribute("height").Build()
}

func registerWaypointType(mb *ModelBuilder) {
	tb := mb.DefineType((*Waypoint)(nil), DI_ELEMENT_WAYPOINT).
		Namespace(DI_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &WaypointImpl{NewBaseInstance(ctx)}
		})
	AttrWaypointX = tb.Float64Attribute("x").Build()
	AttrWaypointY = tb.Float64Attribute("y").Build()
}

func registerBpmnDiagramType(mb *ModelBuilder) {
	tb := mb.DefineType((*BpmnDiagram)(nil), BPMNDI_ELEMENT_DIAGRAM).
		Namespace(BPMNDI_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &BpmnDiagramImpl{NewBaseInstance(ctx)}
		})
	AttrDiagramId = tb.StringAttribute("id").IdAttribute().Build()
}

func registerBpmnPlaneType(mb *ModelBuilder) {
	tb := mb.DefineType((*BpmnPlane)(nil), BPMNDI_ELEMENT_PLANE).
		Namespace(BPMNDI_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &BpmnPlaneImpl{NewBaseInstance(ctx)}
		})
	AttrPlaneId = tb.StringAttribute("id").IdAttribute().Build()
	AttrPlaneBpmnElement = tb.StringAttribute(BPMNDI_ATTRIBUTE_BPMN_ELEMENT).Build()
}

func registerBpmnShapeType(mb *ModelBuilder) {
	tb := mb.DefineType((*BpmnShape)(nil), BPMNDI_ELEMENT_SHAPE).
		Namespace(BPMNDI_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &BpmnShapeImpl{NewBaseInstance(ctx)}
		})
	AttrShapeId = tb.StringAttribute("id").IdAttribute().Build()
	AttrShapeBpmnElement = tb.StringAttribute(BPMNDI_ATTRIBUTE_BPMN_ELEMENT).Build()
	AttrShapeIsHorizontal = tb.BoolAttribute(BPMNDI_ATTRIBUTE_IS_HORIZONTAL).Build()
	AttrShapeIsExpanded = tb.BoolAttribute(BPMNDI_ATTRIBUTE_IS_EXPANDED).Build()
	AttrShapeIsMarkerVisible = tb.BoolAttribute(BPMNDI_ATTRIBUTE_IS_MARKER_VISIBLE).Build()
}

func registerBpmnEdgeType(mb *ModelBuilder) {
	tb := mb.DefineType((*BpmnEdge)(nil), BPMNDI_ELEMENT_EDGE).
		Namespace(BPMNDI_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &BpmnEdgeImpl{NewBaseInstance(ctx)}
		})
	AttrEdgeId = tb.StringAttribute("id").IdAttribute().Build()
	AttrEdgeBpmnElement = tb.StringAttribute(BPMNDI_ATTRIBUTE_BPMN_ELEMENT).Build()
}

func registerBpmnLabelType(mb *ModelBuilder) {
	mb.DefineType((*BpmnLabel)(nil), BPMNDI_ELEMENT_LABEL).
		Namespace(BPMNDI_NS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &BpmnLabelImpl{NewBaseInstance(ctx)}
		})
}

// --- DI Generation for builders ---

// Standard BPMN element sizes
const (
	diEventWidth    = 36.0
	diEventHeight   = 36.0
	diTaskWidth     = 100.0
	diTaskHeight    = 80.0
	diGatewayWidth  = 50.0
	diGatewayHeight = 50.0
	diSpacingX      = 50.0
	diStartX        = 179.0
	diCenterY       = 97.0
)

// diState tracks position during DI generation.
type diState struct {
	plane  BpmnPlane
	shapes map[string]BpmnShape
	nextX  float64
}

func newDiState(plane BpmnPlane) *diState {
	return &diState{
		nextX:  diStartX,
		plane:  plane,
		shapes: make(map[string]BpmnShape),
	}
}

func (ds *diState) addShape(mi *ModelInstance, elementId string, width, height float64) BpmnShape {
	shapeType := bpmnModel.GetTypeByQName(BPMNDI_NS, BPMNDI_ELEMENT_SHAPE)
	shapeInst, _ := mi.NewInstance(shapeType)
	shape := shapeInst.(BpmnShape)
	shape.SetId(elementId + "_di")
	shape.SetBpmnElement(elementId)

	boundsType := bpmnModel.GetTypeByQName(DC_NS, DC_ELEMENT_BOUNDS)
	boundsInst, _ := mi.NewInstance(boundsType)
	bounds := boundsInst.(Bounds)
	bounds.SetX(ds.nextX)
	bounds.SetY(diCenterY - height/2)
	bounds.SetWidth(width)
	bounds.SetHeight(height)

	shape.GetDomElement().AppendChild(bounds.GetDomElement())
	ds.plane.GetDomElement().AppendChild(shape.GetDomElement())
	ds.shapes[elementId] = shape

	ds.nextX += width + diSpacingX
	return shape
}

func (ds *diState) addEdge(mi *ModelInstance, flowId, sourceId, targetId string) {
	srcShape := ds.shapes[sourceId]
	tgtShape := ds.shapes[targetId]
	if srcShape == nil || tgtShape == nil {
		return
	}

	edgeType := bpmnModel.GetTypeByQName(BPMNDI_NS, BPMNDI_ELEMENT_EDGE)
	edgeInst, _ := mi.NewInstance(edgeType)
	edge := edgeInst.(BpmnEdge)
	edge.SetId(flowId + "_di")
	edge.SetBpmnElement(flowId)

	srcBounds := srcShape.GetBounds()
	tgtBounds := tgtShape.GetBounds()
	if srcBounds != nil && tgtBounds != nil {
		// Source: right center
		edge.AddWaypoint(srcBounds.GetX()+srcBounds.GetWidth(), srcBounds.GetY()+srcBounds.GetHeight()/2)
		// Target: left center
		edge.AddWaypoint(tgtBounds.GetX(), tgtBounds.GetY()+tgtBounds.GetHeight()/2)
	}

	ds.plane.GetDomElement().AppendChild(edge.GetDomElement())
}

// generateDI creates BPMNDiagram, BPMNPlane, shapes and edges for the process.
func generateDI(bmi *BpmnModelInstance, proc Process) {
	mi := bmi.ModelInstance
	def := bmi.GetDefinitions()

	// Create diagram
	diagType := bpmnModel.GetTypeByQName(BPMNDI_NS, BPMNDI_ELEMENT_DIAGRAM)
	diagInst, _ := mi.NewInstance(diagType)
	diag := diagInst.(BpmnDiagram)
	diag.SetId("BPMNDiagram_1")

	// Create plane
	planeType := bpmnModel.GetTypeByQName(BPMNDI_NS, BPMNDI_ELEMENT_PLANE)
	planeInst, _ := mi.NewInstance(planeType)
	plane := planeInst.(BpmnPlane)
	plane.SetId("BPMNPlane_1")
	plane.SetBpmnElement(proc.GetId())

	diag.GetDomElement().AppendChild(plane.GetDomElement())
	def.GetDomElement().AppendChild(diag.GetDomElement())

	ds := newDiState(plane)

	// Create shapes for flow elements (in order)
	for _, fe := range proc.GetFlowElements() {
		if sf, ok := fe.(SequenceFlow); ok {
			_ = sf // handle edges later
			continue
		}
		width, height := elementSize(fe)
		ds.addShape(mi, fe.GetId(), width, height)
	}

	// Create edges for sequence flows
	for _, fe := range proc.GetFlowElements() {
		if sf, ok := fe.(SequenceFlow); ok {
			ds.addEdge(mi, sf.GetId(), sf.GetSourceRef(), sf.GetTargetRef())
		}
	}
}

func elementSize(fe FlowElement) (width, height float64) {
	switch fe.(type) {
	case StartEvent, EndEvent, IntermediateCatchEvent, IntermediateThrowEvent, BoundaryEvent:
		return diEventWidth, diEventHeight
	case ExclusiveGateway, ParallelGateway, InclusiveGateway, EventBasedGateway, ComplexGateway, Gateway:
		return diGatewayWidth, diGatewayHeight
	default:
		return diTaskWidth, diTaskHeight
	}
}
