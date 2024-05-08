package quintus

import (
	"hago-plat/pcom/nameconv"
	"strings"
)

// ObjectName 对象名, 需要输入驼峰规则的名
type ObjectName struct {
	nameconv.Name
}

// PureLower 生成全小写不带下划线的名称
func (name ObjectName) PureLower() string {
	result := strings.ToLower(string(name.Name))
	return strings.Replace(result, "_", "", -1)
}

// Data 导入template的数据
type Data struct {
	Names  map[string]ObjectName
	Values map[string]string
}
