package store

import (
	"github.com/goravel/framework/contracts/database/orm"

	"dev.jevido/jevidocs/services/api/app/facades"
	"dev.jevido/jevidocs/services/api/app/models"
)

// OrderItem is one page's new position.
type OrderItem struct {
	ID       uint `json:"id"`
	Position int  `json:"position"`
}

// Reorder sets the positions of pages of p in one transaction. Items that
// are not pages of p fail the whole call, so nothing is half applied.
func Reorder(p models.Project, items []OrderItem) error {
	if len(items) == 0 {
		return ValidationError{"items is empty"}
	}
	if len(items) > 500 {
		return ValidationError{"too many items"}
	}
	err := facades.Orm().Transaction(func(tx orm.Query) error {
		for _, it := range items {
			res, err := tx.Exec(`UPDATE pages SET position = ?, updated_at = now() WHERE id = ? AND project_id = ?`,
				it.Position, it.ID, p.ID)
			if err != nil {
				return err
			}
			if res.RowsAffected == 0 {
				return ValidationError{"a page in items does not belong to this project"}
			}
		}
		return nil
	})
	if err == nil {
		touch(p)
	}
	return err
}
