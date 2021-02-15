package roshi

// 最終変更時刻を記録しているもの (map[string]string へのエイリアス)
type MTRecord map[string]string

func (m MTRecord) FileModified(filename, modtime string) bool {
	lastmodtime, ok := m[filename]
	return !ok || lastmodtime != modtime
}
