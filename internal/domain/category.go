package domain
import "time"
type Category struct{
	CatId string
	CatName string
	Description string
	CreateAt time.Time
	UpdateAt time.Time
}
