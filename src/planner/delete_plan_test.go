/*
 * Radon
 *
 * Copyright 2018 The Radon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package planner

import (
	"testing"

	"router"

	"github.com/stretchr/testify/assert"
	"github.com/xelabs/go-mysqlstack/sqlparser"
	"github.com/xelabs/go-mysqlstack/xlog"
)

func TestDeletePlan(t *testing.T) {
	results := []string{
		`{
	"RawQuery": "delete LOW_PRIORITY LOW_PRIORITY from sbtest.A where id=1",
	"Partitions": [
		{
			"Query": "delete low_priority low_priority from sbtest.A6 where id = 1",
			"Backend": "backend6",
			"Range": "[512-4096)"
		}
	]
}`,
		`{
	"RawQuery": "delete QUICK QUICK from sbtest.A where id=1 order by xx",
	"Partitions": [
		{
			"Query": "delete quick quick from sbtest.A6 where id = 1 order by xx asc",
			"Backend": "backend6",
			"Range": "[512-4096)"
		}
	]
}`,
		`{
	"RawQuery": "delete IGNORE IGNORE from sbtest.A where name='xx'",
	"Partitions": [
		{
			"Query": "delete ignore ignore from sbtest.A1 where name = 'xx'",
			"Backend": "backend1",
			"Range": "[0-32)"
		},
		{
			"Query": "delete ignore ignore from sbtest.A2 where name = 'xx'",
			"Backend": "backend2",
			"Range": "[32-64)"
		},
		{
			"Query": "delete ignore ignore from sbtest.A3 where name = 'xx'",
			"Backend": "backend3",
			"Range": "[64-96)"
		},
		{
			"Query": "delete ignore ignore from sbtest.A4 where name = 'xx'",
			"Backend": "backend4",
			"Range": "[96-256)"
		},
		{
			"Query": "delete ignore ignore from sbtest.A5 where name = 'xx'",
			"Backend": "backend5",
			"Range": "[256-512)"
		},
		{
			"Query": "delete ignore ignore from sbtest.A6 where name = 'xx'",
			"Backend": "backend6",
			"Range": "[512-4096)"
		}
	]
}`,
		`{
	"RawQuery": "delete LOW_PRIORITY QUICK IGNORE from sbtest.A where id in (1, 2,3)",
	"Partitions": [
		{
			"Query": "delete low_priority quick ignore from sbtest.A6 where id in (1, 2, 3)",
			"Backend": "backend6",
			"Range": "[512-4096)"
		}
	]
}`,
		`{
	"RawQuery": "delete from sbtest.G where id in (1, 2,3)",
	"Partitions": [
		{
			"Query": "delete from sbtest.G where id in (1, 2, 3)",
			"Backend": "backend1",
			"Range": ""
		},
		{
			"Query": "delete from sbtest.G where id in (1, 2, 3)",
			"Backend": "backend2",
			"Range": ""
		}
	]
}`,
		`{
	"RawQuery": "delete from sbtest.S where id in (1, 2,3)",
	"Partitions": [
		{
			"Query": "delete from sbtest.S where id in (1, 2, 3)",
			"Backend": "backend1",
			"Range": ""
		}
	]
}`,
		`{
	"RawQuery": "delete from sbtest.A order by xx limit 1",
	"Partitions": [
		{
			"Query": "delete from sbtest.A1 order by xx asc limit 1",
			"Backend": "backend1",
			"Range": "[0-32)"
		},
		{
			"Query": "delete from sbtest.A2 order by xx asc limit 1",
			"Backend": "backend2",
			"Range": "[32-64)"
		},
		{
			"Query": "delete from sbtest.A3 order by xx asc limit 1",
			"Backend": "backend3",
			"Range": "[64-96)"
		},
		{
			"Query": "delete from sbtest.A4 order by xx asc limit 1",
			"Backend": "backend4",
			"Range": "[96-256)"
		},
		{
			"Query": "delete from sbtest.A5 order by xx asc limit 1",
			"Backend": "backend5",
			"Range": "[256-512)"
		},
		{
			"Query": "delete from sbtest.A6 order by xx asc limit 1",
			"Backend": "backend6",
			"Range": "[512-4096)"
		}
	]
}`,
		`{
	"RawQuery": "delete from sbtest.G order by xx limit 2",
	"Partitions": [
		{
			"Query": "delete from sbtest.G order by xx asc limit 2",
			"Backend": "backend1",
			"Range": ""
		},
		{
			"Query": "delete from sbtest.G order by xx asc limit 2",
			"Backend": "backend2",
			"Range": ""
		}
	]
}`,
		`{
	"RawQuery": "delete from sbtest.S",
	"Partitions": [
		{
			"Query": "delete from sbtest.S",
			"Backend": "backend1",
			"Range": ""
		}
	]
}`,
	}
	querys := []string{
		"delete LOW_PRIORITY LOW_PRIORITY from sbtest.A where id=1",
		"delete QUICK QUICK from sbtest.A where id=1 order by xx",
		"delete IGNORE IGNORE from sbtest.A where name='xx'",
		"delete LOW_PRIORITY QUICK IGNORE from sbtest.A where id in (1, 2,3)",
		"delete from sbtest.G where id in (1, 2,3)",
		"delete from sbtest.S where id in (1, 2,3)",
		"delete from sbtest.A order by xx limit 1",
		"delete from sbtest.G order by xx limit 2",
		"delete from sbtest.S",
	}

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig(), router.MockTableGConfig(), router.MockTableSConfig())
	assert.Nil(t, err)
	planTree := NewPlanTree()
	for i, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		plan := NewDeletePlan(log, database, query, node.(*sqlparser.Delete), route)

		// plan build
		{
			err := plan.Build()
			assert.Nil(t, err)
			{
				err := planTree.Add(plan)
				assert.Nil(t, err)
			}
			got := plan.JSON()
			log.Debug(got)
			want := results[i]
			assert.Equal(t, want, got)
			assert.Equal(t, PlanTypeDelete, plan.Type())
		}
	}
}

func TestDeleteUnsupportedPlan(t *testing.T) {
	querys := []string{
		"delete from sbtest.A where id in (select id from t1)",
		"DELETE a1,a2 FROM db1.t1, db2.t2",
		"delete a1, a2 from db3.t1 as a1, db4.t2 as a2",
		"delete a from a join b on a.id = b.id where b.name = 'test'",
		"DELETE FROM t1, alias USING t1, t2 alias WHERE t1.a = alias.a",
		"delete from t partition (p0) where a = 1",
	}

	results := []string{
		"unsupported: subqueries.in.delete",
		"unsupported: currently.not.support.multitables.in.delete",
		"unsupported: currently.not.support.multitables.in.delete",
		"unsupported: currently.not.support.multitables.in.delete",
		"unsupported: currently.not.support.multitables.in.delete",
		"unsupported: currently.not.support.partitions.in.delete",
	}

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig())
	assert.Nil(t, err)
	for i, query := range querys {
		node, err := sqlparser.Parse(query)
		assert.Nil(t, err)
		plan := NewDeletePlan(log, database, query, node.(*sqlparser.Delete), route)

		// plan build
		{
			err := plan.Build()
			want := results[i]
			got := err.Error()
			assert.Equal(t, want, got)
		}
	}
}

func TestDeleteErrorPlan(t *testing.T) {
	query := "delete from A where id=1"

	log := xlog.NewStdLog(xlog.Level(xlog.PANIC))
	database := "sbtest"

	route, cleanup := router.MockNewRouter(log)
	defer cleanup()

	err := route.CreateDatabase(database)
	assert.Nil(t, err)
	err = route.AddForTest(database, router.MockTableMConfig())
	assert.Nil(t, err)
	databaseNull := ""
	node, err := sqlparser.Parse(query)
	assert.Nil(t, err)
	plan := NewDeletePlan(log, databaseNull, query, node.(*sqlparser.Delete), route)

	// plan build
	{
		planTree := NewPlanTree()
		{
			err := planTree.Add(plan)
			assert.Nil(t, err)
		}

		{
			err := planTree.Build()
			assert.NotNil(t, err)
		}
	}
}
