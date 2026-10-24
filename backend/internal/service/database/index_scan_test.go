package database

import (
	"database/sql"
	"testing"

	"cloudpivot/internal/dto"
)

// TestBuildIndexScanTargets 验证索引 Scan 目标数能适配不同 MySQL 版本的列数。
//
// 2026-10-24 事故：MySQL 8.0 的 SHOW INDEX 返回 15 列（比 5.7 多 Visible、
// Expression 两列），代码写死 13 个 Scan 目标，报
// 「expected 15 destination arguments in Scan, not 13」，
// 导致「表结构」页面 100% 报错。
func TestBuildIndexScanTargets(t *testing.T) {
	for _, colCount := range []int{13, 15} {
		var idx dto.TableIndex
		var subPart, packed sql.NullString
		targets := buildIndexScanTargets(&idx, &subPart, &packed, colCount)
		if len(targets) != colCount {
			t.Errorf("列数 %d 时生成了 %d 个 Scan 目标，数量不匹配", colCount, len(targets))
		}
	}
}

// TestBuildIndexScanTargetsKeepsKnownFields 验证前 13 列仍绑定到结构体字段。
//
// 补齐新列时不能打乱已有列的绑定顺序，否则字段会串位。
func TestBuildIndexScanTargetsKeepsKnownFields(t *testing.T) {
	var idx dto.TableIndex
	var subPart, packed sql.NullString
	targets := buildIndexScanTargets(&idx, &subPart, &packed, 15)

	// 前 13 个目标应分别指向已知字段
	if targets[0] != &idx.Table {
		t.Error("第 1 个目标应绑定 Table")
	}
	if targets[3] != &idx.SeqInIndex {
		t.Error("第 4 个目标应绑定 SeqInIndex")
	}
	if targets[4] != &idx.ColumnName {
		t.Error("第 5 个目标应绑定 ColumnName")
	}
	if targets[10] != &idx.IndexType {
		t.Error("第 11 个目标应绑定 IndexType")
	}
	if targets[12] != &idx.IndexComment {
		t.Error("第 13 个目标应绑定 IndexComment")
	}
}
