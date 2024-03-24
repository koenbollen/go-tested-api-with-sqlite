package steps

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
)

type DatabaseSteps struct {
	DB *sql.DB
}

func (s *DatabaseSteps) InitializeSuite(suite *godog.TestSuiteContext) error {
	return nil
}

func (s *DatabaseSteps) InitializeScenario(scenario *godog.ScenarioContext) error {
	scenario.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		return ctx, nil
	})
	scenario.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		return ctx, nil
	})

	scenario.Step(`^the follow "([^"]*)" record exist:$`, s.GivenTheRecordExist)
	scenario.Step(`^(?:these|this) "([^"]*)" records exist:$`, s.GivenTheseRecordsExist)

	scenario.Step(`^this "([^"]*)" record exists:$`, s.ThenThisRecordExists)
	scenario.Step(`^no "([^"]*)" record exists with id "([^"]*)"$`, s.ThenNoRecordExists)

	return nil
}

func (s *DatabaseSteps) CreateRecord(ctx context.Context, table string, data map[string]any) error {
	columns := []string{}
	placeholders := []string{}
	values := []any{}
	seen := map[string]bool{}
	for key, value := range data {
		columns = append(columns, key)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(columns)))
		if value == "<nil>" {
			value = nil
		}
		values = append(values, value)
		seen[key] = true
	}
	_, err := s.DB.ExecContext(ctx, `INSERT INTO "`+table+`" ("`+strings.Join(columns, `","`)+`") VALUES (`+strings.Join(placeholders, ",")+`)`, values...)
	return err
}

func (s *DatabaseSteps) GivenTheseRecordsExist(ctx context.Context, table string, record *godog.Table) error {
	keys := []string{}
	for _, cell := range record.Rows[0].Cells {
		keys = append(keys, cell.Value)
	}
	for _, row := range record.Rows[1:] {
		data := make(map[string]any)
		for i, cell := range row.Cells {
			data[keys[i]] = cell.Value
		}
		if err := s.CreateRecord(ctx, table, data); err != nil {
			return err
		}
	}
	return nil
}

func (s *DatabaseSteps) GivenTheRecordExist(ctx context.Context, table string, record *godog.Table) error {
	data := make(map[string]any)
	for _, row := range record.Rows {
		key := row.Cells[0].Value
		value := row.Cells[1].Value
		data[key] = value
	}
	return s.CreateRecord(ctx, table, data)
}

func (s *DatabaseSteps) ThenThisRecordExists(ctx context.Context, table string, record *godog.Table) error {
	primarykeys, err := s.fetchPrimaryKeys(ctx, table)
	if err != nil {
		return err
	}
	if len(primarykeys) == 0 {
		return fmt.Errorf("no primary key found for table %s", table)
	}

	columns := []string{}
	primaryvalues := make([]any, len(primarykeys))
	for _, row := range record.Rows {
		columns = append(columns, row.Cells[0].Value)
		if ix := sort.SearchStrings(primarykeys, row.Cells[0].Value); ix < len(primarykeys) && primarykeys[ix] == row.Cells[0].Value {
			var val any = row.Cells[1].Value
			if n, ok := strconv.ParseInt(row.Cells[1].Value, 10, 64); ok == nil {
				val = n
			}
			primaryvalues[ix] = val
		}
	}

	q := "SELECT " + strings.Join(columns, ", ") + " FROM " + table + " WHERE 1=1"
	for i, pk := range primarykeys {
		q += " AND " + pk + " = $" + strconv.Itoa(i+1)
	}
	row := s.DB.QueryRowContext(ctx, q, primaryvalues...)
	var values []interface{}
	for i := 0; i < len(columns); i++ {
		var v any
		values = append(values, &v)
	}
	if err := row.Scan(values...); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no record found for table %s with primary keys %v", table, primarykeys)
		}
		return err
	}

	transformJSON := cmp.FilterValues(func(x, y string) bool {
		return json.Valid([]byte(x)) && json.Valid([]byte(y))
	}, cmp.Transformer("ParseJSON", func(in string) (out interface{}) {
		if err := json.Unmarshal([]byte(in), &out); err != nil {
			panic(err)
		}
		return out
	}))

	for i, row := range record.Rows {
		want := row.Cells[1].Value
		var got any
		v := *values[i].(*any)
		switch v := v.(type) {
		case string:
			got = v
		case [16]uint8: // uuid
			uuid, err := uuid.FromBytes(v[:])
			if err != nil {
				got = fmt.Sprintf("%v", v)
			} else {
				got = uuid.String()
			}
		case []uint8: // bytea
			got = string(v[:])
		case map[string]any: // jsonb
			raw, err := json.Marshal(v)
			if err != nil {
				return err
			}
			got = string(raw)
		case time.Time:
			got = v.UTC().Format(time.RFC3339)
		case netip.Prefix:
			got = v.Addr().String()
		default:
			got = fmt.Sprintf("%v", v)
		}
		if diff := cmp.Diff(want, got, transformJSON); diff != "" {
			return fmt.Errorf("column %q mismatched (-want, +got):\n%s", row.Cells[0].Value, diff)
		}
	}
	return nil
}

func (s *DatabaseSteps) ThenNoRecordExists(ctx context.Context, table, id string) error {
	q := `SELECT COUNT(*) FROM ` + table + ` WHERE "id" = $1`
	row := s.DB.QueryRowContext(ctx, q, id)
	var count int
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("expected no record with id %q, got %d", id, count)
	}
	return nil
}

func (s *DatabaseSteps) fetchPrimaryKeys(ctx context.Context, table string) ([]string, error) {
	var primarykeys []string
	pkquery := `
		SELECT l.name
		FROM pragma_table_info("` + table + `") AS l
		WHERE l.pk = 1
	`
	rows, err := s.DB.QueryContext(ctx, pkquery)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var pk string
		if err := rows.Scan(&pk); err != nil {
			return nil, err
		}
		primarykeys = append(primarykeys, pk)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Strings(primarykeys)
	return primarykeys, nil
}
