package algorithm

// 最小编辑距离  s1--> s2
func EditDistance(s1, s2 string) int {
	m, n := len(s1), len(s2)
	if m == 0 {
		return n // 如果 s1 为空，返回 s2 的长度
	}
	if n == 0 {
		return m // 如果 s2 为空，返回 s1 的长度
	}

	// 创建二维 DP 数组
	dp := make([][]int, m) // dp[i][j] 表示 s1 的前 i 个字符转换为 s2 的前 j 个字符所需的最小操作数
	for i := range dp {
		dp[i] = make([]int, n)
	}

	// 初始化第一行和第一列
	for i := 0; i < m; i++ { // dp[i][0] 表示将 s1 的前 i 个字符转换为空字符串所需的操作数
		dp[i][0] = i // 删除操作
	}
	for j := 0; j < n; j++ { // dp[0][j] 表示将空字符串转换为 s2 的前 j 个字符所需的操作数
		dp[0][j] = j // 插入操作
	}

	// 填充 DP 数组
	for i := 1; i < m; i++ {
		for j := 1; j < n; j++ {
			if s1[i] == s2[j] {
				dp[i][j] = dp[i-1][j-1] // 如果字符相同，不需要操作
			} else {
				// 只需要当前行和前一行的值
				// 取删除、插入和替换操作的最小值
				dp[i][j] = min(dp[i][j-1]+1, dp[i-1][j]+1, dp[i-1][j-1]+1)
			}
		}
	}

	return dp[m-1][n-1] // 返回将 s1 转换为 s2 所需的最小操作数
}

func EditDistanceWithButtomUp(s1, s2 string) int {
	m, n := len(s1), len(s2)
	if m == 0 {
		return n // 如果 s1 为空，返回 s2 的长度
	}
	if n == 0 {
		return m // 如果 s2 为空，返回 s1 的长度
	}

	prevRow := make([]int, n+1)    // 前一行的 DP 数组
	currentRow := make([]int, n+1) // 当前行的 DP 数组
	for j := 0; j <= n; j++ {
		prevRow[j] = j // 初始化第一行
	}
	for i := 1; i <= m; i++ {
		currentRow[0] = i // 初始化当前行的第一列
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				currentRow[j] = prevRow[j-1] // 如果字符相同，不需要操作
			} else {
				// 取删除、插入和替换操作的最小值
				currentRow[j] = min(currentRow[j-1]+1, prevRow[j]+1, prevRow[j-1]+1)
			}
		}
		// 交换当前行和前一行
		prevRow, currentRow = currentRow, prevRow
	}
	return prevRow[n] // 返回将 s1 转换为 s2 所需的最小操作数
}
