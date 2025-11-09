package paginator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func Prepare(
	fieldNameForOrder,
	filterQuery,
	filterField string,
	limit,
	offset int32,
	isSortByAsc bool,
) (string, string, pgx.StrictNamedArgs, error) {
	strBuilder := strings.Builder{}
	sort := "DESC"
	namedArgs := pgx.StrictNamedArgs{}

	if fieldNameForOrder == "" {
		return "", "", nil, errors.New("field name for order is required")
	}
	if filterQuery != "" {
		if filterField == "" {
			return "", "", nil, errors.New("filter-field is required")
		}

		strBuilder.WriteString(fmt.Sprintf("WHERE %s ILIKE '%%' || @query || '%%'", filterField))
		namedArgs["query"] = filterQuery
	}
	if isSortByAsc {
		sort = "ASC"
	}

	resultWithoutLimit := strBuilder.String()

	strBuilder.WriteString(fmt.Sprintf("ORDER BY %s %s ", fieldNameForOrder, sort))

	if limit > 0 {
		strBuilder.WriteString(fmt.Sprintf("LIMIT %d ", limit))
	}
	if offset > 0 {
		strBuilder.WriteString(fmt.Sprintf("OFFSET %d ", offset))
	}

	resultWithLimit := strBuilder.String()

	return resultWithLimit, resultWithoutLimit, namedArgs, nil
}
