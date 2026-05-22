package cart

type UpsertItemCommand struct {
	UserID    int64
	ProductID int64
	Quantity  int
}

type ListItemsCommand struct {
	UserID int64
}

type RemoveItemCommand struct {
	UserID    int64
	ProductID int64
}
