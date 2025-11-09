package base_table

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	transactorX "github.com/volodya-nrg/tools/pkg/db/postgres/transactor"
)

type transactor interface {
	Conn(ctx context.Context) transactorX.Connection
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type BaseTable struct {
	Transactor transactor
	TblName    string
	Fields     []string
}

func (b *BaseTable) Columns(fields []string, addMarkers, addFirstField bool) []string {
	result := make([]string, 0, len(fields))
	j := 0

	for i, field := range fields {
		if i == 0 && !addFirstField {
			continue
		}

		var suffix string
		if addMarkers {
			suffix = fmt.Sprintf("=$%d", j+1)
			j++
		}

		result = append(result, field+suffix)
	}

	return result
}

func (b *BaseTable) Markers(amount int) string {
	str := make([]string, amount)
	for i := range amount {
		str[i] = "$" + strconv.Itoa(i+1)
	}

	return strings.Join(str, ",")
}

func (b *BaseTable) Total(
	ctx context.Context,
	tableName,
	countValue string,
	sqlSuffix string,
	queryTimeout time.Duration,
	params ...any,
) (uint32, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var (
		result uint32
		query  = fmt.Sprintf(`SELECT COUNT(%s) FROM %s %s`, countValue, tableName, sqlSuffix)
	)

	err := b.Transactor.
		Conn(ctx).
		QueryRow(ctx, query, params...).
		Scan(&result)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("failed to scan: %w", err)
	}

	return result, nil
}

func NewBaseTable(transactor transactor, tblName string, fields []string) BaseTable {
	return BaseTable{
		Transactor: transactor,
		TblName:    tblName,
		Fields:     fields,
	}
}
