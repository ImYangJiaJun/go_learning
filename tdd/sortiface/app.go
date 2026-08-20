package sortiface

// App 是排行榜里的一个应用。两个字段都是 int，因此 App 可比较（支持 == / !=）
type App struct {
	ID        int // 应用唯一标识
	Downloads int // 累计下载量
}

// ByDownloads 让 []App 具备"按下载量降序"的排序能力。
// 实现 sort.Interface：Less(i, j) 返回 true 表示第 i 个元素应排在第 j 个前面
type ByDownloads []App

func (a ByDownloads) Len() int { // 固定套路：return len(a)
	return len(a)
}
func (a ByDownloads) Less(i, j int) bool { // 本类型唯一有语义的方法：下载量大的在前（降序）
	return a[i].Downloads > a[j].Downloads
}
func (a ByDownloads) Swap(i, j int) { // 固定套路：a[i], a[j] = a[j], a[i]
	a[i], a[j] = a[j], a[i]
}

// ByID 让 []App 具备"按 ID 升序"的排序能力，同样实现 sort.Interface
type ByID []App

func (a ByID) Len() int { // 固定套路
	return len(a)
}
func (a ByID) Less(i, j int) bool { // ID 小的在前（升序）
	return a[i].ID < a[j].ID
}
func (a ByID) Swap(i, j int) { // 固定套路
	a[i], a[j] = a[j], a[i]
}
