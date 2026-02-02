package base_table

import (
	"context"
	"strings"

	transactorX "github.com/volodya-nrg/tools/pkg/db/sqlite/transactor"
)

type Transactor interface {
	Conn(ctx context.Context) transactorX.Connection
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type BaseTable struct {
	Transactor Transactor
	TblName    string
	Fields     []string
}

func (b *BaseTable) Markers(amount int) string {
	return strings.Join(strings.Split(strings.Repeat("?", amount), ""), ",")
}

func NewBaseTbl(transactor Transactor, tblName string, fields []string) BaseTable {
	return BaseTable{
		Transactor: transactor,
		TblName:    tblName,
		Fields:     fields,
	}
}
