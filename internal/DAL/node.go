package dal

import (
	"bytes"
	"encoding/binary"
)

type Item struct {
	Key   []byte
	Value []byte
}
type Node struct {
	*DAL
	PageNum    uint64
	Items      []*Item
	Childnodes []uint64
}

func NewEmptyNode() *Node {
	return &Node{}
}

func newItem(key []byte, value []byte) *Item {
	return &Item{
		Key:   key,
		Value: value,
	}
}

func (n *Node) isLeaf() bool {
	return len(n.Childnodes) == 0
}

func (n *Node) Serialize(buf []byte) []byte {
	leftPos := 0
	rightPos := len(buf) - 1

	// Add page header: isLeaf, key-value pairs count, node num
	// isLeaf
	isLeaf := n.isLeaf()
	var bitSetVar uint64
	if isLeaf {
		bitSetVar = 1
	}
	buf[leftPos] = byte(bitSetVar)
	leftPos += 1

	// key-value pairs count
	binary.LittleEndian.PutUint16(buf[leftPos:], uint16(len(n.Items)))
	leftPos += 2

	// We use slotted pages for storing data in the page. It means the actual keys and values (the cells) are appended
	// to right of the page whereas offsets have a fixed size and are appended from the left.
	// It's easier to preserve the logical order (alphabetical in the case of b-tree) using the metadata and performing
	// pointer arithmetic. Using the data itself is harder as it varies by size.

	// Page structure is:
	// ----------------------------------------------------------------------------------
	// |  Page  | key-value /  child node    key-value 		      |    key-value		 |
	// | Header |   offset /	 pointer	  offset         .... |      data      ..... |
	// ----------------------------------------------------------------------------------

	for i := 0; i < len(n.Items); i++ {
		item := n.Items[i]
		if !isLeaf {
			childNode := n.Childnodes[i]

			// Write the child page as a fixed size of 8 bytes
			binary.LittleEndian.PutUint64(buf[leftPos:], uint64(childNode))
			leftPos += PageNumSize
		}

		klen := len(item.Key)
		vlen := len(item.Value)

		// write offset
		offset := rightPos - klen - vlen - 2
		binary.LittleEndian.PutUint16(buf[leftPos:], uint16(offset))
		leftPos += 2

		rightPos -= vlen
		copy(buf[rightPos:], item.Value)

		rightPos -= 1
		buf[rightPos] = byte(vlen)

		rightPos -= klen
		copy(buf[rightPos:], item.Key)

		rightPos -= 1
		buf[rightPos] = byte(klen)
	}

	if !isLeaf {
		// Write the last child node
		lastChildNode := n.Childnodes[len(n.Childnodes)-1]
		// Write the child page as a fixed size of 8 bytes
		binary.LittleEndian.PutUint64(buf[leftPos:], uint64(lastChildNode))
	}

	return buf
}

func (n *Node) Deserialize(buf []byte) {
	leftPos := 0

	// Read header
	isLeaf := uint16(buf[0])

	itemsCount := int(binary.LittleEndian.Uint16(buf[1:3]))
	leftPos += 3

	// Read body
	for i := 0; i < itemsCount; i++ {
		if isLeaf == 0 { // False
			pageNum := binary.LittleEndian.Uint64(buf[leftPos:])
			leftPos += n.DAL.PageSize
			// checkkk ^
			n.Childnodes = append(n.Childnodes, uint64(pageNum))
		}

		// Read offset
		offset := binary.LittleEndian.Uint16(buf[leftPos:])
		leftPos += 2

		klen := uint16(buf[int(offset)])
		offset += 1

		key := buf[offset : offset+klen]
		offset += klen

		vlen := uint16(buf[int(offset)])
		offset += 1

		value := buf[offset : offset+vlen]
		offset += vlen
		n.Items = append(n.Items, newItem(key, value))
	}

	if isLeaf == 0 { // False
		// Read the last child node
		pageNum := uint64(binary.LittleEndian.Uint64(buf[leftPos:]))
		n.Childnodes = append(n.Childnodes, pageNum)
	}
}

// B-tree node on-disk accessors migrated from btree/node.go
func (d *DAL) GetNode(pageNum uint64) (*Node, error) {
	p, err := d.ReadPage(pageNum)
	if err != nil {
		return nil, err
	}
	node := NewEmptyNode()
	node.Deserialize(p.Data)
	node.PageNum = pageNum
	node.DAL = d
	return node, nil
}

func (d *DAL) NewNode(items []*Item, childNodes []uint64) *Node {
	node := NewEmptyNode()
	node.Items = items
	node.Childnodes = childNodes
	node.PageNum = d.GetNextPage()
	node.DAL = d
	return node
}

func (d *DAL) WriteNode(n *Node) (*Node, error) {
	p := d.AllocateEmptyPage()
	if n.PageNum == 0 {
		p.Num = d.GetNextPage()
		n.PageNum = p.Num
	} else {
		p.Num = n.PageNum
	}
	p.Data = n.Serialize(p.Data)
	err := d.WritePage(p)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (d *DAL) DeleteNode(pageNum uint64) {
	d.ReleasePage(pageNum)
}

// Returns the root node (page 1); creates if necessary.
// func (d *DAL) getOrCreateRoot() (*Node, error) {
// 	n, err := d.GetNode(1)
// 	if err == nil {
// 		return n, nil
// 	}
// 	// If not found or any error, create new root.
// 	n = NewEmptyNode()
// 	n.PageNum = 1
// 	n.DAL = d
// 	_, err = d.WriteNode(n)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return n, nil
// }

// // Put adds or updates a key-value pair in the root node only (no splits/children).
// func (d *DAL) Put(key, value []byte) error {
// 	root, err := d.getOrCreateRoot()
// 	if err != nil {
// 		return err
// 	}
// 	found := false
// 	for _, item := range root.Items {
// 		if string(item.key) == string(key) {
// 			item.value = value
// 			found = true
// 			break
// 		}
// 	}
// 	if !found {
// 		root.Items = append(root.Items, newItem(key, value))
// 	}
// 	_, err = d.WriteNode(root)
// 	return err
// }

// // Get retrieves a value for a key from the root node only.
// func (d *DAL) Get(key []byte) ([]byte, bool) {
// 	root, err := d.getOrCreateRoot()
// 	if err != nil {
// 		return nil, false
// 	}
// 	for _, item := range root.Items {
// 		if string(item.key) == string(key) {
// 			return item.value, true
// 		}
// 	}
// 	return nil, false
// }

func (n *Node) WriteNode(node *Node) *Node {
	ret, _ := n.DAL.WriteNode(node)
	return ret
}

func (n *Node) WriteNodes(nodes ...*Node) {
	for _, node := range nodes {
		n.WriteNode(node)
	}
}

func (n *Node) GetNode(pageNum uint64) (*Node, error) {
	return n.DAL.GetNode(pageNum)
}

func (n *Node) FindKeyInNode(key []byte) (bool, int) {
	for i, existingItem := range n.Items {
		res := bytes.Compare(existingItem.Key, key)
		if res == 0 { // Keys match
			return true, i
		}

		// The key is bigger than the previous key, so it doesn't exist in the node, but may exist in child nodes.
		if res == 1 {
			return false, i
		}
	}

	// The key isn't bigger than any of the keys which means it's in the last index.
	return false, len(n.Items)
}

func (n *Node) findKey(key []byte, exact bool) (int, *Node, []int, error) {
	ancestorsIndexes := []int{0} // index of root
	index, node, err := findKeyHelper(n, key, exact, &ancestorsIndexes)
	if err != nil {
		return -1, nil, nil, err
	}
	return index, node, ancestorsIndexes, nil
}

func findKeyHelper(node *Node, key []byte, exact bool, ancestorsIndexes *[]int) (int, *Node, error) {
	wasFound, index := node.FindKeyInNode(key)
	if wasFound {
		return index, node, nil
	}

	if node.isLeaf() {
		if exact {
			return -1, nil, nil
		}
		return index, node, nil
	}

	*ancestorsIndexes = append(*ancestorsIndexes, index)
	nextChild, err := node.GetNode(node.Childnodes[index])
	if err != nil {
		return -1, nil, err
	}
	return findKeyHelper(nextChild, key, exact, ancestorsIndexes)
}

func (n *Node) elementSize(i int) int {
	size := 0
	size += len(n.Items[i].Key)
	size += len(n.Items[i].Value)
	size += PageNumSize // 8 is the pgnum size
	return size
}

func (n *Node) nodeSize() int {
	size := 0
	size += NodeHeaderSize

	for i := range n.Items {
		size += n.elementSize(i)
	}

	// Add last page
	size += PageNumSize // 8 is the pgnum size
	return size
}

func (n *Node) addItem(item *Item, insertionIndex int) int {
	if len(n.Items) == insertionIndex { // nil or empty slice or after last element
		n.Items = append(n.Items, item)
		return insertionIndex
	}

	n.Items = append(n.Items[:insertionIndex+1], n.Items[insertionIndex:]...)
	n.Items[insertionIndex] = item
	return insertionIndex
}

func (n *Node) isOverPopulated() bool {
	return n.DAL.isOverPopulated(n)
}

// isUnderPopulated checks if the node size is smaller than the size of a page.
func (n *Node) isUnderPopulated() bool {
	return n.DAL.isUnderPopulated(n)
}

func (n *Node) split(nodeToSplit *Node, nodeToSplitIndex int) {
	// The first index where min amount of bytes to populate a page is achieved. Then add 1 so it will be split one
	// index after.
	splitIndex := nodeToSplit.DAL.getSplitIndex(nodeToSplit)

	middleItem := nodeToSplit.Items[splitIndex]
	var newNode *Node

	if nodeToSplit.isLeaf() {
		// newNode, _ = n.WriteNode(n.DAL.NewNode())
		newNode = n.WriteNode(n.DAL.NewNode(nodeToSplit.Items[splitIndex+1:], []uint64{}))
		nodeToSplit.Items = nodeToSplit.Items[:splitIndex]
	} else {
		newNode = n.WriteNode(n.DAL.NewNode(nodeToSplit.Items[splitIndex+1:], nodeToSplit.Childnodes[splitIndex+1:]))
		nodeToSplit.Items = nodeToSplit.Items[:splitIndex]
		nodeToSplit.Childnodes = nodeToSplit.Childnodes[:splitIndex+1]
	}
	n.addItem(middleItem, nodeToSplitIndex)
	if len(n.Childnodes) == nodeToSplitIndex+1 { // If middle of list, then move items forward
		n.Childnodes = append(n.Childnodes, newNode.PageNum)
	} else {
		n.Childnodes = append(n.Childnodes[:nodeToSplitIndex+1], n.Childnodes[nodeToSplitIndex:]...)
		n.Childnodes[nodeToSplitIndex+1] = newNode.PageNum
	}

	n.WriteNodes(n, nodeToSplit)
}
