package main

type Inventory struct {
	// 自行补充字段
}

// 假定初始库存非负。
func NewInventory(stock int) *Inventory {
	// TODO
	return nil
}

func (i *Inventory) Reserve(requestID string, qty int) error {
	// TODO
	return nil
}

func (i *Inventory) Remaining() int {
	// TODO
	return 0
}
