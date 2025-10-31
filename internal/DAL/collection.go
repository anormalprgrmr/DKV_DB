package dal

import "bytes"

type Collection struct {
	name []byte
	root uint64

	DAL *DAL
}

func NewCollection(name []byte, root uint64) *Collection {
	return &Collection{
		name: name,
		root: root,
	}
}

func (c *Collection) Find(key []byte) (*Item, error) {
	n, err := c.DAL.GetNode(c.root)
	if err != nil {
		return nil, err
	}

	index, containingNode, _, err := n.findKey(key, true)
	if err != nil {
		return nil, err
	}
	if index == -1 {
		return nil, nil
	}
	return containingNode.Items[index], nil
}

func (c *Collection) Put(key []byte, value []byte) error {
	i := newItem(key, value)

	// On first insertion the root node does not exist, so it should be created
	var root *Node
	var err error
	if c.root == 0 {
		root, err = c.DAL.WriteNode(c.DAL.NewNode([]*Item{i}, []uint64{}))
		if err != nil {
			return nil
		}
		c.root = root.PageNum
		return nil
	} else {
		root, err = c.DAL.GetNode(c.root)
		if err != nil {
			return err
		}
	}

	// Find the path to the node where the insertion should happen
	insertionIndex, nodeToInsertIn, ancestorsIndexes, err := root.findKey(i.Key, false)
	if err != nil {
		return err
	}

	// If key already exists
	if nodeToInsertIn.Items != nil && bytes.Compare(nodeToInsertIn.Items[insertionIndex].Key, key) == 0 {
		nodeToInsertIn.Items[insertionIndex] = i
	} else {
		// Add item to the leaf node
		nodeToInsertIn.addItem(i, insertionIndex)
	}
	_, err = c.DAL.WriteNode(nodeToInsertIn)
	if err != nil {
		return err
	}

	ancestors, err := c.getNodes(ancestorsIndexes)
	if err != nil {
		return err
	}

	// Rebalance the nodes all the way up. Start From one node before the last and go all the way up. Exclude root.
	for i := len(ancestors) - 2; i >= 0; i-- {
		pnode := ancestors[i]
		node := ancestors[i+1]
		nodeIndex := ancestorsIndexes[i+1]
		if node.isOverPopulated() {
			pnode.split(node, nodeIndex)
		}
	}

	// Handle root
	rootNode := ancestors[0]
	if rootNode.isOverPopulated() {
		newRoot := c.DAL.NewNode([]*Item{}, []uint64{rootNode.PageNum})
		newRoot.split(rootNode, 0)

		// commit newly created root
		newRoot, err = c.DAL.WriteNode(newRoot)
		if err != nil {
			return err
		}

		c.root = newRoot.PageNum
	}
	return nil
}

func (c *Collection) getNodes(indexes []int) ([]*Node, error) {
	root, err := c.DAL.GetNode(c.root)
	if err != nil {
		return nil, err
	}

	nodes := []*Node{root}
	child := root
	for i := 1; i < len(indexes); i++ {
		child, err = c.DAL.GetNode(child.Childnodes[indexes[i]])
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, child)
	}
	return nodes, nil
}
