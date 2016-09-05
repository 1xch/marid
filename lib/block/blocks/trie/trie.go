package trie

var tt string = `{{ extends "block_base" }}
{{ define "block_root" }}package {{.PackageName}}

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

type Prefix []byte

type VisitorFunc func(Prefix, *{{ .ItemName }}) error

type Trie struct {
	prefix                   Prefix
	maxPrefixPerNode         int
	maxChildrenPerSparseNode int
	children                 childList
	item                     *{{ .ItemName }}
}

const (
	DefaultMaxPrefixPerNode         = 10
	DefaultMaxChildrenPerSparseNode = 8
)

type Option func(*Trie)

func NewTrie(options ...Option) *Trie {
	trie := &Trie{}

	for _, opt := range options {
		opt(trie)
	}

	if trie.maxPrefixPerNode <= 0 {
		trie.maxPrefixPerNode = DefaultMaxPrefixPerNode
	}
	if trie.maxChildrenPerSparseNode <= 0 {
		trie.maxChildrenPerSparseNode = DefaultMaxChildrenPerSparseNode
	}

	trie.children = newSparseChildList(trie.maxChildrenPerSparseNode)
	return trie
}

func MaxPrefixPerNode(value int) Option {
	return func(t *Trie) {
		t.maxPrefixPerNode = value
	}
}

func MaxChildrenPerSparseNode(value int) Option {
	return func(t *Trie) {
		t.maxChildrenPerSparseNode = value
	}
}

func WithPrefix(p string) Option {
	return func(t *Trie) {
		t.prefix = Prefix(p)
	}
}

func (t *Trie) Tagged() string {
	return string(t.prefix)
}

func (t *Trie) {{  .ItemName }}() *{{ .ItemName }} {
	return t.item
}

func (t *Trie) Insert(i *{{ .ItemName }}) bool {
	return t.put(i, false)
}

func (t *Trie) Set(items ...*{{ .ItemName }}) {
	for _, i := range items {
		t.put(i, true)
	}
}

func (t *Trie) Get(p Prefix) *{{ .ItemName }} {
	if p != nil {
		_, node, found, leftover := t.findSubtree(p)
		if !found || len(leftover) != 0 {
			return nil
		}
		return node.item
	}
	return nil
}

var No{{ .ItemName }}Error = Trror("No {{ .ItemName }} with the prefix %s available.").Out

func (t *Trie) Find(k string) (*{{ .ItemName }}, error) {
	key := Prefix(k)
	i := t.Get(key)
	if i == nil {
		return nil, No{{ .ItemName }}Error(k)
	}
	return i, nil
}

func (t *Trie) Match(p Prefix) bool {
	return t.Get(p) != nil
}

func (t *Trie) MatchSubtree(p Prefix) bool {
	_, _, matched, _ := t.findSubtree(p)
	return matched
}

func (t *Trie) Visit(v VisitorFunc) error {
	return t.walk(nil, v)
}

func (t *Trie) VisitSubtree(p Prefix, v VisitorFunc) error {
	// Nil prefix not allowed.
	if p == nil {
		panic(ErrNilPrefix)
	}

	// Empty trie must be handled explicitly.
	if t.prefix == nil {
		return nil
	}

	// Locate the relevant subtree.
	_, root, found, leftover := t.findSubtree(p)
	if !found {
		return nil
	}
	p = append(p, leftover...)

	// Visit it.
	return root.walk(p, v)
}

func (t *Trie) VisitPrefixes(p Prefix, v VisitorFunc) error {
	// Nil key not allowed.
	if p == nil {
		panic(ErrNilPrefix)
	}

	// Empty trie must be handled explicitly.
	if t.prefix == nil {
		return nil
	}

	// Walk the path matching key prefixes.
	node := t
	prefix := p
	offset := 0
	for {
		// Compute what part of prefix matches.
		common := node.longestCommonPrefixLength(p)
		p = p[common:]
		offset += common

		// Partial match means that there is no subtree matching prefix.
		if common < len(node.prefix) {
			return nil
		}

		// Call the visitor.
		if item := node.item; item != nil {
			if err := v(prefix[:offset], item); err != nil {
				return err
			}
		}

		if len(p) == 0 {
			// This node represents key, we are finished.
			return nil
		}

		// There is some key suffix left, move to the children.
		child := node.children.next(p[0])
		if child == nil {
			// There is nowhere to continue, return.
			return nil
		}

		node = child
	}
}

func (t *Trie) Delete(p Prefix) bool {
	// Nil prefix not allowed.
	if p == nil {
		panic(ErrNilPrefix)
	}

	// Empty trie must be handled explicitly.
	if t.prefix == nil {
		return false
	}

	// Find the relevant node.
	path, found, _ := t.findSubtreePath(p)
	if !found {
		return false
	}

	node := path[len(path)-1]
	var parent *Trie
	if len(path) != 1 {
		parent = path[len(path)-2]
	}

	// If the item is already set to nil, there is nothing to do.
	if node.item == nil {
		return false
	}

	// Delete the item.
	node.item = nil

	// Initialise i before goto.
	// Will be used later in a loop.
	i := len(path) - 1

	// In case there are some child nodes, we cannot drop the whole subtree.
	// We can try to compact nodes, though.
	if node.children.length() != 0 {
		goto Compact
	}

	// In case we are at the root, just reset it and we are done.
	if parent == nil {
		node.reset()
		return true
	}

	// We can drop a subtree.
	// Find the first ancestor that has its value set or it has 2 or more child nodes.
	// That will be the node where to drop the subtree at.
	for ; i >= 0; i-- {
		if current := path[i]; current.item != nil || current.children.length() >= 2 {
			break
		}
	}

	// Handle the case when there is no such node.
	// In other words, we can reset the whole tree.
	if i == -1 {
		path[0].reset()
		return true
	}

	// We can just remove the subtree here.
	node = path[i]
	if i == 0 {
		parent = nil
	} else {
		parent = path[i-1]
	}
	// i+1 is always a valid index since i is never pointing to the last node.
	// The loop above skips at least the last node since we are sure that the item
	// is set to nil and it has no children, othewise we would be compacting instead.
	node.children.remove(path[i+1].prefix[0])

Compact:
	// The node is set to the first non-empty ancestor,
	// so try to compact since that might be possible now.
	if compacted := node.compact(); compacted != node {
		if parent == nil {
			*node = *compacted
		} else {
			parent.children.replace(node.prefix[0], compacted)
			*parent = *parent.compact()
		}
	}

	return true
}

func (t *Trie) DeleteSubtree(p Prefix) bool {
	// Nil prefix not allowed.
	if p == nil {
		panic(ErrNilPrefix)
	}

	// Empty trie must be handled explicitly.
	if t.prefix == nil {
		return false
	}

	// Locate the relevant subtree.
	parent, root, found, _ := t.findSubtree(p)
	if !found {
		return false
	}

	// If we are in the root of the trie, reset the trie.
	if parent == nil {
		root.reset()
		return true
	}

	// Otherwise remove the root node from its parent.
	parent.children.remove(root.prefix[0])
	return true
}

func (t *Trie) List() []*{{ .ItemName }} {
	var ret []*{{ .ItemName }}
	v := func(p Prefix, i *{{ .ItemName }}) error {
		ret = append(ret, i)
		return nil
	}
	t.walk(nil, v)
	return ret
}

func (t *Trie) MarshalJSON() ([]byte, error) {
	l := t.List()
	return json.Marshal(&l)
}

func (t *Trie) UnmarshalJSON(b []byte) error {
	var i []*{{ .ItemName }}
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}
	t.Set(i...)
	return nil
}

func (t *Trie) MarshalYAML() (interface{}, error) {
	return t.List(), nil
}

func (t *Trie) UnmarshalYAML(u func(interface{}) error) error {
	var i []*{{ .ItemName }}
	err := u(&i)
	if err != nil {
		return err
	}
	t.Set(i...)
	return nil
}

func (t *Trie) empty() bool {
	return t.item == nil && t.children.length() == 0
}

func (t *Trie) size() int {
	n := 0

	t.walk(nil, func(prefix Prefix, item *{{ .ItemName }}) error {
		n++
		return nil
	})

	return n
}

func (t *Trie) total() int {
	return 1 + t.children.total()
}

func (t *Trie) reset() {
	t.prefix = nil
	t.children = newSparseChildList(t.maxPrefixPerNode)
}

var ErrNilPrefix = Trror("Nil prefix passed into a method call")

func (t *Trie) put(item *{{ .ItemName }}, replace bool) bool {
	key := Prefix(item.Key)
	// Nil prefix not allowed.
	if key == nil {
		panic(ErrNilPrefix)
	}

	var (
		common int
		node   *Trie = t
		child  *Trie
	)

	if node.prefix == nil {
		if len(key) <= t.maxPrefixPerNode {
			node.prefix = key
			goto InsertItem
		}
		node.prefix = key[:t.maxPrefixPerNode]
		key = key[t.maxPrefixPerNode:]
		goto AppendChild
	}

	for {
		// Compute the longest common prefix length.
		common = node.longestCommonPrefixLength(key)
		key = key[common:]

		// Only a part matches, split.
		if common < len(node.prefix) {
			goto SplitPrefix
		}

		// common == len(node.prefix) since never (common > len(node.prefix))
		// common == len(former key) <-> 0 == len(key)
		// -> former key == node.prefix
		if len(key) == 0 {
			goto InsertItem
		}

		// Check children for matching prefix.
		child = node.children.next(key[0])
		if child == nil {
			goto AppendChild
		}
		node = child
	}

SplitPrefix:
	// Split the prefix if necessary.
	child = new(Trie)
	*child = *node
	*node = *NewTrie()
	node.prefix = child.prefix[:common]
	child.prefix = child.prefix[common:]
	child = child.compact()
	node.children = node.children.add(child)

AppendChild:
	// Keep appending children until whole prefix is inserted.
	// This loop starts with empty node.prefix that needs to be filled.
	for len(key) != 0 {
		child := NewTrie()
		if len(key) <= t.maxPrefixPerNode {
			child.prefix = key
			node.children = node.children.add(child)
			node = child
			goto InsertItem
		} else {
			child.prefix = key[:t.maxPrefixPerNode]
			key = key[t.maxPrefixPerNode:]
			node.children = node.children.add(child)
			node = child
		}
	}

InsertItem:
	if replace || node.item == nil {
		node.item = item
		return true
	}
	return false
}

func (t *Trie) longestCommonPrefixLength(prefix Prefix) (i int) {
	for ; i < len(prefix) && i < len(t.prefix) && prefix[i] == t.prefix[i]; i++ {
	}
	return
}

func (t *Trie) compact() *Trie {
	// Only a node with a single child can be compacted.
	if t.children.length() != 1 {
		return t
	}

	child := t.children.head()

	// If any item is set, we cannot compact since we want to retain
	// the ability to do searching by key. This makes compaction less usable,
	// but that simply cannot be avoided.
	if t.item != nil || child.item != nil {
		return t
	}

	// Make sure the combined prefixes fit into a single node.
	if len(t.prefix)+len(child.prefix) > t.maxPrefixPerNode {
		return t
	}

	// Concatenate the prefixes, move the items.
	child.prefix = append(t.prefix, child.prefix...)
	if t.item != nil {
		child.item = t.item
	}

	return child
}

func (t *Trie) findSubtree(prefix Prefix) (parent *Trie, root *Trie, found bool, leftover Prefix) {
	// Find the subtree matching prefix.
	root = t
	for {
		// Compute what part of prefix matches.
		common := root.longestCommonPrefixLength(prefix)
		prefix = prefix[common:]

		// We used up the whole prefix, subtree found.
		if len(prefix) == 0 {
			found = true
			leftover = root.prefix[common:]
			return
		}

		// Partial match means that there is no subtree matching prefix.
		if common < len(root.prefix) {
			leftover = root.prefix[common:]
			return
		}

		// There is some prefix left, move to the children.
		child := root.children.next(prefix[0])
		if child == nil {
			// There is nowhere to continue, there is no subtree matching prefix.
			return
		}

		parent = root
		root = child
	}
}

func (t *Trie) findSubtreePath(prefix Prefix) (path []*Trie, found bool, leftover Prefix) {
	// Find the subtree matching prefix.
	root := t
	var subtreePath []*Trie
	for {
		// Append the current root to the path.
		subtreePath = append(subtreePath, root)

		// Compute what part of prefix matches.
		common := root.longestCommonPrefixLength(prefix)
		prefix = prefix[common:]

		// We used up the whole prefix, subtree found.
		if len(prefix) == 0 {
			path = subtreePath
			found = true
			leftover = root.prefix[common:]
			return
		}

		// Partial match means that there is no subtree matching prefix.
		if common < len(root.prefix) {
			leftover = root.prefix[common:]
			return
		}

		// There is some prefix left, move to the children.
		child := root.children.next(prefix[0])
		if child == nil {
			// There is nowhere to continue, there is no subtree matching prefix.
			return
		}

		root = child
	}
}

var SkipSubtree = Trror("Skip this subtree")

func (t *Trie) walk(p Prefix, v VisitorFunc) error {
	var prefix Prefix
	// Allocate a bit more space for prefix at the beginning.
	if p == nil {
		prefix = make(Prefix, 32+len(t.prefix))
		copy(prefix, t.prefix)
		prefix = prefix[:len(t.prefix)]
	} else {
		prefix = make(Prefix, 32+len(p))
		copy(prefix, p)
		prefix = prefix[:len(p)]
	}

	// Visit the root first. Not that this works for empty trie as well since
	// in that case item == nil && len(children) == 0.
	if t.item != nil {
		if err := v(prefix, t.item); err != nil {
			if err == SkipSubtree {
				return nil
			}
			return err
		}
	}

	// Then continue to the children.
	return t.children.walk(&prefix, v)
}

func (t *Trie) print(writer io.Writer, indent int) {
	fmt.Fprintf(writer, "%s%s %v\n", strings.Repeat(" ", indent), string(t.prefix), t.item)
	t.children.print(writer, indent+2)
}

type childList interface {
	length() int
	head() *Trie
	add(child *Trie) childList
	remove(b byte)
	replace(b byte, child *Trie)
	next(b byte) *Trie
	walk(prefix *Prefix, visitor VisitorFunc) error
	print(w io.Writer, indent int)
	total() int
}

type tries []*Trie

func (t tries) Len() int {
	return len(t)
}

func (t tries) Less(i, j int) bool {
	strings := sort.StringSlice{string(t[i].prefix), string(t[j].prefix)}
	return strings.Less(0, 1)
}

func (t tries) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type sparseChildList struct {
	children tries
}

func newSparseChildList(maxChildrenPerSparseNode int) childList {
	return &sparseChildList{
		children: make(tries, 0, maxChildrenPerSparseNode),
	}
}

func (list *sparseChildList) length() int {
	return len(list.children)
}

func (list *sparseChildList) head() *Trie {
	return list.children[0]
}

func (list *sparseChildList) add(child *Trie) childList {
	// Search for an empty spot and insert the child if possible.
	if len(list.children) != cap(list.children) {
		list.children = append(list.children, child)
		return list
	}

	// Otherwise we have to transform to the dense list type.
	return newDenseChildList(list, child)
}

func (list *sparseChildList) remove(b byte) {
	for i, node := range list.children {
		if node.prefix[0] == b {
			list.children[i] = list.children[len(list.children)-1]
			list.children[len(list.children)-1] = nil
			list.children = list.children[:len(list.children)-1]
			return
		}
	}

	// This is not supposed to be reached.
	panic("removing non-existent child")
}

func (list *sparseChildList) replace(b byte, child *Trie) {
	// Make a consistency check.
	if p0 := child.prefix[0]; p0 != b {
		panic(fmt.Errorf("child prefix mismatch: %v != %v", p0, b))
	}

	// Seek the child and replace it.
	for i, node := range list.children {
		if node.prefix[0] == b {
			list.children[i] = child
			return
		}
	}
}

func (list *sparseChildList) next(b byte) *Trie {
	for _, child := range list.children {
		if child.prefix[0] == b {
			return child
		}
	}
	return nil
}

func (list *sparseChildList) walk(prefix *Prefix, visitor VisitorFunc) error {
	sort.Sort(list.children)

	for _, child := range list.children {
		*prefix = append(*prefix, child.prefix...)
		if child.item != nil {
			err := visitor(*prefix, child.item)
			if err != nil {
				if err == SkipSubtree {
					*prefix = (*prefix)[:len(*prefix)-len(child.prefix)]
					continue
				}
				*prefix = (*prefix)[:len(*prefix)-len(child.prefix)]
				return err
			}
		}

		err := child.children.walk(prefix, visitor)
		*prefix = (*prefix)[:len(*prefix)-len(child.prefix)]
		if err != nil {
			return err
		}
	}

	return nil
}

func (list *sparseChildList) total() int {
	tot := 0
	for _, child := range list.children {
		if child != nil {
			tot = tot + child.total()
		}
	}
	return tot
}

func (list *sparseChildList) print(w io.Writer, indent int) {
	for _, child := range list.children {
		if child != nil {
			child.print(w, indent)
		}
	}
}

type denseChildList struct {
	min         int
	max         int
	numChildren int
	headIndex   int
	children    []*Trie
}

func newDenseChildList(list *sparseChildList, child *Trie) childList {
	var (
		min int = 255
		max int = 0
	)
	for _, child := range list.children {
		b := int(child.prefix[0])
		if b < min {
			min = b
		}
		if b > max {
			max = b
		}
	}

	b := int(child.prefix[0])
	if b < min {
		min = b
	}
	if b > max {
		max = b
	}

	children := make([]*Trie, max-min+1)
	for _, child := range list.children {
		children[int(child.prefix[0])-min] = child
	}
	children[int(child.prefix[0])-min] = child

	return &denseChildList{
		min:         min,
		max:         max,
		numChildren: list.length() + 1,
		headIndex:   0,
		children:    children,
	}
}

func (list *denseChildList) length() int {
	return list.numChildren
}

func (list *denseChildList) head() *Trie {
	return list.children[list.headIndex]
}

func (list *denseChildList) add(child *Trie) childList {
	b := int(child.prefix[0])
	var i int

	switch {
	case list.min <= b && b <= list.max:
		if list.children[b-list.min] != nil {
			panic("dense child list collision detected")
		}
		i = b - list.min
		list.children[i] = child

	case b < list.min:
		children := make([]*Trie, list.max-b+1)
		i = 0
		children[i] = child
		copy(children[list.min-b:], list.children)
		list.children = children
		list.min = b

	default: // b > list.max
		children := make([]*Trie, b-list.min+1)
		i = b - list.min
		children[i] = child
		copy(children, list.children)
		list.children = children
		list.max = b
	}

	list.numChildren++
	if i < list.headIndex {
		list.headIndex = i
	}
	return list
}

func (list *denseChildList) remove(b byte) {
	i := int(b) - list.min
	if list.children[i] == nil {
		// This is not supposed to be reached.
		panic("removing non-existent child")
	}
	list.numChildren--
	list.children[i] = nil

	// Update head index.
	if i == list.headIndex {
		for ; i < len(list.children); i++ {
			if list.children[i] != nil {
				list.headIndex = i
				return
			}
		}
	}
}

func (list *denseChildList) replace(b byte, child *Trie) {
	// Make a consistency check.
	if p0 := child.prefix[0]; p0 != b {
		panic(fmt.Errorf("child prefix mismatch: %v != %v", p0, b))
	}

	// Replace the child.
	list.children[int(b)-list.min] = child
}

func (list *denseChildList) next(b byte) *Trie {
	i := int(b)
	if i < list.min || list.max < i {
		return nil
	}
	return list.children[i-list.min]
}

func (list *denseChildList) walk(prefix *Prefix, visitor VisitorFunc) error {
	for _, child := range list.children {
		if child == nil {
			continue
		}
		*prefix = append(*prefix, child.prefix...)
		if child.item != nil {
			if err := visitor(*prefix, child.item); err != nil {
				if err == SkipSubtree {
					*prefix = (*prefix)[:len(*prefix)-len(child.prefix)]
					continue
				}
				*prefix = (*prefix)[:len(*prefix)-len(child.prefix)]
				return err
			}
		}

		err := child.children.walk(prefix, visitor)
		*prefix = (*prefix)[:len(*prefix)-len(child.prefix)]
		if err != nil {
			return err
		}
	}

	return nil
}

func (list *denseChildList) print(w io.Writer, indent int) {
	for _, child := range list.children {
		if child != nil {
			child.print(w, indent)
		}
	}
}

func (list *denseChildList) total() int {
	tot := 0
	for _, child := range list.children {
		if child != nil {
			tot = tot + child.total()
		}
	}
	return tot
}

type trror struct {
	base  string
	vals []interface{}
}

func (t *trror) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf({{.Letter}}.base, {{.Letter}}.vals...))
}

func (t *trror) Out(vals ...interface{}) *trror {
	{{.Letter}}.vals = vals
	return {{.Letter}}
}

func Trror(base string) *trror {
	return &trror{base: base}
}

{{ end }}
`
