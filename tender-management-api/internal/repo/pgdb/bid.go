package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"tender-management-api/internal/common"
	"tender-management-api/internal/entity"
	"tender-management-api/internal/repo/repo_errors"
	"tender-management-api/pkg/postgres"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type BidRepo struct {
	*postgres.Postgres
}

func NewBidRepo(pgdb *postgres.Postgres) *BidRepo {
	return &BidRepo{pgdb}
}

func (r *BidRepo) CreateBid(ctx context.Context, input *entity.CreateBidInput) (uuid.UUID, error) {
	tx, err := r.Database.Begin()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return uuid.Nil, e
		}

		return uuid.Nil, err
	}

	createBidReq, args, _ := r.SqlBuilder.
		Insert("bid").
		Columns("status", "tender_id", "author_id", "author_type", "current_version").
		Values(common.Created, input.TenderId, input.AuthorId, input.AuthorType, 1).
		Suffix("RETURNING id").
		RunWith(tx).
		ToSql()

	var bidId uuid.UUID
	err = tx.QueryRow(createBidReq, args...).Scan(&bidId)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return uuid.Nil, err
		}

		return uuid.Nil, err
	}

	createVersionReq, args, _ := r.SqlBuilder.
		Insert("bid_version").
		Columns("name", "description", "version", "bid_id").
		Values(input.Name, input.Description, 1, bidId).
		RunWith(tx).
		ToSql()

	_, err = tx.Exec(createVersionReq, args...)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return uuid.Nil, e
		}

		return uuid.Nil, err
	}

	if err = tx.Commit(); err != nil {
		return uuid.Nil, err
	}

	return bidId, nil
}

func (r *BidRepo) GetBidById(ctx context.Context, id string) (*entity.Bid, error) {
	uuidForm, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	getBidReq, args, _ := r.SqlBuilder.
		Select("bid.id, bid_version.name, bid_version.description, bid.status, bid.decision, bid.tender_id, bid.author_id, bid.author_type, bid.created_at, bid.current_version").
		From("bid").
		InnerJoin("bid_version on bid.id = bid_version.bid_id and bid.current_version = bid_version.version").
		Where("bid.id = ?", uuidForm).
		ToSql()

	var bid entity.Bid
	var createdAt time.Time
	row := r.Database.QueryRow(getBidReq, args...)
	err = row.Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.Decision,
		&bid.TenderId, &bid.AuthorId, &bid.AuthorType, &createdAt, &bid.Version)
	bid.CreatedAt = createdAt.Format(time.RFC3339)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &bid, repo_errors.ErrNotFound
		}

		return &bid, err
	}

	return &bid, nil
}

func (r *BidRepo) EditBidById(ctx context.Context, id string, name string, description string) error {
	uuidForm, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	tx, err := r.Database.BeginTx(ctx, nil)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	updateVersionReq, args, _ := r.SqlBuilder.
		Update("bid").
		Set("current_version", squirrel.Expr("current_version + ?", 1)).
		Where("id = ?", uuidForm).
		Suffix("RETURNING current_version").
		RunWith(tx).
		ToSql()

	var current_version int
	err = tx.QueryRow(updateVersionReq, args...).Scan(&current_version)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	getOldValuesReq, args, _ := r.SqlBuilder.
		Select("name", "description").
		From("bid_version").
		Where("bid_id = ?", uuidForm).
		Where("version = ?", current_version-1).
		RunWith(tx).
		ToSql()

	var prevName, prevDescription string
	if err = tx.QueryRow(getOldValuesReq, args...).
		Scan(&prevName, &prevDescription); err != nil {
		return err
	}

	if name == "" {
		name = prevName
	}

	if description == "" {
		description = prevDescription
	}

	createVersionReq, args, _ := r.SqlBuilder.
		Insert("bid_version").
		Columns("name", "description", "version", "bid_id").
		Values(name, description, current_version, uuidForm).
		RunWith(tx).
		ToSql()

	_, err = tx.Exec(createVersionReq, args...)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *BidRepo) UpdateBidStatusById(ctx context.Context, id string, newStatus string) error {
	uuidForm, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	updateStatusSql, args, _ := r.SqlBuilder.
		Update("bid").
		Set("status", newStatus).
		Where("id = ?", uuidForm).
		ToSql()

	_, err = r.Database.Exec(updateStatusSql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *BidRepo) GetUserBids(ctx context.Context, employeeId string, pg *entity.PaginationInput) ([]entity.Bid, error) {
	uuidForm, err := uuid.Parse(employeeId)
	if err != nil {
		return nil, err
	}

	getUserBidsReq, args, _ := r.SqlBuilder.
		Select("bid.id, bid_version.name, bid_version.description, bid.status, bid.decision, bid.tender_id, bid.author_id, bid.author_type, bid.created_at, bid.current_version").
		From("bid").
		InnerJoin("bid_version on bid.id = bid_version.bid_id and bid.current_version = bid_version.version").
		Where("author_id = ?", uuidForm).
		OrderBy("name ASC").
		Offset(uint64(pg.Offset)).
		Limit(uint64(pg.Limit)).
		ToSql()

	rows, err := r.Database.Query(getUserBidsReq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bids := make([]entity.Bid, 0)
	for rows.Next() {
		var createdAt time.Time
		var bid entity.Bid
		if err := rows.Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.Decision,
			&bid.TenderId, &bid.AuthorId, &bid.AuthorType, &createdAt, &bid.Version); err != nil {
			return bids, err
		}
		bid.CreatedAt = createdAt.Format(time.RFC3339)
		bids = append(bids, bid)
	}
	if err = rows.Err(); err != nil {
		return bids, err
	}

	return bids, nil
}

func (r *BidRepo) GetTenderBids(ctx context.Context, tenderId string, pg *entity.PaginationInput) ([]entity.Bid, error) {
	uuidForm, err := uuid.Parse(tenderId)
	if err != nil {
		return nil, err
	}

	getTenderBidsSql, args, _ := r.SqlBuilder.
		Select("bid.id, bid_version.name, bid_version.description, bid.status, bid.decision, bid.tender_id, bid.author_id, bid.author_type, bid.created_at, bid.current_version").
		From("bid").
		InnerJoin("bid_version on bid.id = bid_version.bid_id and bid.current_version = bid_version.version").
		Where("tender_id = ?", uuidForm).
		OrderBy("name ASC").
		Offset(uint64(pg.Offset)).
		Limit(uint64(pg.Limit)).
		ToSql()

	rows, err := r.Database.Query(getTenderBidsSql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bids := make([]entity.Bid, 0)
	for rows.Next() {
		var bid entity.Bid
		var createdAt time.Time
		if err := rows.Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.Decision,
			&bid.TenderId, &bid.AuthorId, &bid.AuthorType, &createdAt, &bid.Version); err != nil {
			return bids, err
		}
		bid.CreatedAt = createdAt.Format(time.RFC3339)
		bids = append(bids, bid)
	}
	if err = rows.Err(); err != nil {
		return bids, err
	}

	return bids, nil
}

func submitReject(r *BidRepo, bidId uuid.UUID) error {
	tx, err := r.Database.Begin()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	updateDecisionSql, args, _ := r.SqlBuilder.
		Update("bid").
		Set("decision", common.RejectedDecision).
		Where("id = ?", bidId).
		RunWith(tx).
		ToSql()

	if _, err := tx.Exec(updateDecisionSql, args...); err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	deleteApprovesSql, args, _ := r.SqlBuilder.
		Delete("approves").
		Where("bid_id = ?", bidId).
		RunWith(tx).
		ToSql()

	if _, err := tx.Exec(deleteApprovesSql, args...); err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *BidRepo) AlreadySubmitApprove(ctx context.Context, bidId string, employeeId string) (bool, error) {
	bidUuid, err := uuid.Parse(bidId)
	if err != nil {
		return false, err
	}

	employeeUuid, err := uuid.Parse(employeeId)
	if err != nil {
		return false, err
	}

	sqlReq, args, _ := r.SqlBuilder.
		Select("id").
		From("approves").
		Where("bid_id = ?", bidUuid).
		Where("employee_id = ?", employeeUuid).
		ToSql()

	var uid uuid.UUID
	err = r.Database.QueryRow(sqlReq, args...).Scan(&uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *BidRepo) SubmitBidDecision(ctx context.Context, bidId string, decision string, employeeId string, organizationId uuid.UUID) error {
	bidUuid, err := uuid.Parse(bidId)
	if err != nil {
		return err
	}

	employeeUuid, err := uuid.Parse(employeeId)
	if err != nil {
		return err
	}

	if decision == common.RejectedDecision {
		return submitReject(r, bidUuid)
	}

	tx, err := r.Database.Begin()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	approvesCntSql, args, _ := r.SqlBuilder.
		Select("count(*)").
		From("approves").
		Where("bid_id = ?", bidId).
		RunWith(tx).
		ToSql()

	var approvesCnt int
	if err = tx.QueryRow(approvesCntSql, args...).Scan(&approvesCnt); err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	countSql, args, _ := r.SqlBuilder.
		Select("count(*)").
		From("organization_responsible").
		Where("organization_id = ?", organizationId).
		RunWith(tx).
		ToSql()

	var responsibleCnt int
	if err = tx.QueryRow(countSql, args...).Scan(&responsibleCnt); err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	quorum := min(responsibleCnt, 3)
	if approvesCnt < quorum-1 {
		addApproveSql, args, _ := r.SqlBuilder.
			Insert("approves").
			Columns("bid_id", "employee_id").
			Values(bidUuid, employeeUuid).
			RunWith(tx).
			ToSql()

		if _, err := tx.Exec(addApproveSql, args...); err != nil {
			if e := tx.Rollback(); e != nil {
				return e
			}

			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	}

	updateDecisionSql, args, _ := r.SqlBuilder.
		Update("bid").
		Set("decision", common.ApprovedDecision).
		Where("id = ?", bidUuid).
		RunWith(tx).
		ToSql()

	if _, err := tx.Exec(updateDecisionSql, args...); err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	deleteApprovesSql, args, _ := r.SqlBuilder.
		Delete("approves").
		Where("bid_id = ?", bidUuid).
		ToSql()

	if _, err := tx.Exec(deleteApprovesSql, args...); err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *BidRepo) RollbackBidVersion(ctx context.Context, bidId string, version int) error {
	uuidForm, err := uuid.Parse(bidId)
	if err != nil {
		return err
	}

	tx, err := r.Database.BeginTx(ctx, nil)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	updateVersionInBidTableSql, args, _ := r.SqlBuilder.
		Update("bid").
		Set("current_version", squirrel.Expr("current_version + ?", 1)).
		Where("id = ?", uuidForm).
		Suffix("RETURNING current_version").
		RunWith(tx).
		ToSql()

	var currentVersion int
	err = tx.QueryRow(updateVersionInBidTableSql, args...).Scan(&currentVersion)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	updateVersionInVersionTableSql, args, _ := r.SqlBuilder.
		Update("bid_version").
		Set("version", currentVersion).
		Where("bid_id = ?", uuidForm).
		Where("version = ?", version).
		RunWith(tx).
		ToSql()

	if _, err = tx.Exec(updateVersionInVersionTableSql, args...); err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *BidRepo) SubmitBidFeedBack(ctx context.Context, bidId string, senderId uuid.UUID, receiverId uuid.UUID, content string) error {
	createFeedbackReq, args, _ := r.SqlBuilder.
		Insert("review").
		Columns("bid_id", "receiver_id", "author_id", "description").
		Values(bidId, receiverId, senderId, content).
		ToSql()

	_, err := r.Database.Exec(createFeedbackReq, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *BidRepo) GetReviewsByReceiverId(ctx context.Context, receiverId string, pg *entity.PaginationInput) ([]entity.Review, error) {
	getTenderBidsSql, args, _ := r.SqlBuilder.
		Select("id, description, created_at, author_id, receiver_id, bid_id").
		From("review").
		Where("receiver_id = ?", receiverId).
		OrderBy("description ASC").
		Offset(uint64(pg.Offset)).
		Limit(uint64(pg.Limit)).
		ToSql()

	rows, err := r.Database.Query(getTenderBidsSql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]entity.Review, 0)
	for rows.Next() {
		var review entity.Review
		var createdAt time.Time
		if err := rows.Scan(&review.Id, &review.Description, &createdAt, &review.AuthorId, &review.ReceiverId, &review.BidId); err != nil {
			return reviews, err
		}
		review.CreatedAt = createdAt.Format(time.RFC3339)
		reviews = append(reviews, review)
	}
	if err = rows.Err(); err != nil {
		return reviews, err
	}

	return reviews, nil
}
