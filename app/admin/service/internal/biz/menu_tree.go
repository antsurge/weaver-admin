package biz

func buildMenuTree(perms []*Menu) []*Menu {
	nodeMap := make(map[string]*Menu)

	// 先创建所有节点
	for _, p := range perms {
		nodeMap[p.ID] = p
	}

	var roots []*Menu

	// 构建树
	for _, p := range perms {
		node := nodeMap[p.ID]

		if p.ParentID == "" {
			roots = append(roots, node)
			continue
		}

		parent, ok := nodeMap[p.ParentID]
		if ok {
			parent.Children = append(parent.Children, node)
		} else {
			// 找不到父节点当作根节点
			roots = append(roots, node)
		}
	}

	return roots
}
