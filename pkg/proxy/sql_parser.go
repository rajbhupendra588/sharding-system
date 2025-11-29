package proxy

import (
	"fmt"
	"regexp"
	"strings"
)

// SQLParser parses SQL queries to extract routing information
type SQLParser struct {
	// Compiled regex patterns for performance
	selectPattern  *regexp.Regexp
	insertPattern  *regexp.Regexp
	updatePattern  *regexp.Regexp
	deletePattern  *regexp.Regexp
	wherePattern   *regexp.Regexp
	tablePattern   *regexp.Regexp
	valuePattern   *regexp.Regexp
}

// ParsedQuery contains extracted information from a SQL query
type ParsedQuery struct {
	Type       string            // SELECT, INSERT, UPDATE, DELETE
	Table      string            // Main table being queried
	ShardKey   string            // Column name of shard key (if found)
	ShardValue string            // Value of shard key (if found)
	IsMultiShard bool            // True if query spans multiple shards
	CanRoute   bool              // True if we can route this query
	WhereConditions map[string]string // Column -> Value mappings from WHERE
}

// NewSQLParser creates a new SQL parser
func NewSQLParser() *SQLParser {
	return &SQLParser{
		selectPattern: regexp.MustCompile(`(?i)^\s*SELECT\s+.+\s+FROM\s+(\w+)`),
		insertPattern: regexp.MustCompile(`(?i)^\s*INSERT\s+INTO\s+(\w+)`),
		updatePattern: regexp.MustCompile(`(?i)^\s*UPDATE\s+(\w+)`),
		deletePattern: regexp.MustCompile(`(?i)^\s*DELETE\s+FROM\s+(\w+)`),
		wherePattern:  regexp.MustCompile(`(?i)\s+WHERE\s+(.+?)(?:\s+ORDER|\s+LIMIT|\s+GROUP|\s*;?\s*$)`),
		tablePattern:  regexp.MustCompile(`(?i)FROM\s+(\w+)`),
		valuePattern:  regexp.MustCompile(`(\w+)\s*=\s*['"]?([^'"=\s,]+)['"]?`),
	}
}

// Parse analyzes a SQL query and extracts routing information
func (p *SQLParser) Parse(sql string, shardKeyColumn string) (*ParsedQuery, error) {
	sql = strings.TrimSpace(sql)
	
	result := &ParsedQuery{
		WhereConditions: make(map[string]string),
		CanRoute:        false,
	}
	
	// Determine query type and extract table
	upperSQL := strings.ToUpper(sql)
	
	switch {
	case strings.HasPrefix(upperSQL, "SELECT"):
		result.Type = "SELECT"
		if matches := p.selectPattern.FindStringSubmatch(sql); len(matches) > 1 {
			result.Table = strings.ToLower(matches[1])
		}
		
	case strings.HasPrefix(upperSQL, "INSERT"):
		result.Type = "INSERT"
		if matches := p.insertPattern.FindStringSubmatch(sql); len(matches) > 1 {
			result.Table = strings.ToLower(matches[1])
		}
		// For INSERT, extract shard key from VALUES
		result.ShardKey, result.ShardValue = p.extractInsertShardKey(sql, shardKeyColumn)
		if result.ShardValue != "" {
			result.CanRoute = true
		}
		return result, nil
		
	case strings.HasPrefix(upperSQL, "UPDATE"):
		result.Type = "UPDATE"
		if matches := p.updatePattern.FindStringSubmatch(sql); len(matches) > 1 {
			result.Table = strings.ToLower(matches[1])
		}
		
	case strings.HasPrefix(upperSQL, "DELETE"):
		result.Type = "DELETE"
		if matches := p.deletePattern.FindStringSubmatch(sql); len(matches) > 1 {
			result.Table = strings.ToLower(matches[1])
		}
		
	default:
		// DDL or other statements - broadcast to all shards
		result.Type = "OTHER"
		result.IsMultiShard = true
		return result, nil
	}
	
	// Extract WHERE conditions
	if whereMatches := p.wherePattern.FindStringSubmatch(sql); len(whereMatches) > 1 {
		whereClause := whereMatches[1]
		
		// Extract column = value pairs
		valueMatches := p.valuePattern.FindAllStringSubmatch(whereClause, -1)
		for _, match := range valueMatches {
			if len(match) > 2 {
				column := strings.ToLower(match[1])
				value := match[2]
				result.WhereConditions[column] = value
				
				// Check if this is the shard key
				if column == strings.ToLower(shardKeyColumn) {
					result.ShardKey = column
					result.ShardValue = value
					result.CanRoute = true
				}
			}
		}
	}
	
	// If no shard key found in WHERE, this might be a cross-shard query
	if result.ShardValue == "" {
		result.IsMultiShard = true
	}
	
	return result, nil
}

// extractInsertShardKey extracts shard key from INSERT statement
func (p *SQLParser) extractInsertShardKey(sql string, shardKeyColumn string) (string, string) {
	// Pattern: INSERT INTO table (col1, col2, ...) VALUES (val1, val2, ...)
	columnsPattern := regexp.MustCompile(`(?i)INSERT\s+INTO\s+\w+\s*\(([^)]+)\)\s*VALUES\s*\(([^)]+)\)`)
	matches := columnsPattern.FindStringSubmatch(sql)
	
	if len(matches) < 3 {
		return "", ""
	}
	
	columns := strings.Split(matches[1], ",")
	values := strings.Split(matches[2], ",")
	
	if len(columns) != len(values) {
		return "", ""
	}
	
	// Find the shard key column
	for i, col := range columns {
		col = strings.TrimSpace(col)
		col = strings.Trim(col, `"'`)
		col = strings.ToLower(col)
		
		if col == strings.ToLower(shardKeyColumn) {
			value := strings.TrimSpace(values[i])
			value = strings.Trim(value, `"'`)
			return col, value
		}
	}
	
	return "", ""
}

// IsReadQuery returns true if the query is a read-only query
func (p *ParsedQuery) IsReadQuery() bool {
	return p.Type == "SELECT"
}

// IsWriteQuery returns true if the query modifies data
func (p *ParsedQuery) IsWriteQuery() bool {
	return p.Type == "INSERT" || p.Type == "UPDATE" || p.Type == "DELETE"
}

// String returns a string representation of the parsed query
func (p *ParsedQuery) String() string {
	return fmt.Sprintf("ParsedQuery{Type: %s, Table: %s, ShardKey: %s=%s, CanRoute: %v, MultiShard: %v}",
		p.Type, p.Table, p.ShardKey, p.ShardValue, p.CanRoute, p.IsMultiShard)
}

// ExtractTableFromSQL extracts the table name from any SQL query
func ExtractTableFromSQL(sql string) string {
	sql = strings.TrimSpace(sql)
	upperSQL := strings.ToUpper(sql)
	
	patterns := []struct {
		prefix  string
		pattern *regexp.Regexp
	}{
		{"SELECT", regexp.MustCompile(`(?i)FROM\s+(\w+)`)},
		{"INSERT", regexp.MustCompile(`(?i)INSERT\s+INTO\s+(\w+)`)},
		{"UPDATE", regexp.MustCompile(`(?i)UPDATE\s+(\w+)`)},
		{"DELETE", regexp.MustCompile(`(?i)DELETE\s+FROM\s+(\w+)`)},
	}
	
	for _, p := range patterns {
		if strings.HasPrefix(upperSQL, p.prefix) {
			if matches := p.pattern.FindStringSubmatch(sql); len(matches) > 1 {
				return strings.ToLower(matches[1])
			}
		}
	}
	
	return ""
}

// RewriteQueryForShard rewrites a query to target a specific shard
// This is useful for scatter-gather operations where we need to query all shards
func RewriteQueryForShard(sql string, shardID string) string {
	// For now, just return the original query
	// In a more advanced implementation, this could add shard hints
	return sql
}

