package xml

import (
	"fmt"
	"strings"

	"github.com/beevik/etree"
)

// Document wraps an etree.Document with namespace registry support.
type Document struct {
	tree       *etree.Document
	nsPrefixes map[string]string // URI -> prefix
	nsURIs     map[string]string // prefix -> URI
	prefixSeq  int
}

// NewDocument creates a new empty XML document.
func NewDocument() *Document {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	return &Document{
		tree:       doc,
		nsPrefixes: make(map[string]string),
		nsURIs:     make(map[string]string),
	}
}

// ParseDocument parses XML bytes into a Document.
func ParseDocument(data []byte) (*Document, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, fmt.Errorf("parse xml: %w", err)
	}

	d := &Document{
		tree:       doc,
		nsPrefixes: make(map[string]string),
		nsURIs:     make(map[string]string),
	}

	// Build namespace registry from root element's xmlns attributes.
	if root := doc.Root(); root != nil {
		d.scanNamespaces(root)
	}

	return d, nil
}

// ParseDocumentString parses an XML string into a Document.
func ParseDocumentString(s string) (*Document, error) {
	return ParseDocument([]byte(s))
}

// scanNamespaces walks the element tree and collects xmlns declarations.
func (d *Document) scanNamespaces(elem *etree.Element) {
	for _, attr := range elem.Attr {
		if attr.Space == "xmlns" {
			d.nsPrefixes[attr.Value] = attr.Key
			d.nsURIs[attr.Key] = attr.Value
		} else if attr.Space == "" && attr.Key == "xmlns" {
			d.nsPrefixes[attr.Value] = ""
			d.nsURIs[""] = attr.Value
		}
	}
	for _, child := range elem.ChildElements() {
		d.scanNamespaces(child)
	}
}

// RegisterNamespace registers a namespace URI with a given prefix.
func (d *Document) RegisterNamespace(prefix, uri string) {
	d.nsPrefixes[uri] = prefix
	d.nsURIs[prefix] = uri
}

// RegisterNamespaceURI registers a namespace URI with an auto-generated prefix.
func (d *Document) RegisterNamespaceURI(uri string) string {
	if prefix, ok := d.nsPrefixes[uri]; ok {
		return prefix
	}
	prefix := fmt.Sprintf("ns%d", d.prefixSeq)
	d.prefixSeq++
	d.RegisterNamespace(prefix, uri)
	return prefix
}

// LookupPrefix returns the prefix for a namespace URI, or empty string if not found.
func (d *Document) LookupPrefix(uri string) string {
	return d.nsPrefixes[uri]
}

// LookupURI returns the namespace URI for a prefix, or empty string if not found.
func (d *Document) LookupURI(prefix string) string {
	return d.nsURIs[prefix]
}

// Root returns the root element, or nil if the document is empty.
func (d *Document) Root() *Element {
	root := d.tree.Root()
	if root == nil {
		return nil
	}
	return &Element{elem: root, doc: d}
}

// SetRoot sets the root element of the document.
func (d *Document) SetRoot(e *Element) {
	if existing := d.tree.Root(); existing != nil {
		d.tree.RemoveChild(existing)
	}
	d.tree.AddChild(e.elem)
	e.doc = d
}

// CreateElement creates a new element with the given namespace URI and local name.
// The element is not attached to the document tree.
func (d *Document) CreateElement(nsURI, localName string) *Element {
	elem := etree.NewElement(localName)
	if nsURI != "" {
		prefix := d.LookupPrefix(nsURI)
		if prefix == "" {
			prefix = d.RegisterNamespaceURI(nsURI)
		}
		elem.Space = prefix
	}
	return &Element{elem: elem, doc: d}
}

// WriteToBytes serializes the document to XML bytes.
func (d *Document) WriteToBytes() ([]byte, error) {
	d.tree.Indent(2)
	return d.tree.WriteToBytes()
}

// WriteToString serializes the document to an XML string.
func (d *Document) WriteToString() (string, error) {
	d.tree.Indent(2)
	return d.tree.WriteToString()
}

// Unwrap returns the underlying etree.Document.
func (d *Document) Unwrap() *etree.Document {
	return d.tree
}

// Element wraps an etree.Element with namespace-aware attribute access.
type Element struct {
	elem *etree.Element
	doc  *Document
}

// WrapElement wraps an existing etree.Element.
func WrapElement(elem *etree.Element, doc *Document) *Element {
	if elem == nil {
		return nil
	}
	return &Element{elem: elem, doc: doc}
}

// LocalName returns the element's local name (tag without namespace prefix).
func (e *Element) LocalName() string {
	return e.elem.Tag
}

// NamespaceURI returns the element's namespace URI.
func (e *Element) NamespaceURI() string {
	if e.elem.Space == "" {
		// Check for default namespace
		if e.doc != nil {
			return e.doc.LookupURI("")
		}
		return ""
	}
	// Try document registry first
	if e.doc != nil {
		if uri := e.doc.LookupURI(e.elem.Space); uri != "" {
			return uri
		}
	}
	// Fallback to etree resolution
	return e.elem.NamespaceURI()
}

// Prefix returns the element's namespace prefix.
func (e *Element) Prefix() string {
	return e.elem.Space
}

// GetAttribute returns the value of an attribute by local name (no namespace).
func (e *Element) GetAttribute(name string) string {
	attr := e.elem.SelectAttr(name)
	if attr != nil && attr.Space == "" {
		return attr.Value
	}
	// Also search without space constraint
	for _, a := range e.elem.Attr {
		if a.Key == name && a.Space == "" {
			return a.Value
		}
	}
	return ""
}

// GetAttributeNS returns the value of a namespaced attribute.
func (e *Element) GetAttributeNS(nsURI, name string) string {
	if nsURI == "" {
		return e.GetAttribute(name)
	}
	prefix := ""
	if e.doc != nil {
		prefix = e.doc.LookupPrefix(nsURI)
	}
	for _, a := range e.elem.Attr {
		if a.Key == name && a.Space == prefix {
			return a.Value
		}
	}
	return ""
}

// SetAttribute sets an attribute value by local name (no namespace).
func (e *Element) SetAttribute(name, value string) {
	e.elem.CreateAttr(name, value)
}

// SetAttributeNS sets a namespaced attribute value.
func (e *Element) SetAttributeNS(nsURI, name, value string) {
	if nsURI == "" {
		e.SetAttribute(name, value)
		return
	}
	prefix := ""
	if e.doc != nil {
		prefix = e.doc.LookupPrefix(nsURI)
		if prefix == "" {
			prefix = e.doc.RegisterNamespaceURI(nsURI)
		}
	}
	// Update existing or add new
	for i, a := range e.elem.Attr {
		if a.Key == name && a.Space == prefix {
			e.elem.Attr[i].Value = value
			return
		}
	}
	e.elem.Attr = append(e.elem.Attr, etree.Attr{Space: prefix, Key: name, Value: value})
}

// RemoveAttribute removes an attribute by local name.
func (e *Element) RemoveAttribute(name string) {
	e.elem.RemoveAttr(name)
}

// HasAttribute returns true if the element has the given attribute.
func (e *Element) HasAttribute(name string) bool {
	return e.elem.SelectAttr(name) != nil
}

// GetChildElements returns all direct child elements.
func (e *Element) GetChildElements() []*Element {
	children := e.elem.ChildElements()
	result := make([]*Element, len(children))
	for i, child := range children {
		result[i] = &Element{elem: child, doc: e.doc}
	}
	return result
}

// GetChildElementsByNS returns child elements matching the given namespace URI and local name.
func (e *Element) GetChildElementsByNS(nsURI, localName string) []*Element {
	var result []*Element
	prefix := ""
	if nsURI != "" && e.doc != nil {
		prefix = e.doc.LookupPrefix(nsURI)
	}
	for _, child := range e.elem.ChildElements() {
		if child.Tag == localName && child.Space == prefix {
			result = append(result, &Element{elem: child, doc: e.doc})
		}
	}
	return result
}

// AppendChild adds a child element.
func (e *Element) AppendChild(child *Element) {
	e.elem.AddChild(child.elem)
	child.doc = e.doc
}

// RemoveChild removes a child element.
func (e *Element) RemoveChild(child *Element) {
	e.elem.RemoveChild(child.elem)
}

// InsertChildAfter inserts newChild after the reference element.
// If after is nil, the child is prepended.
func (e *Element) InsertChildAfter(newChild, after *Element) {
	if after == nil {
		// Prepend: insert at index 0
		e.elem.InsertChildAt(0, newChild.elem)
	} else {
		idx := after.elem.Index()
		e.elem.InsertChildAt(idx+1, newChild.elem)
	}
	newChild.doc = e.doc
}

// ReplaceChild replaces oldChild with newChild.
func (e *Element) ReplaceChild(oldChild, newChild *Element) {
	idx := oldChild.elem.Index()
	if idx < 0 {
		return
	}
	e.elem.InsertChildAt(idx, newChild.elem)
	e.elem.RemoveChildAt(idx + 1)
	newChild.doc = e.doc
}

// Parent returns the parent element, or nil if this is the root.
func (e *Element) Parent() *Element {
	parent := e.elem.Parent()
	if parent == nil {
		return nil
	}
	return &Element{elem: parent, doc: e.doc}
}

// GetTextContent returns the text content of the element.
func (e *Element) GetTextContent() string {
	return strings.TrimSpace(e.elem.Text())
}

// SetTextContent sets the text content of the element.
func (e *Element) SetTextContent(text string) {
	e.elem.SetText(text)
}

// GetDocument returns the owning document.
func (e *Element) GetDocument() *Document {
	return e.doc
}

// SetDocument sets the owning document.
func (e *Element) SetDocument(doc *Document) {
	e.doc = doc
}

// Unwrap returns the underlying etree.Element.
func (e *Element) Unwrap() *etree.Element {
	return e.elem
}

// AddNamespaceDeclaration adds an xmlns:prefix="uri" attribute to this element.
func (e *Element) AddNamespaceDeclaration(prefix, uri string) {
	if prefix == "" {
		e.elem.CreateAttr("xmlns", uri)
	} else {
		e.elem.CreateAttr("xmlns:"+prefix, uri)
	}
}

// FindElementById searches this element's subtree for an element with the given id attribute.
func (e *Element) FindElementById(id string) *Element {
	if e.GetAttribute("id") == id {
		return e
	}
	for _, child := range e.elem.ChildElements() {
		wrapped := &Element{elem: child, doc: e.doc}
		if found := wrapped.FindElementById(id); found != nil {
			return found
		}
	}
	return nil
}

// FindElementsByNS finds all elements in this subtree matching namespace URI and local name.
func (e *Element) FindElementsByNS(nsURI, localName string) []*Element {
	var result []*Element
	prefix := ""
	if nsURI != "" && e.doc != nil {
		prefix = e.doc.LookupPrefix(nsURI)
	}
	e.findElementsByNSRecursive(prefix, localName, &result)
	return result
}

func (e *Element) findElementsByNSRecursive(prefix, localName string, result *[]*Element) {
	if e.elem.Tag == localName && e.elem.Space == prefix {
		*result = append(*result, e)
	}
	for _, child := range e.elem.ChildElements() {
		wrapped := &Element{elem: child, doc: e.doc}
		wrapped.findElementsByNSRecursive(prefix, localName, result)
	}
}
