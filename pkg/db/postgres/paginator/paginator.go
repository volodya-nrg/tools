package paginator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

var ErrMissingField = errors.New("missing field")

func Prepare(
	fieldNameForOrder,
	filterQuery,
	filterField string,
	limitSrc,
	offset int32,
	isSortByAsc bool,
) (string, string, pgx.StrictNamedArgs, error) {
	strBuilder := strings.Builder{}
	sort := "DESC"
	namedArgs := pgx.StrictNamedArgs{}

	if fieldNameForOrder == "" {
		return "", "", nil, ErrMissingField
	}
	if filterQuery != "" {
		if filterField == "" {
			return "", "", nil, ErrMissingField
		}

		strBuilder.WriteString(fmt.Sprintf("WHERE %s ILIKE '%%' || @query || '%%'", filterField))
		namedArgs["query"] = filterQuery
	}
	if isSortByAsc {
		sort = "ASC"
	}

	resultWithoutLimit := strBuilder.String()

	strBuilder.WriteString(fmt.Sprintf("ORDER BY %s %s ", fieldNameForOrder, sort))

	if limitSrc > 0 {
		strBuilder.WriteString(fmt.Sprintf("LIMIT %d ", limitSrc))
	}
	if offset > 0 {
		strBuilder.WriteString(fmt.Sprintf("OFFSET %d ", offset))
	}

	resultWithLimit := strBuilder.String()

	return resultWithLimit, resultWithoutLimit, namedArgs, nil
}
