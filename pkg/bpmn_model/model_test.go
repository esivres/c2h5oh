package bpmn_model

import (
	"testing"

	xmlm "github.com/esivres/c2h5oh/pkg/bpmn_model/xml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test types: Animal model ---

const testNS = "http://example.com/animals"

// AnimalInstance is the interface for all animal elements.
type AnimalInstance interface {
	ModelElementInstance
	GetName() string
	SetName(name string)
}

// BirdInstance extends AnimalInstance.
type BirdInstance interface {
	AnimalInstance
	GetWingspan() int
	SetWingspan(ws int)
}

// NestInstance is a child element of Bird.
type NestInstance interface {
	ModelElementInstance
	GetMaterial() string
	SetMaterial(material string)
}

// --- Implementations ---

var (
	animalIdAttr   *Attribute[string]
	animalNameAttr *Attribute[string]
	birdWingspan   *Attribute[int]
	nestMaterial   *Attribute[string]
)

type AnimalImpl struct {
	BaseInstance
}

func (a *AnimalImpl) GetName() string     { return animalNameAttr.Get(a) }
func (a *AnimalImpl) SetName(name string) { animalNameAttr.Set(a, name) }

type BirdImpl struct {
	AnimalImpl
}

func (b *BirdImpl) GetWingspan() int   { return birdWingspan.Get(b) }
func (b *BirdImpl) SetWingspan(ws int) { birdWingspan.Set(b, ws) }

type NestImpl struct {
	BaseInstance
}

func (n *NestImpl) GetMaterial() string         { return nestMaterial.Get(n) }
func (n *NestImpl) SetMaterial(material string) { nestMaterial.Set(n, material) }

// --- Build the test model ---

func buildAnimalModel() *Model {
	mb := NewModelBuilder("animals")

	// Define Animal type (abstract)
	animalTB := mb.DefineType((*AnimalInstance)(nil), "animal").
		Namespace(testNS).
		AbstractType()

	animalIdAttr = animalTB.StringAttribute("id").IdAttribute().Required().Build()
	animalNameAttr = animalTB.StringAttribute("name").Build()
	animalTB.AddChildSpec("nest", testNS, 0, -1)

	// Define Bird type
	birdTB := mb.DefineType((*BirdInstance)(nil), "bird").
		Namespace(testNS).
		ExtendsType((*AnimalInstance)(nil)).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &BirdImpl{AnimalImpl{NewBaseInstance(ctx)}}
		})

	birdWingspan = birdTB.IntAttribute("wingspan").DefaultValue(0).Build()

	// Define Nest type
	nestTB := mb.DefineType((*NestInstance)(nil), "nest").
		Namespace(testNS).
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &NestImpl{NewBaseInstance(ctx)}
		})

	nestMaterial = nestTB.StringAttribute("material").Build()
	_ = nestTB

	return mb.Build()
}

// --- Tests ---

func TestModelBuilder(t *testing.T) {
	model := buildAnimalModel()

	assert.Equal(t, "animals", model.Name())

	// Check type lookup by QName
	animalType := model.GetTypeByQName(testNS, "animal")
	require.NotNil(t, animalType)
	assert.Equal(t, "animal", animalType.TypeName)
	assert.True(t, animalType.IsAbstract)

	birdType := model.GetTypeByQName(testNS, "bird")
	require.NotNil(t, birdType)
	assert.Equal(t, "bird", birdType.TypeName)
	assert.False(t, birdType.IsAbstract)

	// Check type hierarchy
	assert.Equal(t, animalType, birdType.BaseType)
	assert.Contains(t, animalType.ExtendingTypes, birdType)
}

func TestModelInstance_CreateAndQuery(t *testing.T) {
	model := buildAnimalModel()

	doc := xmlm.NewDocument()
	doc.RegisterNamespace("a", testNS)
	mi := NewModelInstance(model, doc)

	birdType := model.GetTypeByQName(testNS, "bird")
	require.NotNil(t, birdType)

	// Create a bird
	inst, err := mi.NewInstance(birdType)
	require.NoError(t, err)

	bird, ok := inst.(BirdInstance)
	require.True(t, ok)

	bird.SetName("Sparrow")
	bird.SetWingspan(25)

	assert.Equal(t, "Sparrow", bird.GetName())
	assert.Equal(t, 25, bird.GetWingspan())

	// Set as root
	mi.SetDocumentElement(bird)

	// Query by DOM
	domElem := bird.GetDomElement()
	found := mi.GetElementByDom(domElem)
	assert.Equal(t, bird, found)
}

func TestAttribute_IdRegistration(t *testing.T) {
	model := buildAnimalModel()

	doc := xmlm.NewDocument()
	doc.RegisterNamespace("a", testNS)
	mi := NewModelInstance(model, doc)

	birdType := model.GetTypeByQName(testNS, "bird")
	inst, err := mi.NewInstance(birdType)
	require.NoError(t, err)

	bird := inst.(BirdInstance)
	animalIdAttr.Set(bird, "bird-1")

	// Should be findable by ID
	found := mi.GetElementById("bird-1")
	require.NotNil(t, found)
	assert.Equal(t, bird, found)

	// Generic helper
	typedBird, ok := GetTypedElementById[BirdInstance](mi, "bird-1")
	require.True(t, ok)
	assert.Equal(t, "bird-1", animalIdAttr.GetRaw(typedBird))
}

func TestAttribute_DefaultValue(t *testing.T) {
	model := buildAnimalModel()

	doc := xmlm.NewDocument()
	doc.RegisterNamespace("a", testNS)
	mi := NewModelInstance(model, doc)

	birdType := model.GetTypeByQName(testNS, "bird")
	inst, err := mi.NewInstance(birdType)
	require.NoError(t, err)

	bird := inst.(BirdInstance)
	// Wingspan has default value 0
	assert.Equal(t, 0, bird.GetWingspan())
}

func TestModelInstance_Serialize(t *testing.T) {
	model := buildAnimalModel()

	doc := xmlm.NewDocument()
	doc.RegisterNamespace("a", testNS)
	mi := NewModelInstance(model, doc)

	birdType := model.GetTypeByQName(testNS, "bird")
	inst, err := mi.NewInstance(birdType)
	require.NoError(t, err)

	bird := inst.(BirdInstance)
	animalIdAttr.Set(bird, "bird-1")
	bird.SetName("Eagle")
	bird.SetWingspan(200)

	mi.SetDocumentElement(bird)

	// Serialize
	output, err := doc.WriteToString()
	require.NoError(t, err)

	assert.Contains(t, output, "bird")
	assert.Contains(t, output, "bird-1")
	assert.Contains(t, output, "Eagle")
	assert.Contains(t, output, "200")
}

func TestChildElementCollection(t *testing.T) {
	model := buildAnimalModel()

	doc := xmlm.NewDocument()
	doc.RegisterNamespace("a", testNS)
	mi := NewModelInstance(model, doc)

	birdType := model.GetTypeByQName(testNS, "bird")
	nestType := model.GetTypeByQName(testNS, "nest")

	birdInst, err := mi.NewInstance(birdType)
	require.NoError(t, err)
	bird := birdInst.(BirdInstance)
	animalIdAttr.Set(bird, "bird-1")

	// Create nests
	nest1Inst, err := mi.NewInstance(nestType)
	require.NoError(t, err)
	nest1 := nest1Inst.(NestInstance)
	nest1.SetMaterial("twigs")

	nest2Inst, err := mi.NewInstance(nestType)
	require.NoError(t, err)
	nest2 := nest2Inst.(NestInstance)
	nest2.SetMaterial("leaves")

	// Add nests as children
	nestCollection := &ChildElementCollection[NestInstance]{
		Spec: &ChildSpec{
			ChildTypeName: "nest",
			ChildTypeNS:   testNS,
			MaxOccurs:     -1,
		},
	}

	nestCollection.Add(bird, nest1)
	nestCollection.Add(bird, nest2)

	// Query children
	nests := nestCollection.Get(bird)
	require.Len(t, nests, 2)
	assert.Equal(t, "twigs", nests[0].GetMaterial())
	assert.Equal(t, "leaves", nests[1].GetMaterial())
}

func TestReference(t *testing.T) {
	model := buildAnimalModel()

	doc := xmlm.NewDocument()
	doc.RegisterNamespace("a", testNS)
	mi := NewModelInstance(model, doc)

	birdType := model.GetTypeByQName(testNS, "bird")

	bird1Inst, err := mi.NewInstance(birdType)
	require.NoError(t, err)
	bird1 := bird1Inst.(BirdInstance)
	animalIdAttr.Set(bird1, "bird-1")

	bird2Inst, err := mi.NewInstance(birdType)
	require.NoError(t, err)
	bird2 := bird2Inst.(BirdInstance)
	animalIdAttr.Set(bird2, "bird-2")

	// Create a reference from bird2 to bird1
	ref := &Reference[BirdInstance]{
		AttributeName: "mateRef",
	}

	ref.Set(bird2, bird1)
	assert.Equal(t, "bird-1", ref.GetId(bird2))

	resolved, ok := ref.Get(bird2)
	require.True(t, ok)
	assert.Equal(t, bird1, resolved)
}

func TestAlternativeNamespace(t *testing.T) {
	model := buildAnimalModel()
	model.RegisterAlternativeNamespace("http://old-example.com/animals", testNS)

	// Should resolve via alternative namespace
	found := model.GetTypeByQName("http://old-example.com/animals", "bird")
	require.NotNil(t, found)
	assert.Equal(t, "bird", found.TypeName)

	// Direct lookup still works
	found2 := model.GetTypeByQName(testNS, "bird")
	require.NotNil(t, found2)
	assert.Equal(t, found, found2)
}

func TestResolveElements_FromXML(t *testing.T) {
	model := buildAnimalModel()

	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<a:bird xmlns:a="http://example.com/animals" id="eagle-1" name="Eagle" wingspan="220">
  <a:nest material="branches"/>
  <a:nest material="feathers"/>
</a:bird>`

	doc, err := xmlm.ParseDocumentString(xmlStr)
	require.NoError(t, err)

	mi := NewModelInstance(model, doc)
	err = mi.ResolveElements()
	require.NoError(t, err)

	// Find bird by ID
	birdInst := mi.GetElementById("eagle-1")
	require.NotNil(t, birdInst)

	bird, ok := birdInst.(BirdInstance)
	require.True(t, ok)
	assert.Equal(t, "Eagle", bird.GetName())
	assert.Equal(t, 220, bird.GetWingspan())

	// Find nests
	nestCollection := &ChildElementCollection[NestInstance]{
		Spec: &ChildSpec{
			ChildTypeName: "nest",
			ChildTypeNS:   testNS,
			MaxOccurs:     -1,
		},
	}
	nests := nestCollection.Get(bird)
	require.Len(t, nests, 2)
	assert.Equal(t, "branches", nests[0].GetMaterial())
	assert.Equal(t, "feathers", nests[1].GetMaterial())
}

func TestGetTypedElements(t *testing.T) {
	model := buildAnimalModel()

	doc := xmlm.NewDocument()
	doc.RegisterNamespace("a", testNS)
	mi := NewModelInstance(model, doc)

	birdType := model.GetTypeByQName(testNS, "bird")
	nestType := model.GetTypeByQName(testNS, "nest")

	bird1, _ := mi.NewInstance(birdType)
	bird2, _ := mi.NewInstance(birdType)
	nest1, _ := mi.NewInstance(nestType)

	_ = bird1
	_ = bird2
	_ = nest1

	birds := GetTypedElements[BirdInstance](mi)
	assert.Len(t, birds, 2)

	nests := GetTypedElements[NestInstance](mi)
	assert.Len(t, nests, 1)
}

func TestModelElementType_IsTypeOf(t *testing.T) {
	model := buildAnimalModel()

	doc := xmlm.NewDocument()
	doc.RegisterNamespace("a", testNS)
	mi := NewModelInstance(model, doc)

	birdType := model.GetTypeByQName(testNS, "bird")
	animalType := model.GetTypeByQName(testNS, "animal")

	inst, err := mi.NewInstance(birdType)
	require.NoError(t, err)

	assert.True(t, birdType.IsTypeOf(inst))
	assert.True(t, animalType.IsTypeOf(inst))
}

func TestAbstractType_CannotInstantiate(t *testing.T) {
	model := buildAnimalModel()

	doc := xmlm.NewDocument()
	doc.RegisterNamespace("a", testNS)
	mi := NewModelInstance(model, doc)

	animalType := model.GetTypeByQName(testNS, "animal")
	_, err := mi.NewInstance(animalType)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "abstract")
}

func TestIdsReference(t *testing.T) {
	model := buildAnimalModel()

	doc := xmlm.NewDocument()
	doc.RegisterNamespace("a", testNS)
	mi := NewModelInstance(model, doc)

	birdType := model.GetTypeByQName(testNS, "bird")

	bird1, _ := mi.NewInstance(birdType)
	animalIdAttr.Set(bird1, "b1")
	bird2, _ := mi.NewInstance(birdType)
	animalIdAttr.Set(bird2, "b2")
	bird3, _ := mi.NewInstance(birdType)
	animalIdAttr.Set(bird3, "b3")

	// Set multiple IDs reference on bird3
	bird3.GetDomElement().SetAttribute("friendRefs", "b1 b2")

	idsRef := &IdsReference[BirdInstance]{
		AttributeName: "friendRefs",
	}

	ids := idsRef.GetIds(bird3)
	assert.Equal(t, []string{"b1", "b2"}, ids)

	friends := idsRef.Get(bird3)
	require.Len(t, friends, 2)
}

func TestEnumAttribute(t *testing.T) {
	mb := NewModelBuilder("test")
	tb := mb.DefineType((*ModelElementInstance)(nil), "elem").
		Namespace("http://test").
		InstanceProvider(func(ctx *TypeInstanceContext) ModelElementInstance {
			return &BaseInstance{
				DomElement:  ctx.DomElement,
				ModelInst:   ctx.ModelInstance,
				ElementType: ctx.ElementType,
			}
		})

	statusAttr := tb.EnumAttribute("status", "active", "inactive", "pending").
		DefaultValue("pending").
		Build()

	model := mb.Build()

	doc := xmlm.NewDocument()
	doc.RegisterNamespace("t", "http://test")
	mi := NewModelInstance(model, doc)

	elemType := model.GetTypeByQName("http://test", "elem")
	inst, err := mi.NewInstance(elemType)
	require.NoError(t, err)

	// Default value
	assert.Equal(t, "pending", statusAttr.Get(inst))

	// Set valid value
	statusAttr.Set(inst, "active")
	assert.Equal(t, "active", statusAttr.Get(inst))
}
