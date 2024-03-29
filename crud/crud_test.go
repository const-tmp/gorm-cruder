package crud

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
	"time"
)

type testSuite struct {
	suite.Suite
	db   *gorm.DB
	crud GenericCRUD[User]
}

func TestCRUD(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) SetupSuite() {
	db, err := gorm.Open(postgres.Open(fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=%s",
		"localhost",
		"postgres",
		"password",
		"postgres",
		5432,
		"Europe/Kiev",
	)))
	s.Require().NoError(err)
	s.db = db
	s.crud = New[User](db)

	s.Run("migrate", func() {
		tables, err := s.db.Migrator().GetTables()
		s.Require().NoError(err)
		for _, table := range tables {
			s.Require().NoError(s.db.Debug().Migrator().DropTable(table))
		}
		s.Require().NoError(s.db.Debug().AutoMigrate(&User{}))
	})
}

var (
	a1 = sql.NullInt16{Int16: 11, Valid: true}
	a2 = sql.NullInt16{Int16: 12, Valid: true}
	a3 = sql.NullInt16{Int16: 111, Valid: true}
	a4 = sql.NullInt16{Int16: 1111, Valid: true}
)

func (s *testSuite) TestCRUD() {
	var user User
	s.Run("create", func() {
		v, err := s.crud.Create(context.TODO(), User{Name: "test", Age: a1})
		s.NoError(err)
		s.T().Logf("%+v", v)
		s.NotZero(v.PrimaryKey())
		user = *v
	})
	s.Run("get or create get", func() {
		v, err := s.crud.GetOrCreate(context.TODO(), User{Name: "test", Age: a1})
		s.NoError(err)
		s.T().Logf("%+v", v)
		s.NotZero(v.PrimaryKey())
		s.Equal(v.PrimaryKey(), user.PrimaryKey())
	})
	s.Run("get or create create", func() {
		v, err := s.crud.GetOrCreate(context.TODO(), User{Name: "test2", Age: a2})
		s.NoError(err)
		s.T().Logf("%+v", v)
		s.NotZero(v.PrimaryKey())
		s.NotEqual(v.PrimaryKey(), user.PrimaryKey())
	})
	s.Run("query", func() {
		testCases := []struct {
			user User
			len  int
		}{
			{User{}, 2},
			{User{Name: "test"}, 1},
			{User{Name: "test2"}, 1},
			{User{Age: a1}, 1},
			{User{Age: a2}, 1},
			{User{Name: "test2", Age: a1}, 0},
			{User{Name: "test2", Age: a2}, 1},
			{User{Name: "test", Age: a2}, 0},
			{User{Name: "test", Age: a1}, 1},
		}
		for _, tc := range testCases {
			s.Run("query", func() {
				v, err := s.crud.Query(context.TODO(), tc.user)
				s.NoError(err)
				s.Len(v, tc.len)
			})
		}
	})
	s.Run("query one", func() {
		testCases := []struct {
			user User
			err  error
		}{
			{User{}, MultipleResultsError},
			{User{Name: "test"}, nil},
			{User{Name: "test2"}, nil},
			{User{Age: a1}, nil},
			{User{Age: a2}, nil},
			{User{Name: "test2", Age: a1}, gorm.ErrRecordNotFound},
			{User{Name: "test2", Age: a2}, nil},
			{User{Name: "test", Age: a2}, gorm.ErrRecordNotFound},
			{User{Name: "test", Age: a1}, nil},
		}
		for _, tc := range testCases {
			s.Run("query", func() {
				v, err := s.crud.QueryOne(context.TODO(), tc.user)
				switch err {
				case MultipleResultsError:
				case gorm.ErrRecordNotFound:
				case nil:
				default:
					s.Error(err)
				}
				s.T().Logf("%+v", v)
			})
		}
	})
	s.Run("update field", func() {
		err := s.crud.UpdateField(context.TODO(), user, "name", "test!")
		s.Require().NoError(err)
	})
	s.Run("update field", func() {
		err := s.crud.UpdateField(context.TODO(), user, "age", 111)
		s.Require().NoError(err)
	})
	s.Run("query", func() {
		_, err := s.crud.QueryOne(context.TODO(), user)
		s.Require().ErrorIs(err, gorm.ErrRecordNotFound)
	})
	s.Run("query", func() {
		v, err := s.crud.QueryOne(context.TODO(), User{Name: "test!"})
		s.Require().NoError(err)
		s.T().Logf("%+v", v)
	})
	s.Run("query", func() {
		v, err := s.crud.QueryOne(context.TODO(), User{Age: a3})
		s.Require().NoError(err)
		s.T().Logf("%+v", v)
	})
	s.Run("update", func() {
		err := s.crud.Update(context.TODO(), User{Model: gorm.Model{ID: user.Model.ID}, Name: "test!!", Age: a4})
		s.Require().NoError(err)
	})
	s.Run("query", func() {
		v, err := s.crud.QueryOne(context.TODO(), User{Model: gorm.Model{ID: user.Model.ID}})
		s.Require().NoError(err)
		s.T().Logf("%+v", v)
		s.Require().Equal(a4, v.Age)
		s.Require().Equal("test!!", v.Name)
	})
	s.Run("query map", func() {
		v, err := s.crud.QueryMapOne(context.TODO(), map[string]any{
			"id": user.ID,
		})
		s.Require().NoError(err)
		s.T().Logf("%+v", v)
		s.Require().Equal(a4, v.Age)
		s.Require().Equal("test!!", v.Name)
	})
	s.Run("query map", func() {
		v, err := s.crud.QueryMapOne(context.TODO(), map[string]any{
			"id":  user.ID,
			"age": 1111,
		})
		s.Require().NoError(err)
		s.T().Logf("%+v", v)
		s.Require().Equal(a4, v.Age)
		s.Require().Equal("test!!", v.Name)
	})
	s.Run("smart query", func() {
		v, err := s.crud.SmartQuery(context.TODO(), Query{OrderBy: map[string]OrderBy{"id": ASC}})
		s.Require().NoError(err)
		for i, u := range v {
			s.T().Log(i, u)
		}
	})
	s.Run("smart query", func() {
		v, err := s.crud.SmartQuery(context.TODO(), Query{OrderBy: map[string]OrderBy{"id": DESC}})
		s.Require().NoError(err)
		for i, u := range v {
			s.T().Log(i, u)
		}
	})
	s.Run("smart query", func() {
		v, err := s.crud.SmartQuery(context.TODO(), Query{OrderBy: map[string]OrderBy{"created_at": ASC}})
		s.Require().NoError(err)
		for i, u := range v {
			s.T().Log(i, u)
		}
	})
	s.Run("smart query", func() {
		v, err := s.crud.SmartQuery(context.TODO(), Query{OrderBy: map[string]OrderBy{"created_at": DESC}})
		s.Require().NoError(err)
		for i, u := range v {
			s.T().Log(i, u)
		}
	})
	s.Run("smart query", func() {
		v, err := s.crud.SmartQuery(context.TODO(), Query{
			OrderBy: map[string]OrderBy{"created_at": DESC},
			Like:    map[string]string{"name": "test"},
		})
		s.Require().NoError(err)
		for i, u := range v {
			s.T().Log(i, u)
		}
	})
	s.Run("smart query", func() {
		v, err := s.crud.SmartQuery(context.TODO(), Query{
			OrderBy: map[string]OrderBy{"created_at": DESC},
			Like:    map[string]string{"name": "test"},
			Equal:   map[string]any{"name": "test2"},
		})
		s.Require().NoError(err)
		for i, u := range v {
			s.T().Log(i, u)
		}
	})
	s.Run("smart query", func() {
		v, err := s.crud.SmartQuery(context.TODO(), Query{
			OrderBy: map[string]OrderBy{"created_at": DESC},
			Like:    map[string]string{"name": "test"},
			Equal:   map[string]any{"name": "test2"},
			Between: map[string]struct{ From, To any }{"created_at": {
				From: time.Date(2023, 1, 23, 0, 0, 0, 0, time.Local),
				To:   time.Date(2023, 1, 24, 0, 0, 0, 0, time.Local),
			}},
		})
		s.Require().NoError(err)
		for i, u := range v {
			s.T().Log(i, u)
		}
	})
}

type User struct {
	gorm.Model
	Name string
	Age  sql.NullInt16
}

func (u User) PrimaryKey() any {
	return u.Model.ID
}
