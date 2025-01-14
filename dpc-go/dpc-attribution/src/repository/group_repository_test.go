package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CMSgov/dpc/attribution/attributiontest"
	"github.com/CMSgov/dpc/attribution/middleware"
	"github.com/CMSgov/dpc/attribution/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GroupRepositoryTestSuite struct {
	suite.Suite
	fakeGrp *model.Group
}

func (suite *GroupRepositoryTestSuite) SetupTest() {
	g := model.Group{}
	err := faker.FakeData(&g)
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}
	var i model.Info
	_ = json.Unmarshal([]byte(attributiontest.Groupjson), &i)
	g.Info = i
	suite.fakeGrp = &g
}

func TestGroupRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(GroupRepositoryTestSuite))
}

func (suite *GroupRepositoryTestSuite) TestInsertErrorInRepo() {
	db, mock := newMock()
	defer db.Close()
	repo := NewGroupRepo(db)
	ctx := context.WithValue(context.Background(), middleware.ContextKeyOrganization, "12345")

	expectedInsertQuery := `INSERT INTO "groups" \(info, organization_id, version\) VALUES \(\$1, \$2, \$3\) returning id, version, created_at, updated_at, info, organization_id`

	rows := sqlmock.NewRows([]string{"id", "version", "created_at", "updated_at", "info"})

	mock.ExpectQuery(expectedInsertQuery).WithArgs(suite.fakeGrp.Info, "12345", 0).WillReturnRows(rows)

	b, _ := json.Marshal(suite.fakeGrp.Info)
	group, err := repo.Insert(ctx, b)
	assert.EqualError(suite.T(), err, "sql: no rows in result set")
	assert.Empty(suite.T(), group)
}

func (suite *GroupRepositoryTestSuite) TestInsert() {
	db, mock := newMock()
	defer db.Close()
	repo := NewGroupRepo(db)
	ctx := context.WithValue(context.Background(), middleware.ContextKeyOrganization, "12345")

	expectedCountQuery := `SELECT COUNT\(id\) AS c FROM "groups" WHERE organization_id = \$1`
	rows := sqlmock.NewRows([]string{"count"}).
		AddRow(0)

	mock.ExpectQuery(expectedCountQuery).WillReturnRows(rows)

	expectedInsertQuery := `INSERT INTO "groups" \(info, organization_id, version\) VALUES \(\$1, \$2, \$3\) returning id, version, created_at, updated_at, info, organization_id`

	rows = sqlmock.NewRows([]string{"id", "version", "created_at", "updated_at", "info", "organization_id"}).
		AddRow(suite.fakeGrp.ID, suite.fakeGrp.Version, suite.fakeGrp.CreatedAt, suite.fakeGrp.UpdatedAt, suite.fakeGrp.Info, suite.fakeGrp.OrganizationID)

	mock.ExpectQuery(expectedInsertQuery).WithArgs(suite.fakeGrp.Info, "12345", 0).WillReturnRows(rows)

	b, _ := json.Marshal(suite.fakeGrp.Info)
	group, err := repo.Insert(ctx, b)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.fakeGrp.ID, group.ID)
}

func (suite *GroupRepositoryTestSuite) TestFindByID() {
	db, mock := newMock()
	defer db.Close()
	repo := NewGroupRepo(db)
	ctx := context.WithValue(context.Background(), middleware.ContextKeyOrganization, "12345")

	expectedSelectQuery := `SELECT id, version, created_at, updated_at, info, organization_id FROM "groups" WHERE organization_id = \$1 AND id = \$2`

	rows := sqlmock.NewRows([]string{"id", "version", "created_at", "updated_at", "info", "organization_id"}).
		AddRow(suite.fakeGrp.ID, suite.fakeGrp.Version, suite.fakeGrp.CreatedAt, suite.fakeGrp.UpdatedAt, suite.fakeGrp.Info, suite.fakeGrp.OrganizationID)

	mock.ExpectQuery(expectedSelectQuery).WithArgs("12345", suite.fakeGrp.ID).WillReturnRows(rows)

	group, err := repo.FindByID(ctx, suite.fakeGrp.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.fakeGrp.ID, group.ID)
}
