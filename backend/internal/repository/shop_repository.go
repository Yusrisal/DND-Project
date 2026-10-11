package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"dndproject/backend/internal/model"
)

type ShopRepository struct {
	db *sql.DB
}

func NewShopRepository(db *sql.DB) *ShopRepository {
	return &ShopRepository{db: db}
}

var shopSortColumns = []string{"id", "name", "created_at", "updated_at"}

func ShopSortColumns() []string { return shopSortColumns }

const shopColumns = `id, campaign_id, name, description, created_at, updated_at`
const shopItemColumns = `id, shop_id, name, cost_cp, quantity, weight, attack_bonus, damage, damage_type, description, created_at, updated_at`

func scanShop(row interface{ Scan(...any) error }) (model.Shop, error) {
	var s model.Shop
	err := row.Scan(&s.ID, &s.CampaignID, &s.Name, &s.Description, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func scanShopItem(row interface{ Scan(...any) error }) (model.ShopItem, error) {
	var i model.ShopItem
	err := row.Scan(&i.ID, &i.ShopID, &i.Name, &i.CostCP, &i.Quantity, &i.Weight,
		&i.AttackBonus, &i.Damage, &i.DamageType, &i.Description, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

func (r *ShopRepository) List(ctx context.Context, campaignID int64, q model.Query) ([]model.Shop, int64, error) {
	where, args := "WHERE 1 = 1", []any{}
	if campaignID > 0 {
		where += " AND campaign_id = ?"
		args = append(args, campaignID)
	}
	if q.Search != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM shops "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM shops %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		shopColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Shop
	for rows.Next() {
		s, err := scanShop(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, rows.Err()
}

func (r *ShopRepository) Get(ctx context.Context, id int64) (model.Shop, error) {
	return scanShop(r.db.QueryRowContext(ctx, `SELECT `+shopColumns+` FROM shops WHERE id = ?`, id))
}

func (r *ShopRepository) Create(ctx context.Context, in model.ShopInput) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO shops (campaign_id, name, description) VALUES (?, ?, ?)`,
		in.CampaignID, in.Name, in.Description)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ShopRepository) Update(ctx context.Context, id int64, in model.ShopInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE shops SET name = ?, description = ? WHERE id = ?`,
		in.Name, in.Description, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *ShopRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM shops WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// Items

func (r *ShopRepository) Items(ctx context.Context, shopID int64) ([]model.ShopItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+shopItemColumns+` FROM shop_items WHERE shop_id = ? ORDER BY name`, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.ShopItem{}
	for rows.Next() {
		i, err := scanShopItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (r *ShopRepository) Item(ctx context.Context, id int64) (model.ShopItem, error) {
	return scanShopItem(r.db.QueryRowContext(ctx, `SELECT `+shopItemColumns+` FROM shop_items WHERE id = ?`, id))
}

func (r *ShopRepository) CreateItem(ctx context.Context, shopID int64, in model.ShopItemInput) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO shop_items (shop_id, name, cost_cp, quantity, weight, attack_bonus, damage, damage_type, description)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		shopID, in.Name, in.CostCP, in.Quantity, in.Weight, in.AttackBonus, in.Damage, in.DamageType, in.Description)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *ShopRepository) UpdateItem(ctx context.Context, id int64, in model.ShopItemInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE shop_items SET name = ?, cost_cp = ?, quantity = ?, weight = ?, attack_bonus = ?, damage = ?, damage_type = ?, description = ?
		 WHERE id = ?`,
		in.Name, in.CostCP, in.Quantity, in.Weight, in.AttackBonus, in.Damage, in.DamageType, in.Description, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *ShopRepository) DeleteItem(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM shop_items WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// BuyItem executes the purchase in a transaction: deducts money from character,
// creates character item, and decrements shop stock if limited.
func (r *ShopRepository) BuyItem(ctx context.Context, shopItemID, characterID int64, quantity int) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 1. Read shop item
	item, err := scanShopItem(tx.QueryRowContext(ctx, `SELECT `+shopItemColumns+` FROM shop_items WHERE id = ? FOR UPDATE`, shopItemID))
	if err != nil {
		return 0, err
	}
	if item.Quantity >= 0 && item.Quantity < quantity {
		return 0, errors.New("not enough stock in shop")
	}

	totalCost := item.CostCP * quantity

	// 2. Read character currency
	var cp, sp, ep, gp, pp int
	if err := tx.QueryRowContext(ctx, `SELECT cp, sp, ep, gp, pp FROM characters WHERE id = ? FOR UPDATE`, characterID).
		Scan(&cp, &sp, &ep, &gp, &pp); err != nil {
		return 0, err
	}

	charTotalCP := cp + sp*10 + ep*50 + gp*100 + pp*1000
	if charTotalCP < totalCost {
		return 0, errors.New("not enough money")
	}

	// 3. Deduct totalCost from coins starting from CP, then SP, then EP, then GP, then PP
	remaining := totalCost
	// Greedy deduction with change conversion
	if cp >= remaining {
		cp -= remaining
		remaining = 0
	} else {
		remaining -= cp
		cp = 0
		if sp*10 >= remaining {
			spSpend := (remaining + 9) / 10
			sp -= spSpend
			cp += (spSpend * 10) - remaining
			remaining = 0
		} else {
			remaining -= sp * 10
			sp = 0
			if gp*100 >= remaining {
				gpSpend := (remaining + 99) / 100
				gp -= gpSpend
				cp += (gpSpend * 100) - remaining
				remaining = 0
			} else {
				remaining -= gp * 100
				gp = 0
				ppSpend := (remaining + 999) / 1000
				pp -= ppSpend
				cp += (ppSpend * 1000) - remaining
				remaining = 0
			}
		}
	}

	// 4. Update character currency
	if _, err := tx.ExecContext(ctx, `UPDATE characters SET cp = ?, sp = ?, ep = ?, gp = ?, pp = ? WHERE id = ?`,
		cp, sp, ep, gp, pp, characterID); err != nil {
		return 0, err
	}

	// 5. Decrement shop item if not unlimited
	if item.Quantity > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE shop_items SET quantity = quantity - ? WHERE id = ?`, quantity, shopItemID); err != nil {
			return 0, err
		}
	}

	// 6. Insert into character_items
	res, err := tx.ExecContext(ctx,
		`INSERT INTO character_items (character_id, name, quantity, weight, cost_cp, equipped, attack_bonus, damage, damage_type, description)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		characterID, item.Name, quantity, item.Weight, item.CostCP, 0,
		item.AttackBonus, item.Damage, item.DamageType, item.Description)
	if err != nil {
		return 0, err
	}

	createdID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return createdID, tx.Commit()
}

// SellItem handles selling an item back to the merchant for 50% of cost_cp (or at least 1 CP).
// It reduces quantity or deletes the item, and credits coins to the character.
func (r *ShopRepository) SellItem(ctx context.Context, characterID, itemID int64, quantity int) (model.SellItemResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.SellItemResult{}, err
	}
	defer tx.Rollback()

	// 1. Get character item
	var name string
	var costCP, currentQty int
	err = tx.QueryRowContext(ctx,
		`SELECT name, cost_cp, quantity FROM character_items WHERE id = ? AND character_id = ? FOR UPDATE`,
		itemID, characterID).Scan(&name, &costCP, &currentQty)
	if err != nil {
		return model.SellItemResult{}, err
	}
	if quantity <= 0 || quantity > currentQty {
		quantity = currentQty
	}

	unitSellPrice := costCP / 2
	if unitSellPrice <= 0 && costCP > 0 {
		unitSellPrice = 1 // min 1 cp if item has any value
	}
	totalRefundCP := unitSellPrice * quantity

	// 2. Decrement or delete item
	if quantity >= currentQty {
		if _, err := tx.ExecContext(ctx, `DELETE FROM character_items WHERE id = ?`, itemID); err != nil {
			return model.SellItemResult{}, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `UPDATE character_items SET quantity = quantity - ? WHERE id = ?`, quantity, itemID); err != nil {
			return model.SellItemResult{}, err
		}
	}

	// 3. Add refund coins to character (prefer GP/SP/CP distribution)
	var cp, sp, ep, gp, pp int
	if err := tx.QueryRowContext(ctx, `SELECT cp, sp, ep, gp, pp FROM characters WHERE id = ? FOR UPDATE`, characterID).
		Scan(&cp, &sp, &ep, &gp, &pp); err != nil {
		return model.SellItemResult{}, err
	}

	// Convert refund to GP, SP, CP
	refundGP := totalRefundCP / 100
	remainderCP := totalRefundCP % 100
	refundSP := remainderCP / 10
	refundCP := remainderCP % 10

	gp += refundGP
	sp += refundSP
	cp += refundCP

	if _, err := tx.ExecContext(ctx, `UPDATE characters SET cp = ?, sp = ?, gp = ? WHERE id = ?`,
		cp, sp, gp, characterID); err != nil {
		return model.SellItemResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.SellItemResult{}, err
	}

	return model.SellItemResult{
		ItemName:     name,
		QuantitySold: quantity,
		RefundCP:     totalRefundCP,
		CP:           cp,
		SP:           sp,
		EP:           ep,
		GP:           gp,
		PP:           pp,
	}, nil
}
