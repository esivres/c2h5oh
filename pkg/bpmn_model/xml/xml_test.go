package xml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocument(t *testing.T) {
	doc := NewDocument()
	assert.NotNil(t, doc)
	assert.Nil(t, doc.Root())
}

func TestCreateElementAndSetRoot(t *testing.T) {
	doc := NewDocument()
	doc.RegisterNamespace("bpmn", "http://www.omg.org/spec/BPMN/20100524/MODEL")

	root := doc.CreateElement("http://www.omg.org/spec/BPMN/20100524/MODEL", "definitions")
	assert.Equal(t, "definitions", root.LocalName())
	assert.Equal(t, "bpmn", root.Prefix())

	doc.SetRoot(root)
	assert.NotNil(t, doc.Root())
	assert.Equal(t, "definitions", doc.Root().LocalName())
}

func TestNamespaceRegistry(t *testing.T) {
	doc := NewDocument()
	doc.RegisterNamespace("bpmn", "http://www.omg.org/spec/BPMN/20100524/MODEL")
	doc.RegisterNamespace("zeebe", "http://camunda.org/schema/zeebe/1.0")

	assert.Equal(t, "bpmn", doc.LookupPrefix("http://www.omg.org/spec/BPMN/20100524/MODEL"))
	assert.Equal(t, "zeebe", doc.LookupPrefix("http://camunda.org/schema/zeebe/1.0"))
	assert.Equal(t, "http://www.omg.org/spec/BPMN/20100524/MODEL", doc.LookupURI("bpmn"))
	assert.Equal(t, "", doc.LookupPrefix("http://unknown"))
}

func TestAutoGeneratePrefix(t *testing.T) {
	doc := NewDocument()
	prefix := doc.RegisterNamespaceURI("http://example.com/ns1")
	assert.Equal(t, "ns0", prefix)
	assert.Equal(t, "ns0", doc.LookupPrefix("http://example.com/ns1"))

	prefix2 := doc.RegisterNamespaceURI("http://example.com/ns2")
	assert.Equal(t, "ns1", prefix2)
}

func TestElementAttributes(t *testing.T) {
	doc := NewDocument()
	elem := doc.CreateElement("", "process")

	elem.SetAttribute("id", "process1")
	elem.SetAttribute("name", "My Process")

	assert.Equal(t, "process1", elem.GetAttribute("id"))
	assert.Equal(t, "My Process", elem.GetAttribute("name"))
	assert.Equal(t, "", elem.GetAttribute("nonexistent"))
	assert.True(t, elem.HasAttribute("id"))

	elem.RemoveAttribute("name")
	assert.Equal(t, "", elem.GetAttribute("name"))
}

func TestElementAttributesNS(t *testing.T) {
	doc := NewDocument()
	doc.RegisterNamespace("zeebe", "http://camunda.org/schema/zeebe/1.0")

	elem := doc.CreateElement("", "taskDefinition")
	elem.SetAttributeNS("http://camunda.org/schema/zeebe/1.0", "type", "my-worker")

	assert.Equal(t, "my-worker", elem.GetAttributeNS("http://camunda.org/schema/zeebe/1.0", "type"))
	assert.Equal(t, "", elem.GetAttributeNS("http://other", "type"))
}

func TestChildElements(t *testing.T) {
	doc := NewDocument()
	doc.RegisterNamespace("bpmn", "http://www.omg.org/spec/BPMN/20100524/MODEL")

	parent := doc.CreateElement("http://www.omg.org/spec/BPMN/20100524/MODEL", "process")
	child1 := doc.CreateElement("http://www.omg.org/spec/BPMN/20100524/MODEL", "startEvent")
	child1.SetAttribute("id", "start1")
	child2 := doc.CreateElement("http://www.omg.org/spec/BPMN/20100524/MODEL", "endEvent")
	child2.SetAttribute("id", "end1")
	child3 := doc.CreateElement("http://www.omg.org/spec/BPMN/20100524/MODEL", "startEvent")
	child3.SetAttribute("id", "start2")

	parent.AppendChild(child1)
	parent.AppendChild(child2)
	parent.AppendChild(child3)

	children := parent.GetChildElements()
	assert.Len(t, children, 3)

	startEvents := parent.GetChildElementsByNS("http://www.omg.org/spec/BPMN/20100524/MODEL", "startEvent")
	assert.Len(t, startEvents, 2)

	endEvents := parent.GetChildElementsByNS("http://www.omg.org/spec/BPMN/20100524/MODEL", "endEvent")
	assert.Len(t, endEvents, 1)
}

func TestRemoveChild(t *testing.T) {
	doc := NewDocument()
	parent := doc.CreateElement("", "root")
	child1 := doc.CreateElement("", "a")
	child2 := doc.CreateElement("", "b")

	parent.AppendChild(child1)
	parent.AppendChild(child2)
	assert.Len(t, parent.GetChildElements(), 2)

	parent.RemoveChild(child1)
	assert.Len(t, parent.GetChildElements(), 1)
	assert.Equal(t, "b", parent.GetChildElements()[0].LocalName())
}

func TestTextContent(t *testing.T) {
	doc := NewDocument()
	elem := doc.CreateElement("", "expression")
	elem.SetTextContent("=isValid")
	assert.Equal(t, "=isValid", elem.GetTextContent())
}

func TestFindElementById(t *testing.T) {
	doc := NewDocument()
	root := doc.CreateElement("", "root")
	child := doc.CreateElement("", "child")
	child.SetAttribute("id", "myId")
	grandchild := doc.CreateElement("", "grandchild")
	grandchild.SetAttribute("id", "deepId")

	root.AppendChild(child)
	child.AppendChild(grandchild)
	doc.SetRoot(root)

	found := root.FindElementById("myId")
	require.NotNil(t, found)
	assert.Equal(t, "child", found.LocalName())

	deep := root.FindElementById("deepId")
	require.NotNil(t, deep)
	assert.Equal(t, "grandchild", deep.LocalName())

	notFound := root.FindElementById("nope")
	assert.Nil(t, notFound)
}

func TestFindElementsByNS(t *testing.T) {
	doc := NewDocument()
	doc.RegisterNamespace("bpmn", "http://www.omg.org/spec/BPMN/20100524/MODEL")

	root := doc.CreateElement("http://www.omg.org/spec/BPMN/20100524/MODEL", "definitions")
	proc := doc.CreateElement("http://www.omg.org/spec/BPMN/20100524/MODEL", "process")
	start := doc.CreateElement("http://www.omg.org/spec/BPMN/20100524/MODEL", "startEvent")
	end := doc.CreateElement("http://www.omg.org/spec/BPMN/20100524/MODEL", "startEvent")

	proc.AppendChild(start)
	proc.AppendChild(end)
	root.AppendChild(proc)
	doc.SetRoot(root)

	found := root.FindElementsByNS("http://www.omg.org/spec/BPMN/20100524/MODEL", "startEvent")
	assert.Len(t, found, 2)
}

func TestParseDocument(t *testing.T) {
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" id="Definitions_1">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="StartEvent_1"/>
    <bpmn:endEvent id="EndEvent_1"/>
  </bpmn:process>
</bpmn:definitions>`

	doc, err := ParseDocumentString(xmlStr)
	require.NoError(t, err)

	// Namespace should be auto-detected
	assert.Equal(t, "bpmn", doc.LookupPrefix("http://www.omg.org/spec/BPMN/20100524/MODEL"))

	root := doc.Root()
	require.NotNil(t, root)
	assert.Equal(t, "definitions", root.LocalName())
	assert.Equal(t, "bpmn", root.Prefix())
	assert.Equal(t, "Definitions_1", root.GetAttribute("id"))

	processes := root.GetChildElementsByNS("http://www.omg.org/spec/BPMN/20100524/MODEL", "process")
	require.Len(t, processes, 1)
	assert.Equal(t, "Process_1", processes[0].GetAttribute("id"))
	assert.Equal(t, "true", processes[0].GetAttribute("isExecutable"))

	children := processes[0].GetChildElements()
	assert.Len(t, children, 2)
}

func TestRoundTrip(t *testing.T) {
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" id="Definitions_1">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="StartEvent_1"/>
    <bpmn:endEvent id="EndEvent_1"/>
  </bpmn:process>
</bpmn:definitions>`

	doc, err := ParseDocumentString(xmlStr)
	require.NoError(t, err)

	output, err := doc.WriteToString()
	require.NoError(t, err)

	// Verify key parts are preserved
	assert.Contains(t, output, "definitions")
	assert.Contains(t, output, "Process_1")
	assert.Contains(t, output, "StartEvent_1")
	assert.Contains(t, output, "EndEvent_1")

	// Parse again
	doc2, err := ParseDocumentString(output)
	require.NoError(t, err)
	assert.Equal(t, "definitions", doc2.Root().LocalName())
}

func TestInsertChildAfter(t *testing.T) {
	doc := NewDocument()
	parent := doc.CreateElement("", "root")
	a := doc.CreateElement("", "a")
	b := doc.CreateElement("", "b")
	c := doc.CreateElement("", "c")

	parent.AppendChild(a)
	parent.AppendChild(c)
	doc.SetRoot(parent)

	parent.InsertChildAfter(b, a)

	children := parent.GetChildElements()
	require.Len(t, children, 3)
	assert.Equal(t, "a", children[0].LocalName())
	assert.Equal(t, "b", children[1].LocalName())
	assert.Equal(t, "c", children[2].LocalName())
}

func TestInsertChildAfterNil(t *testing.T) {
	doc := NewDocument()
	parent := doc.CreateElement("", "root")
	a := doc.CreateElement("", "a")
	b := doc.CreateElement("", "b")

	parent.AppendChild(b)
	parent.InsertChildAfter(a, nil) // prepend

	children := parent.GetChildElements()
	require.Len(t, children, 2)
	assert.Equal(t, "a", children[0].LocalName())
	assert.Equal(t, "b", children[1].LocalName())
}

func TestReplaceChild(t *testing.T) {
	doc := NewDocument()
	parent := doc.CreateElement("", "root")
	old := doc.CreateElement("", "old")
	new_ := doc.CreateElement("", "new")
	other := doc.CreateElement("", "other")

	parent.AppendChild(old)
	parent.AppendChild(other)
	doc.SetRoot(parent)

	parent.ReplaceChild(old, new_)

	children := parent.GetChildElements()
	require.Len(t, children, 2)
	assert.Equal(t, "new", children[0].LocalName())
	assert.Equal(t, "other", children[1].LocalName())
}

func TestNamespaceURIResolution(t *testing.T) {
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL">
  <bpmn:process id="p1"/>
</bpmn:definitions>`

	doc, err := ParseDocumentString(xmlStr)
	require.NoError(t, err)

	root := doc.Root()
	assert.Equal(t, "http://www.omg.org/spec/BPMN/20100524/MODEL", root.NamespaceURI())

	proc := root.GetChildElements()[0]
	assert.Equal(t, "http://www.omg.org/spec/BPMN/20100524/MODEL", proc.NamespaceURI())
}

func TestCreateDocumentWithNamespaces(t *testing.T) {
	doc := NewDocument()
	doc.RegisterNamespace("bpmn", "http://www.omg.org/spec/BPMN/20100524/MODEL")
	doc.RegisterNamespace("zeebe", "http://camunda.org/schema/zeebe/1.0")

	root := doc.CreateElement("http://www.omg.org/spec/BPMN/20100524/MODEL", "definitions")
	root.AddNamespaceDeclaration("bpmn", "http://www.omg.org/spec/BPMN/20100524/MODEL")
	root.AddNamespaceDeclaration("zeebe", "http://camunda.org/schema/zeebe/1.0")
	doc.SetRoot(root)

	proc := doc.CreateElement("http://www.omg.org/spec/BPMN/20100524/MODEL", "process")
	proc.SetAttribute("id", "p1")
	root.AppendChild(proc)

	output, err := doc.WriteToString()
	require.NoError(t, err)

	assert.True(t, strings.Contains(output, "bpmn:definitions"))
	assert.True(t, strings.Contains(output, "bpmn:process"))
	assert.True(t, strings.Contains(output, `xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"`))
}
