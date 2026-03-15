package biz

func buildPermissionTree(perms []*Permission) []*Permission {
	nodeMap := make(map[string]*Permission)

	// 先创建所有节点
	for _, p := range perms {
		nodeMap[p.ID] = p
	}

	var roots []*Permission

	// 构建树
	for _, p := range perms {
		node := nodeMap[p.ID]

		if p.ParentId == "" {
			roots = append(roots, node)
			continue
		}

		parent, ok := nodeMap[p.ParentId]
		if ok {
			parent.Children = append(parent.Children, node)
		} else {
			// 找不到父节点当作根节点
			roots = append(roots, node)
		}
	}

	return roots
}
