package valet

import (
	"context"
	"strings"
	"sync"
)

// ============================================================================
// OBJECT POOLS FOR PERFORMANCE
// ============================================================================

// dbCheckSlicePool reuses DB check slices
var dbCheckSlicePool = sync.Pool{
	New: func() any {
		slice := make([]DBCheck, 0, 16)
		return &slice
	},
}

func getDBCheckSlice() *[]DBCheck {
	s := dbCheckSlicePool.Get().(*[]DBCheck)
	*s = (*s)[:0]
	return s
}

func releaseDBCheckSlice(s *[]DBCheck) {
	if s != nil && cap(*s) <= 256 { // Don't pool very large slices
		dbCheckSlicePool.Put(s)
	}
}

// batchGroup holds checks for a single table+column+where combination
type batchGroup struct {
	table  string
	column string
	wheres []WhereClause
	checks []DBCheck
	values []any
}

// batchGroupPool reuses batchGroup instances
var batchGroupPool = sync.Pool{
	New: func() any {
		return &batchGroup{
			checks: make([]DBCheck, 0, 8),
			values: make([]any, 0, 8),
		}
	},
}

func getBatchGroup() *batchGroup {
	g := batchGroupPool.Get().(*batchGroup)
	g.checks = g.checks[:0]
	g.values = g.values[:0]
	g.wheres = nil
	g.table = ""
	g.column = ""
	return g
}

func releaseBatchGroup(g *batchGroup) {
	if g != nil && cap(g.checks) <= 64 {
		batchGroupPool.Put(g)
	}
}

// batchKeyPool reuses strings.Builder for batch key generation
var batchKeyPool = sync.Pool{
	New: func() any {
		return &strings.Builder{}
	},
}

// makeBatchKey creates a unique key for grouping similar checks
func makeBatchKey(table, column string, wheres []WhereClause) string {
	if len(wheres) == 0 {
		return table + ":" + column
	}

	sb := batchKeyPool.Get().(*strings.Builder)
	sb.Reset()
	defer batchKeyPool.Put(sb)

	sb.WriteString(table)
	sb.WriteByte(':')
	sb.WriteString(column)

	for _, w := range wheres {
		sb.WriteByte(':')
		sb.WriteString(w.Column)
		sb.WriteString(w.Operator)
	}

	return sb.String()
}

// ============================================================================
// VALIDATION FUNCTIONS
// ============================================================================

// Validate validates data against a schema
func Validate(data DataObject, schema Schema, opts ...Options) *ValidationError {
	var options Options
	if len(opts) > 0 {
		options = opts[0]
	}

	ctx := &ValidationContext{
		Ctx:      options.Context,
		RootData: data,
		Path:     []string{},
		Options:  &options,
	}

	if ctx.Ctx == nil {
		ctx.Ctx = context.Background()
	}

	allErrors := make(map[string][]string)

	// Get pooled slice for DB checks
	dbChecksPtr := getDBCheckSlice()
	defer releaseDBCheckSlice(dbChecksPtr)
	dbChecks := dbChecksPtr

	// Validate each field
	for field, validator := range schema {
		fieldCtx := &ValidationContext{
			Ctx:      ctx.Ctx,
			RootData: data,
			Path:     []string{field},
			Options:  ctx.Options,
		}

		value := data[field]
		fieldErrors := validator.Validate(fieldCtx, value)

		// Merge field errors into allErrors
		for path, errs := range fieldErrors {
			allErrors[path] = append(allErrors[path], errs...)
		}

		if len(fieldErrors) > 0 && options.AbortEarly {
			return &ValidationError{Errors: allErrors}
		}

		// Collect DB checks
		if collector, ok := validator.(DBCheckCollector); ok {
			checks := collector.GetDBChecks(field, value)
			*dbChecks = append(*dbChecks, checks...)
		}
	}

	// Execute DB checks if we have a checker and no errors so far
	if options.DBChecker != nil && len(*dbChecks) > 0 && len(allErrors) == 0 {
		dbErrors := executeBatchedDBChecks(ctx.Ctx, options.DBChecker, *dbChecks)
		for field, errs := range dbErrors {
			allErrors[field] = errs
		}
	}

	if len(allErrors) > 0 {
		return &ValidationError{Errors: allErrors}
	}

	return nil
}

// Parse is an alias for Validate (Zod-like naming)
func Parse(data DataObject, schema Schema, opts ...Options) *ValidationError {
	return Validate(data, schema, opts...)
}

// SafeParse returns (data, error) instead of just error
func SafeParse(data DataObject, schema Schema, opts ...Options) (DataObject, *ValidationError) {
	err := Validate(data, schema, opts...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ValidateWithDB validates data with database checks using provided DBChecker
func ValidateWithDB(ctx context.Context, data DataObject, schema Schema, checker DBChecker) *ValidationError {
	return Validate(data, schema, Options{
		Context:   ctx,
		DBChecker: checker,
	})
}

// ValidateWithDBContext validates data with full options including DB checker
func ValidateWithDBContext(ctx context.Context, data DataObject, schema Schema, opts Options) (DataObject, error) {
	opts.Context = ctx
	err := Validate(data, schema, opts)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ============================================================================
// BATCHED DB CHECK EXECUTION WITH PARALLEL QUERIES
// ============================================================================

// batchResult holds the result of a single batch query
type batchResult struct {
	group     *batchGroup
	existsMap map[any]bool
	err       error
}

// executeBatchedDBChecks runs all DB checks with batching and parallel execution
func executeBatchedDBChecks(ctx context.Context, checker DBChecker, checks []DBCheck) map[string][]string {
	if len(checks) == 0 {
		return nil
	}

	// Group checks by table+column+where for batching
	// Pre-allocate with estimated size
	groups := make(map[string]*batchGroup, len(checks)/2+1)
	groupList := make([]*batchGroup, 0, len(checks)/2+1) // Track for cleanup

	for _, check := range checks {
		key := makeBatchKey(check.Rule.Table, check.Rule.Column, check.Rule.Where)
		if groups[key] == nil {
			g := getBatchGroup()
			g.table = check.Rule.Table
			g.column = check.Rule.Column
			g.wheres = check.Rule.Where
			groups[key] = g
			groupList = append(groupList, g)
		}
		groups[key].checks = append(groups[key].checks, check)
		groups[key].values = append(groups[key].values, check.Value)
	}

	// Defer cleanup of all groups
	defer func() {
		for _, g := range groupList {
			releaseBatchGroup(g)
		}
	}()

	errs := make(map[string][]string)

	// For single group, execute directly (no goroutine overhead)
	if len(groups) == 1 {
		for _, group := range groups {
			existsMap, err := checker.CheckExists(ctx, group.table, group.column, group.values, group.wheres)
			processGroupResult(group, existsMap, err, errs)
		}
		return errs
	}

	// Multiple groups: execute in parallel
	results := make(chan batchResult, len(groups))
	var wg sync.WaitGroup

	for _, group := range groups {
		wg.Add(1)
		go func(g *batchGroup) {
			defer wg.Done()
			// Check context cancellation before executing
			select {
			case <-ctx.Done():
				results <- batchResult{group: g, existsMap: nil, err: ctx.Err()}
				return
			default:
			}
			existsMap, err := checker.CheckExists(ctx, g.table, g.column, g.values, g.wheres)
			results <- batchResult{group: g, existsMap: existsMap, err: err}
		}(group)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		processGroupResult(result.group, result.existsMap, result.err, errs)
	}

	return errs
}

// processGroupResult processes the result of a single batch query
func processGroupResult(group *batchGroup, existsMap map[any]bool, err error, errs map[string][]string) {
	if err != nil {
		// On DB error, add error to all fields in this group
		for _, check := range group.checks {
			errs[check.Field] = append(errs[check.Field], "database error: "+err.Error())
		}
		return
	}

	for _, check := range group.checks {
		exists := existsMap[check.Value]

		// Create message context for resolving dynamic messages
		// Note: Data is nil here as DB checks don't have access to root data
		msgCtx := MessageContext{
			Field: check.Field,
			Path:  check.Field,
			Index: extractIndex(check.Field),
			Value: check.Value,
		}

		if check.IsUnique {
			// For unique: should NOT exist (unless it's the ignored value)
			if exists && check.Value != check.Ignore {
				var errMsg string
				if check.Message != nil {
					errMsg = resolveMessage(check.Message, msgCtx)
				} else {
					errMsg = check.Field + " already exists"
				}
				errs[check.Field] = append(errs[check.Field], errMsg)
			}
		} else {
			// For exists: should exist
			if !exists {
				var errMsg string
				if check.Message != nil {
					errMsg = resolveMessage(check.Message, msgCtx)
				} else {
					errMsg = check.Field + " does not exist"
				}
				errs[check.Field] = append(errs[check.Field], errMsg)
			}
		}
	}
}
