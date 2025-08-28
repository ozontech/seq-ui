package sql_builder

import sq "github.com/Masterminds/squirrel"

func Select(columns ...string) sq.SelectBuilder {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select(columns...)
}

func Insert(into string) sq.InsertBuilder {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Insert(into)
}

func Update(table string) sq.UpdateBuilder {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Update(table)
}

func Delete(from string) sq.DeleteBuilder {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Delete(from)
}
